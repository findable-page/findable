import { afterEach, expect, test } from "bun:test";
import worker, { type Check } from "../src/index";

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

test("/api/scan reports the composite score", async () => {
  globalThis.fetch = (async () =>
    new Response(
      "<html><head><title>A page title of a sensible length</title></head><body><h1>Hi</h1></body></html>",
    )) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "https://example.com/" }) }),
  );
  const scan = (await resp.json()) as {
    overall: number;
    seo: { score: number };
    aeo: { score: number };
    geo: { score: number };
  };
  expect(scan.overall).toBe(Math.round(scan.seo.score * 0.5 + scan.aeo.score * 0.25 + scan.geo.score * 0.25));
});

test("/api/scan rejects private targets with 400", async () => {
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "http://127.0.0.1/" }) }),
  );
  expect(resp.status).toBe(400);
});

test("/api/scan answers 429 with a JSON error when the rate limiter refuses the client", async () => {
  const keys: string[] = [];
  const env = {
    SCAN_LIMITER: {
      limit: async ({ key }: { key: string }) => {
        keys.push(key);
        return { success: false };
      },
    },
  };
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", {
      method: "POST",
      body: "{}",
      headers: { "CF-Connecting-IP": "203.0.113.9" },
    }),
    env,
  );
  expect(resp.status).toBe(429);
  expect(resp.headers.get("Access-Control-Allow-Origin")).toBe("*");
  expect(((await resp.json()) as { error: string }).error).toContain("too many");
  expect(keys).toEqual(["203.0.113.9"]);
});

test("/api/scan keys IPv6 clients on their /64 and IPv4 clients on the full address", async () => {
  const keys: string[] = [];
  const env = {
    SCAN_LIMITER: {
      limit: async ({ key }: { key: string }) => {
        keys.push(key);
        return { success: true };
      },
    },
  };
  for (const ip of ["2001:DB8:1:2:3:4:5:6", "2001:db8:1:2::9", "2001:db8::1", "::1", "198.51.100.7"]) {
    await worker.fetch(
      new Request("https://w.test/api/scan", { method: "POST", body: "{}", headers: { "CF-Connecting-IP": ip } }),
      env,
    );
  }
  expect(keys).toEqual(["2001:db8:1:2::/64", "2001:db8:1:2::/64", "2001:db8:0:0::/64", "0:0:0:0::/64", "198.51.100.7"]);
});

test("/api/scan falls back to an in-memory limit of 10 a minute when the binding is absent", async () => {
  const scan = () =>
    worker.fetch(
      new Request("https://w.test/api/scan", {
        method: "POST",
        body: "{}",
        headers: { "CF-Connecting-IP": "203.0.113.50" },
      }),
    );
  for (let i = 0; i < 10; i++) expect((await scan()).status).toBe(400);
  expect((await scan()).status).toBe(429);
});

test("/api/scan proceeds when the rate limiter allows the client", async () => {
  const env = { SCAN_LIMITER: { limit: async () => ({ success: true }) } };
  const resp = await worker.fetch(new Request("https://w.test/api/scan", { method: "POST", body: "{}" }), env);
  expect(resp.status).toBe(400);
});

test("/api/scan analyses only the first 1MB of HTML", async () => {
  const scanH2 = async (padding: number) => {
    globalThis.fetch = (async () =>
      new Response(`<html><body>${"a".repeat(padding)}<h2>Late</h2></body></html>`)) as unknown as typeof fetch;
    const resp = await worker.fetch(
      new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "https://example.com/" }) }),
    );
    const scan = (await resp.json()) as { seo: { checks: { id: string; pass: boolean }[] } };
    return scan.seo.checks.find((c) => c.id === "h2_exists")?.pass;
  };
  expect([await scanH2(1000), await scanH2(1_100_000)]).toEqual([true, false]);
});

test("/api/scan resolves links against the final URL after redirects", async () => {
  globalThis.fetch = (async (input: RequestInfo | URL) =>
    String(input) === "https://example.com/a"
      ? new Response(null, { status: 301, headers: { location: "/dir/page" } })
      : new Response('<html><body><a href="next">next</a></body></html>')) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "https://example.com/a" }) }),
  );
  const scan = (await resp.json()) as { url: string; internalLinks: string[] };
  expect([scan.url, scan.internalLinks]).toEqual(["https://example.com/a", ["https://example.com/dir/next"]]);
});

const allowAll = { SCAN_LIMITER: { limit: async () => ({ success: true }) } };

test("/api/scan of its own host scans the real UI in-process instead of fetching", async () => {
  globalThis.fetch = (async () => {
    throw new Error("network fetch to own host");
  }) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://findable.page/api/scan", {
      method: "POST",
      body: JSON.stringify({ url: "https://findable.page/", site: true }),
    }),
    allowAll,
  );
  const scan = (await resp.json()) as {
    seo: { score: number; checks: { id: string; pass: boolean; label: string }[] };
  };
  const check = (id: string) => scan.seo.checks.find((c) => c.id === id);
  expect(resp.status).toBe(200);
  expect(check("title_exists")?.label).toContain("findable");
  expect(Number.parseInt(check("word_count")?.label ?? "0", 10)).toBeGreaterThan(150);
  expect(check("x_content_type")?.pass).toBe(true);
  expect(scan.seo.score).toBeGreaterThan(40);
});

test("/api/scan refuses an own-host /api/ target instead of recursing", async () => {
  let fetched = false;
  globalThis.fetch = (async () => {
    fetched = true;
    return new Response("");
  }) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://findable.page/api/scan", {
      method: "POST",
      body: JSON.stringify({ url: "https://findable.page/api/scan" }),
    }),
    allowAll,
  );
  expect([resp.status, fetched]).toEqual([400, false]);
});

async function directAnswersPass(body: string): Promise<boolean> {
  globalThis.fetch = (async () => new Response(`<html><body>${body}</body></html>`)) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "https://example.com/" }) }),
  );
  const scan = (await resp.json()) as { aeo: { checks: { id: string; pass: boolean }[] } };
  return scan.aeo.checks.find((c) => c.id === "aeo_direct_answers")?.pass ?? false;
}

const words = (n: number) => Array.from({ length: n }, () => "word").join(" ");

test("aeo_direct_answers needs one 30 to 80 word paragraph, like the Go engine", async () => {
  expect(await directAnswersPass(`<div>${words(200)}</div>`)).toBe(false);
  expect(await directAnswersPass(`<p>${words(29)}</p><p>${words(81)}</p>`)).toBe(false);
  expect(await directAnswersPass(`<p>${words(30)}</p>`)).toBe(true);
  expect(await directAnswersPass(`<p>${words(80)}</p>`)).toBe(true);
});

test("there is no fake scan-count endpoint", async () => {
  const resp = await worker.fetch(new Request("https://w.test/api/scan-count"));
  expect(resp.status).not.toBe(200);
});

test("/api/scan answers JSON, not an unhandled throw, when the client IP header is malformed", async () => {
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", {
      method: "POST",
      body: "{}",
      headers: { "CF-Connecting-IP": "1:2:3:4:5:6:7:8:9::1" },
    }),
  );
  expect(resp.headers.get("Content-Type")).toBe("application/json");
  expect(typeof ((await resp.json()) as { error: string }).error).toBe("string");
});

test("/api/scan treats a bare host as https", async () => {
  const fetched: string[] = [];
  globalThis.fetch = (async (input: RequestInfo | URL) => {
    fetched.push(String(input));
    return new Response("<html><head><title>t</title></head><body></body></html>");
  }) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: " example.com/x " }) }),
    { SCAN_LIMITER: { limit: async () => ({ success: true }) } },
  );
  expect(resp.status).toBe(200);
  expect(fetched[0]).toBe("https://example.com/x");
});

test("/api/scan answers 400 for bodies it cannot use", async () => {
  const bodies = [
    "not json",
    "[]",
    "null",
    JSON.stringify({ url: 5 }),
    JSON.stringify({ url: "example.com", keyphrase: 7 }),
  ];
  const env = { SCAN_LIMITER: { limit: async () => ({ success: true }) } };
  for (const body of bodies) {
    const resp = await worker.fetch(new Request("https://w.test/api/scan", { method: "POST", body }), env);
    expect([body, resp.status]).toEqual([body, 400]);
    expect(((await resp.json()) as { error: string }).error).toBeTruthy();
  }
});

test("/api/scan finds JSON-LD types in top-level arrays, @type arrays and nested nodes", async () => {
  const ld = JSON.stringify([
    { "@type": ["FAQPage", "WebPage"] },
    { "@type": "Article", author: { "@type": "Person", name: "A" }, publisher: { "@type": "Organization" } },
  ]);
  globalThis.fetch = (async () =>
    new Response(
      `<html><head><title>t</title><script type="application/ld+json">${ld}</script></head><body></body></html>`,
    )) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "https://example.com/" }) }),
    allowAll,
  );
  const scan = (await resp.json()) as { aeo: { checks: Check[] }; geo: { checks: Check[] } };
  const pass = (id: string) => [...scan.aeo.checks, ...scan.geo.checks].find((c) => c.id === id)?.pass;
  expect([pass("aeo_faq"), pass("geo_author_schema"), pass("geo_org_schema")]).toEqual([true, true, true]);
});

test("/api/scan counts a heading that ends in a question mark as a question heading", async () => {
  globalThis.fetch = (async () =>
    new Response(
      "<html><head><title>t</title></head><body><h2>Is it free?</h2></body></html>",
    )) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "https://example.com/" }) }),
    allowAll,
  );
  const scan = (await resp.json()) as { aeo: { checks: Check[] } };
  expect(scan.aeo.checks.find((c) => c.id === "aeo_question_headings")?.pass).toBe(true);
});

test("/api/scan does not count a Person nested under review as page author schema", async () => {
  const ld = JSON.stringify({ "@type": "Product", review: { "@type": "Review", author: { "@type": "Person" } } });
  globalThis.fetch = (async () =>
    new Response(
      `<html><head><title>t</title><script type="application/ld+json">${ld}</script></head><body></body></html>`,
    )) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "https://example.com/" }) }),
    allowAll,
  );
  const scan = (await resp.json()) as { geo: { checks: Check[] } };
  expect(scan.geo.checks.find((c) => c.id === "geo_author_schema")?.pass).toBe(false);
});

test("/api/scan matches the Go engine: what/how/why only count at the start of a heading", async () => {
  globalThis.fetch = (async () =>
    new Response(
      "<html><head><title>t</title></head><body><h2>Pricing: how it works</h2></body></html>",
    )) as unknown as typeof fetch;
  const resp = await worker.fetch(
    new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "https://example.com/" }) }),
    allowAll,
  );
  const scan = (await resp.json()) as { aeo: { checks: Check[] } };
  expect(scan.aeo.checks.find((c) => c.id === "aeo_question_headings")?.pass).toBe(false);
});
