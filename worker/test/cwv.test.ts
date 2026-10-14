import { expect, test } from "bun:test";
import { cwvFromLighthouse } from "../src/cwv";
import worker, { type Env } from "../src/index";

const audit = (numericValue: number) => ({ numericValue });
const psi = (perf: number, lcp: number, cls: number, tbt: number, fcp: number) => ({
  lighthouseResult: {
    categories: { performance: { score: perf } },
    audits: {
      "largest-contentful-paint": audit(lcp),
      "cumulative-layout-shift": audit(cls),
      "total-blocking-time": audit(tbt),
      "first-contentful-paint": audit(fcp),
    },
  },
});

test("a fast page passes every vital and scores Lighthouse performance", () => {
  const cwv = cwvFromLighthouse(psi(0.95, 1800, 0.02, 90, 1200));
  expect(cwv.score).toBe(95);
  expect(cwv.passed).toBe(4);
  expect(cwv.total).toBe(4);
  expect(cwv.checks.map((c) => c.id)).toEqual(["cwv_lcp", "cwv_cls", "cwv_tbt", "cwv_fcp"]);
});

test("each vital fails just past Google's good threshold", () => {
  const cwv = cwvFromLighthouse(psi(0.4, 2600, 0.11, 210, 1900));
  expect(cwv.passed).toBe(0);
  expect(cwv.checks.find((c) => c.id === "cwv_lcp")?.label).toContain("2.6 s");
});

test("a response without lighthouse data is an error, not a zero score", () => {
  expect(() => cwvFromLighthouse({ error: { message: "quota" } })).toThrow("quota");
});

const cwvPost = (url: string) =>
  new Request("https://w.test/api/cwv", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ url }),
  });

test("/api/cwv without a PageSpeed key says it is unavailable", async () => {
  const resp = await worker.fetch(cwvPost("deno.com"), {
    SCAN_LIMITER: { limit: async () => ({ success: true }) },
  } as unknown as Env);
  expect(resp.status).toBe(503);
});

test("/api/cwv returns vitals for the page from PageSpeed Insights", async () => {
  const realFetch = globalThis.fetch;
  let asked = "";
  globalThis.fetch = (async (u: URL) => {
    asked = String(u);
    return Response.json(psi(0.8, 2000, 0.05, 100, 1000));
  }) as unknown as typeof fetch;
  try {
    const env = { SCAN_LIMITER: { limit: async () => ({ success: true }) }, PSI_KEY: "k" } as unknown as Env;
    const resp = await worker.fetch(cwvPost("deno.com"), env);
    const body = (await resp.json()) as { score: number; total: number };
    expect(resp.status).toBe(200);
    expect(body.score).toBe(80);
    expect(body.total).toBe(4);
    expect(asked).toContain("url=https%3A%2F%2Fdeno.com");
  } finally {
    globalThis.fetch = realFetch;
  }
});
