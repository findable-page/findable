import { afterEach, expect, test } from "bun:test";
import worker from "../src/index";

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

test("/api/scan reports the composite score", async () => {
  globalThis.fetch = (async () =>
    new Response("<html><head><title>A page title of a sensible length</title></head><body><h1>Hi</h1></body></html>")) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "https://example.com/" }) }),
  );
  const scan = (await resp.json()) as { overall: number; seo: { score: number }; aeo: { score: number }; geo: { score: number } };
  expect(scan.overall).toBe(Math.round(scan.seo.score * 0.5 + scan.aeo.score * 0.25 + scan.geo.score * 0.25));
});

test("/api/scan rejects private targets with 400", async () => {
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "http://127.0.0.1/" }) }),
  );
  expect(resp.status).toBe(400);
});
