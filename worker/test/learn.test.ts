import { describe, expect, test } from "bun:test";
import { readFileSync } from "node:fs";
import { generateLearnContent, OUTPUT_PATH, renderMarkdown } from "../../scripts/build-learn";
import worker from "../src/index";
import { learnPaths, renderLearnIndex, renderLearnPost } from "../src/learn";
import { learnPosts } from "../src/learn-content";

const ORIGIN = "https://findable.piply.xyz";
let scans = 0;
const get = (path: string) => worker.fetch(new Request(`${ORIGIN}${path}`));

interface Scan {
  seo: { score: number; checks: { id: string; pass: boolean }[] };
  aeo: { score: number; checks: { id: string; pass: boolean }[] };
  geo: { score: number; checks: { id: string; pass: boolean }[] };
}

async function scan(path: string): Promise<Scan> {
  const resp = await worker.fetch(
    new Request(`${ORIGIN}/api/scan`, {
      method: "POST",
      headers: { "CF-Connecting-IP": `203.0.113.${++scans}` },
      body: JSON.stringify({ url: `${ORIGIN}${path}` }),
    }),
  );
  return (await resp.json()) as Scan;
}

const failing = (cat: Scan["seo"]) => cat.checks.filter((c) => !c.pass).map((c) => c.id);

test("learn-content.ts is current with docs/learn", () => {
  expect(readFileSync(OUTPUT_PATH, "utf8")).toBe(generateLearnContent());
});

describe("markdown subset", () => {
  test("escapes raw html and keeps code literal", () => {
    expect(renderMarkdown("a <script>x</script> **b** `<i>`")).toBe(
      "<p>a &lt;script&gt;x&lt;/script&gt; <strong>b</strong> <code>&lt;i&gt;</code></p>",
    );
  });

  test("renders headings, lists, fences, links and tables", () => {
    const html = renderMarkdown(
      ["## Why?", "", "- one", "- two", "", "```sh", "# not a heading", "```", "", "| A | B |", "| - | - |", "| 1 | [x](https://a.test/?q=1&r=2) |"].join("\n"),
    );
    expect(html).toBe(
      [
        '<h2 id="why">Why?</h2>',
        "<ul><li>one</li><li>two</li></ul>",
        "<pre><code># not a heading</code></pre>",
        '<table><thead><tr><th>A</th><th>B</th></tr></thead><tbody><tr><td>1</td><td><a href="https://a.test/?q=1&amp;r=2" rel="noopener">x</a></td></tr></tbody></table>',
      ].join("\n"),
    );
  });

  test("does not turn javascript: links into anchors", () => {
    expect(renderMarkdown("[x](javascript:alert(1))")).not.toContain("<a ");
  });
});

describe("routing", () => {
  test("GET /learn serves the index with security headers", async () => {
    const resp = await get("/learn");
    expect(resp.status).toBe(200);
    expect(resp.headers.get("content-type")).toContain("text/html");
    expect(resp.headers.get("x-content-type-options")).toBe("nosniff");
    expect(await resp.text()).toBe(renderLearnIndex(ORIGIN));
  });

  test("the CSP style hash matches the inline style of every page", async () => {
    for (const path of learnPaths()) {
      const resp = await get(path);
      const html = await resp.text();
      const style = html.match(/<style>([\s\S]*?)<\/style>/)?.[1] ?? "";
      const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(style)));
      expect(resp.headers.get("content-security-policy")).toContain(`'sha256-${btoa(String.fromCharCode(...digest))}'`);
    }
  });

  test("GET /learn/<slug> serves the post", async () => {
    const slug = learnPosts[0]?.slug ?? "";
    const resp = await get(`/learn/${slug}`);
    expect(resp.status).toBe(200);
    expect(await resp.text()).toBe(renderLearnPost(slug, ORIGIN));
  });

  test("unknown slug returns 404 html with the same headers", async () => {
    const resp = await get("/learn/nope");
    expect(resp.status).toBe(404);
    expect(resp.headers.get("content-type")).toContain("text/html");
    expect(resp.headers.get("content-security-policy")).toContain("default-src 'none'");
    expect(await resp.text()).toContain("Page not found");
  });

  test("learnPaths lists the index and every post", () => {
    expect(learnPaths()).toEqual(["/learn", ...learnPosts.map((p) => `/learn/${p.slug}`)]);
  });
});

describe("page markup", () => {
  for (const post of learnPosts) {
    test(`${post.slug} has one h1, canonical, byline and structured data`, () => {
      const html = renderLearnPost(post.slug, ORIGIN) ?? "";
      expect(html.match(/<h1[ >]/g)?.length).toBe(1);
      expect(html).toContain(`<link rel="canonical" href="${ORIGIN}/learn/${post.slug}">`);
      expect(html).toContain(`<time datetime="${post.datePublished}">`);
      expect(html).toContain("Nadim Tuhin");
      expect(html).toContain('"@type":"Article"');
      expect(html.includes('"@type":"FAQPage"')).toBe(post.faq.length > 0);
      expect(html).toContain('<a href="/">');
      expect(html).toContain('href="/learn/');
    });
  }
});

describe("findable audits its own learn pages", () => {
  for (const path of learnPaths()) {
    test(`${path} scores SEO>=90 AEO>=85 GEO>=80`, async () => {
      const result = await scan(path);
      const misses = { seo: failing(result.seo), aeo: failing(result.aeo), geo: failing(result.geo) };
      expect([result.seo.score >= 90, result.aeo.score >= 85, result.geo.score >= 80, misses]).toEqual([
        true,
        true,
        true,
        misses,
      ]);
    });
  }
});

describe("site files list the learn pages", () => {
  test("sitemap.xml and llms.txt include /learn and every post", async () => {
    const sitemap = await (await get("/sitemap.xml")).text();
    const llms = await (await get("/llms.txt")).text();
    for (const path of ["/learn", ...learnPosts.map((p) => `/learn/${p.slug}`)]) {
      expect(sitemap).toContain(`<loc>${ORIGIN}${path}</loc>`);
      expect(llms).toContain(`${ORIGIN}${path}`);
    }
  });

  test("learn pages carry the full security headers", async () => {
    const res = await get("/learn");
    expect(res.headers.get("Referrer-Policy")).toBe("strict-origin-when-cross-origin");
    expect(res.headers.get("Strict-Transport-Security")).toContain("includeSubDomains");
  });
});
