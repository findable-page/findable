import { afterEach, expect, test } from "bun:test";
import worker, { type Env } from "../src/index";
import { type StoredReport, saveReport } from "../src/report";

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

interface Row {
  id: string;
  expires_at: string;
  json: string;
  host: string;
}

function fakeDb() {
  const rows: Row[] = [];
  const db = {
    prepare: (sql: string) => ({
      bind: (...a: unknown[]) => ({
        run: async () => {
          if (sql.startsWith("INSERT")) {
            rows.push({ id: a[0] as string, host: a[1] as string, expires_at: a[4] as string, json: a[5] as string });
          }
        },
        first: async () => rows.find((r) => r.id === a[0]) ?? null,
      }),
    }),
  };
  return { rows, db };
}

const env = (db: unknown, success = true): Env =>
  ({ SCAN_LIMITER: { limit: async () => ({ success }) }, WATCH_DB: db }) as unknown as Env;

const create = (url: string) =>
  new Request("https://w.test/api/report", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ url }),
  });

function serve(html: string) {
  globalThis.fetch = (async () => new Response(html)) as unknown as typeof fetch;
}

test("creating a report stores a 90-day row and the link serves it with noindex", async () => {
  serve("<html><head><title>Short</title></head><body><h1>Hi</h1></body></html>");
  const { rows, db } = fakeDb();
  const resp = await worker.fetch(create("https://example.com/"), env(db));
  expect(resp.status).toBe(200);
  const { id, url } = (await resp.json()) as { id: string; url: string };
  expect(id).toMatch(/^[0-9A-Za-z]{12}$/);
  expect(url).toBe(`https://findable.page/r/${id}`);
  expect(rows[0]?.host).toBe("example.com");
  const days = (Date.parse(rows[0]?.expires_at ?? "") - Date.now()) / 86_400_000;
  expect(Math.round(days)).toBe(90);

  const page = await worker.fetch(new Request(`https://w.test/r/${id}`), env(db));
  const html = await page.text();
  expect(page.status).toBe(200);
  expect(html).toContain('<meta name="robots" content="noindex">');
  expect(html).toContain(`<link rel="canonical" href="https://findable.page/r/${id}">`);
  expect(html).toContain("/og/card.png?h=example.com");
  expect(html).toContain("Add a meta description with a short summary of the page.");
});

test("an unknown report id is a 404", async () => {
  const { db } = fakeDb();
  const resp = await worker.fetch(new Request("https://w.test/r/AbCdEfGhIjKl"), env(db));
  expect(resp.status).toBe(404);
});

test("an expired report is a 404", async () => {
  const { rows, db } = fakeDb();
  const g = { score: 50, checks: [] };
  const id = await saveReport(
    db as unknown as D1Database,
    { url: "https://example.com/", scannedAt: "x", overall: 50, seo: g, aeo: g, geo: g },
    new Date("2020-01-01"),
  );
  expect(rows.length).toBe(1);
  expect((await worker.fetch(new Request(`https://w.test/r/${id}`), env(db))).status).toBe(404);
});

test("hostile text in a stored report is escaped", async () => {
  const { rows, db } = fakeDb();
  const checks = [{ id: "<img src=x onerror=alert(1)>", label: "<script>alert(1)</script>", pass: false }];
  const group = { score: 10, checks };
  const id = await saveReport(db as unknown as D1Database, {
    url: 'https://example.com/"><script>x</script>',
    scannedAt: "<b>",
    overall: 10,
    seo: group,
    aeo: group,
    geo: group,
  });
  expect(rows.length).toBe(1);
  const html = await (await worker.fetch(new Request(`https://w.test/r/${id}`), env(db))).text();
  expect(html).not.toContain("<script>alert");
  expect(html).not.toContain("<script>x");
  expect(html).not.toContain("<img src=x");
  expect(html).toContain("&lt;script&gt;alert(1)&lt;/script&gt;");
});

test("report creation is refused when the rate limiter says no", async () => {
  const { rows, db } = fakeDb();
  const resp = await worker.fetch(create("https://example.com/"), env(db, false));
  expect(resp.status).toBe(429);
  expect(rows.length).toBe(0);
});

test("report creation rejects private targets", async () => {
  const { rows, db } = fakeDb();
  const resp = await worker.fetch(create("http://127.0.0.1/"), env(db));
  expect(resp.status).toBe(400);
  expect(rows.length).toBe(0);
});

const group = (label: string, n = 1) => ({
  score: 10,
  checks: Array.from({ length: n }, (_, i) => ({ id: `c${i}`, label, pass: false })),
});

test("an oversized check label is clamped before storing", async () => {
  const { rows, db } = fakeDb();
  const g = group("x".repeat(5000));
  await saveReport(db as unknown as D1Database, {
    url: "https://example.com/",
    scannedAt: "t",
    overall: 1,
    seo: g,
    aeo: g,
    geo: g,
  });
  const stored = JSON.parse(rows[0]?.json ?? "{}") as StoredReport;
  expect(stored.seo.checks[0]?.label.length).toBe(200);
});

test("saveReport refuses a report over 64000 chars and stores nothing", async () => {
  serve("<html><head><title>Short</title></head><body></body></html>");
  const { rows, db } = fakeDb();
  const g = group("x".repeat(200), 400);
  await expect(
    saveReport(db as unknown as D1Database, {
      url: "https://example.com/",
      scannedAt: "t",
      overall: 1,
      seo: g,
      aeo: g,
      geo: g,
    }),
  ).rejects.toThrow("too large");
  expect(rows.length).toBe(0);
});

test("a url over 2048 chars is a 400", async () => {
  const { rows, db } = fakeDb();
  const resp = await worker.fetch(create(`https://example.com/${"a".repeat(2100)}`), env(db));
  expect(resp.status).toBe(400);
  expect(rows.length).toBe(0);
});

test("the scheduled handler deletes expired rows", async () => {
  const sqls: string[] = [];
  const db = { prepare: (sql: string) => ({ bind: () => ({ run: async () => void sqls.push(sql) }) }) };
  await worker.scheduled(null, env(db));
  expect(sqls.some((s) => s.startsWith("DELETE FROM reports WHERE expires_at"))).toBe(true);
});

test("creating a report also purges expired rows", async () => {
  serve("<html><head><title>Short</title></head><body></body></html>");
  const sqls: string[] = [];
  const db = { prepare: (sql: string) => ({ bind: () => ({ run: async () => void sqls.push(sql) }) }) };
  await worker.fetch(create("https://example.com/"), env(db));
  expect(sqls.some((s) => s.startsWith("DELETE FROM reports"))).toBe(true);
});

test("the report page sends X-Robots-Tag noindex", async () => {
  const { db } = fakeDb();
  const g = group("Some check");
  const id = await saveReport(db as unknown as D1Database, {
    url: "https://example.com/",
    scannedAt: "t",
    overall: 1,
    seo: g,
    aeo: g,
    geo: g,
  });
  const resp = await worker.fetch(new Request(`https://w.test/r/${id}`), env(db));
  expect(resp.headers.get("X-Robots-Tag")).toBe("noindex");
});

test("a cross-origin create is a 403", async () => {
  const { db } = fakeDb();
  const req = new Request("https://w.test/api/report", {
    method: "POST",
    headers: { "Content-Type": "application/json", Origin: "https://evil.example" },
    body: JSON.stringify({ url: "https://example.com/" }),
  });
  expect((await worker.fetch(req, env(db))).status).toBe(403);
});

test("a non-json create is a 415", async () => {
  const { db } = fakeDb();
  const req = new Request("https://w.test/api/report", { method: "POST", body: "url=x" });
  expect((await worker.fetch(req, env(db))).status).toBe(415);
});

test("a malformed id is a 404", async () => {
  const { db } = fakeDb();
  for (const id of ["short", "has-dash-12345", "a".repeat(40)]) {
    expect((await worker.fetch(new Request(`https://w.test/r/${id}`), env(db))).status).toBe(404);
  }
});

test("a stored row with a bad url renders as 404", async () => {
  const { rows, db } = fakeDb();
  const g = group("c");
  const id = await saveReport(db as unknown as D1Database, {
    url: "https://example.com/",
    scannedAt: "t",
    overall: 1,
    seo: g,
    aeo: g,
    geo: g,
  });
  const row = rows[0] as Row;
  row.json = JSON.stringify({ ...JSON.parse(row.json), url: "not a url" });
  expect((await worker.fetch(new Request(`https://w.test/r/${id}`), env(db))).status).toBe(404);
});
