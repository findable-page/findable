import { expect, test } from "bun:test";
import worker from "../src/index";

async function sha256(text: string): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(text));
  return `'sha256-${Buffer.from(digest).toString("base64")}'`;
}

test("the UI is served with a CSP that hashes every inline block and refuses framing", async () => {
  const resp = await worker.fetch(new Request("https://w.test/", { headers: { Accept: "text/html" } }));
  const html = await resp.text();
  const csp = resp.headers.get("Content-Security-Policy") ?? "";
  const directive = (name: string) => csp.split(";").find((d) => d.trim().startsWith(`${name} `)) ?? "";
  const scripts = [...html.matchAll(/<script(?:\s[^>]*)?>([\s\S]*?)<\/script>/g)].map((m) => m[1] ?? "");
  const styles = [...html.matchAll(/<style(?:\s[^>]*)?>([\s\S]*?)<\/style>/g)].map((m) => m[1] ?? "");
  expect(scripts.length).toBeGreaterThan(0);
  for (const s of scripts) expect(directive("script-src")).toContain(await sha256(s));
  for (const s of styles) expect(directive("style-src-elem")).toContain(await sha256(s));
  expect(directive("script-src")).not.toContain("unsafe-inline");
  expect(directive("frame-ancestors")).toContain("'none'");
  expect(directive("connect-src")).toContain("'self'");
  expect(resp.headers.get("X-Content-Type-Options")).toBe("nosniff");
});

test("GET / serves the UI whatever the Accept header says", async () => {
  for (const headers of [{ Accept: "*/*" }, {}]) {
    const resp = await worker.fetch(new Request("https://w.test/", { headers }));
    expect([resp.status, resp.headers.get("Content-Type")]).toEqual([200, "text/html; charset=utf-8"]);
  }
});

test("the UI CSP lets the Cloudflare Web Analytics beacon load and report", async () => {
  const csp = (await worker.fetch(new Request("https://w.test/"))).headers.get("Content-Security-Policy") ?? "";
  const directive = (name: string) => csp.split(";").find((d) => d.trim().startsWith(`${name} `)) ?? "";
  expect(directive("script-src")).toContain("https://static.cloudflareinsights.com");
  expect(directive("connect-src")).toContain("https://cloudflareinsights.com");
});

test("the UI takes bare domains, offers Share and keeps the ?url= autorun", async () => {
  const html = await (await worker.fetch(new Request("https://w.test/"))).text();
  const input = html.match(/<input[^>]*id="urlInput"[^>]*>/)?.[0] ?? "";
  for (const attr of ['type="text"', 'inputmode="url"', 'autocapitalize="off"']) expect(input).toContain(attr);
  expect(html).toContain('id="btnShareReport"');
  expect(html).toMatch(/id="shareStatus"[^>]*aria-live="polite"/);
  expect(html).toContain("navigator.share");
  expect(html).toContain('"https://" +');
  expect(html).toContain('urlParams.get("url")');
  expect(html).toContain("/og.jpg");
  expect(html).not.toContain("/og.png");
});

test("toMarkdown renders the report and escapes table cells", async () => {
  const html = await (await worker.fetch(new Request("https://w.test/"))).text();
  const src = html.split("// toMarkdown:begin")[1]?.split("// toMarkdown:end")[0] ?? "";
  const toMarkdown = new Function(`${src}; return toMarkdown;`)() as (r: unknown) => string;
  const md = toMarkdown({
    url: "https://example.com",
    link: "https://w.test/?url=https%3A%2F%2Fexample.com",
    date: "2026-10-15",
    overall: 72,
    grade: "B Fair",
    scores: { SEO: 80, AEO: 60, GEO: 70 },
    top: [{ label: "Add a title", worth: 2.5 }],
    checks: [
      { id: "a", cat: "SEO", label: "Has a|pipe", pass: false, worth: 1 },
      { id: "b", cat: "AEO", label: "Fine", pass: true, worth: 1 },
      { id: "c", cat: "GEO", label: "Skipped", applicable: false, pass: false, worth: 1 },
    ],
    advice: { a: "line one\nline two" },
  });
  expect(md).toContain("# findable report: https://example.com");
  expect(md).toContain("Overall: **72/100** (B Fair)");
  expect(md).toContain("| SEO | 80 |");
  expect(md).toContain("1. Add a title (+2.5 pts)");
  expect(md).toContain("| FAIL | Has a\\|pipe | line one line two |");
  expect(md).toContain("| PASS | Fine |  |");
  expect(md).toContain("| N/A | Skipped |  |");
  expect(md).toContain("[Share this report](https://w.test/?url=https%3A%2F%2Fexample.com)");
});

test("registerWebMcp exposes a scan_url tool that calls /api/scan and summarises the result", async () => {
  const html = await (await worker.fetch(new Request("https://w.test/"))).text();
  const src = html.split("// webmcp:begin")[1]?.split("// webmcp:end")[0] ?? "";
  type Tool = {
    name: string;
    inputSchema: { required: string[] };
    execute: (a: unknown) => Promise<{ content: { text: string }[] }>;
  };
  const registerWebMcp = new Function(`${src}; return registerWebMcp;`)() as (
    mc: unknown,
    f: (u: string, init: RequestInit) => Promise<Response>,
  ) => void;
  const tools: Tool[] = [];
  const calls: [string, string][] = [];
  const category = (score: number, label: string, pass: boolean) => ({ score, checks: [{ id: label, label, pass }] });
  registerWebMcp({ registerTool: (t: Tool) => tools.push(t) }, async (u, init) => {
    calls.push([u, String(init.body)]);
    return Response.json({
      overall: 72,
      seo: category(80, "Has a title", true),
      aeo: category(60, "FAQ schema", false),
      geo: category(70, "Author schema", false),
    });
  });
  expect(tools.map((t) => t.name)).toEqual(["scan_url"]);
  expect(tools[0]?.inputSchema.required).toEqual(["url"]);
  const out = JSON.parse((await tools[0]?.execute({ url: "example.com", keyphrase: "seo" }))?.content[0]?.text ?? "{}");
  expect(calls).toEqual([["/api/scan", JSON.stringify({ url: "example.com", keyphrase: "seo" })]]);
  expect(out).toEqual({ overall: 72, seo: 80, aeo: 60, geo: 70, failing: ["FAQ schema", "Author schema"] });
  expect(() => registerWebMcp(undefined, fetch)).not.toThrow();
});

test("the WebMCP scan_url tool reports a plain-text API error instead of throwing", async () => {
  const html = await (await worker.fetch(new Request("https://w.test/"))).text();
  const src = html.split("// webmcp:begin")[1]?.split("// webmcp:end")[0] ?? "";
  const registerWebMcp = new Function(`${src}; return registerWebMcp;`)() as (mc: unknown, f: unknown) => void;
  let execute: ((a: unknown) => Promise<{ content: { text: string }[] }>) | undefined;
  registerWebMcp(
    {
      registerTool: (t: { execute: typeof execute }) => {
        execute = t.execute;
        return Promise.resolve();
      },
    },
    async () => new Response("upstream fetch failed\n", { status: 502 }),
  );
  const out = JSON.parse((await execute?.({ url: "example.com" }))?.content[0]?.text ?? "{}");
  expect(out).toEqual({ error: "upstream fetch failed" });
});

test("the crawl report has a most-common-issues panel with its i18n keys in every language", async () => {
  const html = await (await worker.fetch(new Request("https://w.test/", { headers: { Accept: "text/html" } }))).text();
  expect(html).toContain('id="crawlCommon"');
  const keys = ["footPrivacy", "footTerms", "commonIssuesHeading", "commonIssuesCount"];
  for (const lang of ["en", "de", "es", "tr", "ru"]) {
    const start = html.indexOf(`\n      ${lang}: {`, html.indexOf("const I18N"));
    expect(start).toBeGreaterThan(0);
    const block = html.slice(start, html.indexOf("\n      },", start));
    for (const key of keys) expect(block).toContain(`${key}:`);
  }
  expect(html).toContain("{c} of {n} pages");
});
