import { afterEach, expect, test } from "bun:test";
import worker from "../src/index";
import { robotsAllows } from "../src/site";
import { HEAD_SLOT_MARKER, headSlot, injectHead } from "../src/site-files";

const ORIGIN = "https://w.test";
const get = (path: string) => worker.fetch(new Request(`${ORIGIN}${path}`));
const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

test("robots.txt is plain text, allows the four AI bots, hides /api/ and names the sitemap", async () => {
  const resp = await get("/robots.txt");
  const body = await resp.text();
  expect(resp.status).toBe(200);
  expect(resp.headers.get("Content-Type")).toBe("text/plain; charset=utf-8");
  for (const bot of ["GPTBot", "ClaudeBot", "PerplexityBot", "Google-Extended"]) {
    expect(robotsAllows(body, bot, "/")).toBe(true);
    expect(robotsAllows(body, bot, "/api/scan")).toBe(false);
  }
  expect(body).toContain(`Sitemap: ${ORIGIN}/sitemap.xml`);
});

test("sitemap.xml is XML and lists the home page", async () => {
  const resp = await get("/sitemap.xml");
  const body = await resp.text();
  expect(resp.status).toBe(200);
  expect(resp.headers.get("Content-Type")).toBe("application/xml; charset=utf-8");
  expect(body).toStartWith('<?xml version="1.0" encoding="UTF-8"?>');
  expect(body).toContain('<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">');
  expect(body).toContain(`<loc>${ORIGIN}/</loc>`);
  expect(body).toEndWith("</urlset>\n");
});

test("llms.txt follows the llmstxt.org shape and llms-full.txt documents the API", async () => {
  const short = await get("/llms.txt");
  const full = await get("/llms-full.txt");
  const shortBody = await short.text();
  const fullBody = await full.text();
  expect([short.status, full.status]).toEqual([200, 200]);
  expect(short.headers.get("Content-Type")).toBe("text/plain; charset=utf-8");
  expect(full.headers.get("Content-Type")).toBe("text/plain; charset=utf-8");
  expect(shortBody).toStartWith("# findable\n\n> ");
  expect(shortBody).toContain(`](${ORIGIN}/)`);
  expect(fullBody).toContain("POST /api/scan");
  expect(fullBody).toContain("npx -y findable-mcp");
});

test("the HTML page carries HSTS and a referrer policy next to the existing hardening", async () => {
  const resp = await get("/");
  expect(resp.headers.get("Strict-Transport-Security")).toBe("max-age=31536000; includeSubDomains");
  expect(resp.headers.get("Referrer-Policy")).toBe("strict-origin-when-cross-origin");
  expect(resp.headers.get("X-Content-Type-Options")).toBe("nosniff");
  expect(resp.headers.get("X-Frame-Options")).toBe("DENY");
  expect(resp.headers.get("Content-Security-Policy")).toContain("frame-ancestors 'none'");
});

function jsonLdBlocks(html: string): Record<string, unknown>[] {
  const scripts = html.matchAll(/<script type="application\/ld\+json">([\s\S]*?)<\/script>/g);
  return [...scripts].map((m) => JSON.parse(m[1] ?? "") as Record<string, unknown>);
}

test("headSlot emits the top-level types and sameAs links the scanner looks for", () => {
  const blocks = jsonLdBlocks(headSlot(ORIGIN));
  const types = blocks.map((b) => b["@type"]);
  expect(types).toEqual(["SoftwareApplication", "Person", "Organization"]);
  const app = blocks[0] as Record<string, unknown>;
  expect(app.applicationCategory).toBe("DeveloperApplication");
  expect(app.url).toBe(ORIGIN);
  expect(app.offers).toMatchObject({ price: "0" });
  expect(app.sameAs).toEqual(["https://github.com/nadimtuhin/findable"]);
  for (const b of blocks.slice(1)) expect((b.sameAs as string[]).length).toBeGreaterThan(0);
  expect(blocks[1]).toMatchObject({ name: "Nadim Tuhin", sameAs: ["https://github.com/nadimtuhin"] });
});

test("headSlot adds FAQPage only when FAQ items are passed and escapes angle brackets", () => {
  expect(headSlot(ORIGIN)).not.toContain("FAQPage");
  const slot = headSlot(ORIGIN, [{ question: "What is it?", answer: "A </script> tester" }]);
  const faq = jsonLdBlocks(slot).find((b) => b["@type"] === "FAQPage");
  expect(faq).toMatchObject({
    mainEntity: [{ name: "What is it?", acceptedAnswer: { text: "A </script> tester" } }],
  });
  expect(slot.match(/<\/script>/g)?.length).toBe(4);
});

test("injectHead replaces the marker and leaves a page without it untouched", () => {
  const slot = headSlot(ORIGIN);
  const page = `<head><title>t</title>${HEAD_SLOT_MARKER}</head>`;
  expect(injectHead(page, slot)).toBe(`<head><title>t</title>${slot}</head>`);
  const bare = "<head><title>t</title></head>";
  expect(injectHead(bare, slot)).toBe(bare);
});

test("site files answer GET only", async () => {
  const resp = await worker.fetch(new Request(`${ORIGIN}/robots.txt`, { method: "POST" }));
  expect(resp.status).toBe(404);
});

test("the worker's llms.txt states the check counts the worker really runs, and the UI states none", async () => {
  globalThis.fetch = (async () => new Response("<html></html>")) as unknown as typeof fetch;
  const scan = (await (
    await worker.fetch(
      new Request(`${ORIGIN}/api/scan`, { method: "POST", body: JSON.stringify({ url: "https://example.com/" }) }),
    )
  ).json()) as Record<"seo" | "aeo" | "geo", { checks: unknown[] }>;
  const [s, a, g] = [scan.seo, scan.aeo, scan.geo].map((c) => c.checks.length);
  const llms = await (await get("/llms.txt")).text();
  expect(llms).toContain(`${(s ?? 0) + (a ?? 0) + (g ?? 0)} page checks (${s} SEO, ${a} AEO, ${g} GEO)`);
  const ui = await (await get("/")).text();
  expect(ui).not.toMatch(/\b(72|36|19) (page )?checks/);
  expect(ui).not.toContain("<td>36</td>");
});
