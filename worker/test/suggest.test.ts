import { expect, test } from "bun:test";
import worker, { type Env } from "../src/index";

const post = (body: object) =>
  new Request("https://w.test/api/suggest", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

const PAGE = `<html><head><title>Old title</title><meta name="description" content="Old desc"></head>
<body><h1>Fresh bread</h1><p>We bake sourdough daily.</p></body></html>`;

const model = (reply: string, seen: { prompt?: string } = {}) =>
  ({
    run: async (_m: string, input: { messages: { content: string }[] }) => {
      seen.prompt = input.messages.map((m) => m.content).join("\n");
      return { response: reply };
    },
  }) as unknown as Ai;

const limiter = { limit: async () => ({ success: true }) };

async function withPage<T>(fn: () => Promise<T>, init: ResponseInit = {}, body = PAGE): Promise<T> {
  const realFetch = globalThis.fetch;
  globalThis.fetch = (async () =>
    new Response(body, { headers: { "Content-Type": "text/html" }, ...init })) as unknown as typeof fetch;
  try {
    return await fn();
  } finally {
    globalThis.fetch = realFetch;
  }
}

const goodReply = JSON.stringify({
  title: "Sourdough Bakery",
  description: "Fresh sourdough baked daily.",
  faq: [
    { q: "Q1?", a: "A1" },
    { q: "Q2?", a: "A2" },
    { q: "Q3?", a: "A3" },
  ],
});

test("/api/suggest returns title, description and 3 faq from the model, fed the page content", async () => {
  const seen: { prompt?: string } = {};
  const env = { SCAN_LIMITER: limiter, AI: model(goodReply, seen) } as unknown as Env;
  const resp = await withPage(() => worker.fetch(post({ url: "bakery.example", keyphrase: "sourdough" }), env));
  const body = (await resp.json()) as { title: string; description: string; faq: { q: string; a: string }[] };
  expect(resp.status).toBe(200);
  expect(body.title).toBe("Sourdough Bakery");
  expect(body.faq).toHaveLength(3);
  expect(seen.prompt).toContain("Fresh bread");
  expect(seen.prompt).toContain("sourdough");
});

test("/api/suggest clips an over-long title and description and extra faq", async () => {
  const long = JSON.stringify({
    title: "T".repeat(100),
    description: "D".repeat(300),
    faq: [1, 2, 3, 4, 5].map((n) => ({ q: `Q${n}`, a: `A${n}` })),
  });
  const env = { SCAN_LIMITER: limiter, AI: model(`Here you go: ${long}`) } as unknown as Env;
  const resp = await withPage(() => worker.fetch(post({ url: "bakery2.example" }), env));
  const body = (await resp.json()) as { title: string; description: string; faq: unknown[] };
  expect(body.title.length).toBeLessThanOrEqual(60);
  expect(body.description.length).toBeLessThanOrEqual(155);
  expect(body.faq).toHaveLength(3);
});

test("/api/suggest answers 502 when the model does not return JSON", async () => {
  const env = { SCAN_LIMITER: limiter, AI: model("sorry, no") } as unknown as Env;
  const resp = await withPage(() => worker.fetch(post({ url: "bakery3.example" }), env));
  expect(resp.status).toBe(502);
  expect(((await resp.json()) as { error: string }).error).toBeTruthy();
});

test("/api/suggest answers 503 without an AI binding", async () => {
  const resp = await worker.fetch(post({ url: "bakery.example" }), { SCAN_LIMITER: limiter } as unknown as Env);
  expect(resp.status).toBe(503);
});

test("/api/suggest rejects a keyphrase over 100 chars with 400", async () => {
  const env = { SCAN_LIMITER: limiter, AI: model(goodReply) } as unknown as Env;
  const resp = await worker.fetch(post({ url: "bakery.example", keyphrase: "k".repeat(101) }), env);
  expect(resp.status).toBe(400);
});

test("/api/suggest trims and clips faq questions and answers", async () => {
  const reply = JSON.stringify({
    title: "T",
    description: "D",
    faq: [{ q: `  ${"q".repeat(200)}  `, a: `  ${"a".repeat(500)}  ` }],
  });
  const env = { SCAN_LIMITER: limiter, AI: model(reply) } as unknown as Env;
  const resp = await withPage(() => worker.fetch(post({ url: "bakery4.example" }), env));
  const body = (await resp.json()) as { faq: { q: string; a: string }[] };
  expect(body.faq[0]?.q).toBe("q".repeat(120));
  expect(body.faq[0]?.a).toBe("a".repeat(300));
});

test("/api/suggest answers 403 to a foreign Origin and sends no wildcard CORS", async () => {
  const env = { SCAN_LIMITER: limiter, AI: model(goodReply) } as unknown as Env;
  const req = new Request("https://w.test/api/suggest", {
    method: "POST",
    headers: { "Content-Type": "application/json", Origin: "https://evil.example" },
    body: JSON.stringify({ url: "bakery.example" }),
  });
  const resp = await worker.fetch(req, env);
  expect(resp.status).toBe(403);
  expect(resp.headers.get("Access-Control-Allow-Origin")).toBeNull();
});

test("/api/suggest answers 422 without calling the AI when the page is an error or not html", async () => {
  let calls = 0;
  const ai = {
    run: async () => {
      calls++;
      return { response: goodReply };
    },
  } as unknown as Ai;
  const env = { SCAN_LIMITER: limiter, AI: ai } as unknown as Env;
  const notFound = await withPage(() => worker.fetch(post({ url: "bakery5.example" }), env), { status: 404 });
  const pdf = await withPage(() => worker.fetch(post({ url: "bakery6.example" }), env), {
    headers: { "Content-Type": "application/pdf" },
  });
  expect(notFound.status).toBe(422);
  expect(pdf.status).toBe(422);
  expect(calls).toBe(0);
});

test("/api/suggest caps the model reply with max_tokens", async () => {
  let seen: { max_tokens?: number } = {};
  const ai = {
    run: async (_m: string, input: { max_tokens?: number }) => {
      seen = input;
      return { response: goodReply };
    },
  } as unknown as Ai;
  const env = { SCAN_LIMITER: limiter, AI: ai } as unknown as Env;
  await withPage(() => worker.fetch(post({ url: "bakery7.example" }), env));
  expect(seen.max_tokens).toBe(600);
});

test("/api/suggest caps the title fed to the model at 300 chars", async () => {
  const seen: { prompt?: string } = {};
  const html = `<head><title>${"t".repeat(1000)}</title></head><h1>${"h".repeat(1000)}</h1><p>x</p>`;
  const env = { SCAN_LIMITER: limiter, AI: model(goodReply, seen) } as unknown as Env;
  await withPage(() => worker.fetch(post({ url: "bakery8.example" }), env), {}, html);
  expect(seen.prompt).toContain("t".repeat(300));
  expect(seen.prompt).not.toContain("t".repeat(301));
});
