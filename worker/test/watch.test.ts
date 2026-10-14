import { expect, test } from "bun:test";
import worker, { type Env } from "../src/index";

function fakeDb() {
  const rows: unknown[][] = [];
  const db = {
    prepare: (_sql: string) => ({
      bind: (...args: unknown[]) => ({ run: async () => void rows.push(args) }),
    }),
  };
  return { rows, db };
}

function envWith(db: unknown): Env {
  return { SCAN_LIMITER: { limit: async () => ({ success: true }) }, WATCH_DB: db } as unknown as Env;
}

const post = (body: unknown) =>
  new Request("https://w.test/api/watch", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

test("a consenting signup is stored with email and host", async () => {
  const { rows, db } = fakeDb();
  const resp = await worker.fetch(
    post({ email: " Ann@Example.com ", url: "https://deno.com/x", consent: true }),
    envWith(db),
  );
  expect(resp.status).toBe(200);
  expect(rows.length).toBe(1);
  expect(rows[0]?.slice(0, 2)).toEqual(["ann@example.com", "deno.com"]);
});

test("bad email, missing consent or a bad url are rejected without storing", async () => {
  const { rows, db } = fakeDb();
  for (const body of [
    { email: "nope", url: "deno.com", consent: true },
    { email: "a@b.co", url: "deno.com" },
    { email: "a@b.co", url: "not a host", consent: true },
  ]) {
    expect((await worker.fetch(post(body), envWith(db))).status).toBe(400);
  }
  expect(rows.length).toBe(0);
});

test("a filled honeypot looks accepted but stores nothing", async () => {
  const { rows, db } = fakeDb();
  const resp = await worker.fetch(post({ email: "a@b.co", url: "deno.com", consent: true, hp: "spam" }), envWith(db));
  expect(resp.status).toBe(200);
  expect(rows.length).toBe(0);
});

test("signups are rate limited", async () => {
  const { db } = fakeDb();
  const env = { SCAN_LIMITER: { limit: async () => ({ success: false }) }, WATCH_DB: db } as unknown as Env;
  expect((await worker.fetch(post({ email: "a@b.co", url: "deno.com", consent: true }), env)).status).toBe(429);
});

test("a database failure answers 503 json instead of crashing", async () => {
  const db = {
    prepare: () => ({
      bind: () => ({
        run: async () => {
          throw new Error("no such table");
        },
      }),
    }),
  };
  const resp = await worker.fetch(post({ email: "a@b.co", url: "deno.com", consent: true }), envWith(db));
  expect(resp.status).toBe(503);
  expect(resp.headers.get("Content-Type")).toBe("application/json");
});

test("signups from other sites' pages or without a json body type are refused", async () => {
  const { rows, db } = fakeDb();
  const body = JSON.stringify({ email: "a@b.co", url: "deno.com", consent: true });
  const foreign = new Request("https://w.test/api/watch", {
    method: "POST",
    headers: { "Content-Type": "application/json", Origin: "https://evil.example" },
    body,
  });
  const plain = new Request("https://w.test/api/watch", {
    method: "POST",
    headers: { "Content-Type": "text/plain" },
    body,
  });
  expect((await worker.fetch(foreign, envWith(db))).status).toBe(403);
  expect((await worker.fetch(plain, envWith(db))).status).toBe(415);
  expect(rows.length).toBe(0);
});

test("the worker's page turns the watch form on", async () => {
  const html = await (await worker.fetch(new Request("https://w.test/"), envWith(null))).text();
  expect(html).toContain('<meta name="findable-watch" content="on">');
});
