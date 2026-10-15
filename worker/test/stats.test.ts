import { afterEach, expect, test } from "bun:test";
import worker, { type Env } from "../src/index";

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

function fakeDb(failWrites = false) {
  const counters = new Map<string, number>();
  const db = {
    prepare: (sql: string) => ({
      bind: (...a: unknown[]) => ({
        run: async () => {
          if (failWrites) throw new Error("d1 down");
          if (sql.includes("INSERT INTO stats")) counters.set(a[0] as string, (counters.get(a[0] as string) ?? 0) + 1);
        },
        first: async () => (counters.has(a[0] as string) ? { value: counters.get(a[0] as string) } : null),
      }),
    }),
  };
  return { counters, db };
}

const env = (db?: unknown): Env =>
  ({ SCAN_LIMITER: { limit: async () => ({ success: true }) }, WATCH_DB: db }) as unknown as Env;

const scan = (url = "https://example.com/") =>
  new Request("https://w.test/api/scan", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ url }),
  });

const serve = () => {
  globalThis.fetch = (async () =>
    new Response("<html><head><title>Hi</title></head><body><h1>Hi</h1></body></html>")) as unknown as typeof fetch;
};

test("a successful scan increments the counter and /api/stats reports it with a 60s cache", async () => {
  serve();
  const { db } = fakeDb();
  expect((await worker.fetch(scan(), env(db))).status).toBe(200);
  expect((await worker.fetch(scan(), env(db))).status).toBe(200);
  const resp = await worker.fetch(new Request("https://w.test/api/stats"), env(db));
  expect(resp.headers.get("Cache-Control")).toBe("public, max-age=60");
  expect(await resp.json()).toEqual({ scans: 2 });
});

test("a failed scan is not counted", async () => {
  serve();
  const { counters, db } = fakeDb();
  const bad = new Request("https://w.test/api/scan", { method: "POST", body: "{}" });
  expect((await worker.fetch(bad, env(db))).status).toBe(400);
  expect(counters.size).toBe(0);
});

test("a counter write failure never fails the scan", async () => {
  serve();
  const { db } = fakeDb(true);
  expect((await worker.fetch(scan(), env(db))).status).toBe(200);
});

test("stats are zero before any scan and unavailable without a database", async () => {
  const { db } = fakeDb();
  const empty = await worker.fetch(new Request("https://w.test/api/stats"), env(db));
  expect(await empty.json()).toEqual({ scans: 0 });
  const none = await worker.fetch(new Request("https://w.test/api/stats"), env());
  expect(none.status).toBe(503);
});

test("the homepage has a hidden scan counter with its i18n key in every language", async () => {
  const html = await (await worker.fetch(new Request("https://w.test/", { headers: { Accept: "text/html" } }))).text();
  expect(html).toContain('id="scanCount"');
  expect(html).toContain('meta name="findable-reports"');
  for (const lang of ["en", "de", "es", "tr", "ru"]) {
    const start = html.indexOf(`\n      ${lang}: {`, html.indexOf("const I18N"));
    const block = html.slice(start, html.indexOf("\n      },", start));
    expect(block).toContain("scanCount:");
  }
});
