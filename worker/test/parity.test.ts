import { afterEach, expect, test } from "bun:test";
import worker, { type Check } from "../src/index";

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

async function scanChecks(html: string, url = "https://example.com/"): Promise<Map<string, Check>> {
  globalThis.fetch = (async () => new Response(html)) as unknown as typeof fetch;
  const env = { SCAN_LIMITER: { limit: async () => ({ success: true }) } };
  const request = new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url }) });
  const resp = await worker.fetch(request, env);
  const scan = (await resp.json()) as { seo: { checks: Check[] }; aeo: { checks: Check[] }; geo: { checks: Check[] } };
  return new Map([...scan.seo.checks, ...scan.aeo.checks, ...scan.geo.checks].map((c) => [c.id, c]));
}

async function passes(id: string, body: string, head = ""): Promise<boolean> {
  const checks = await scanChecks(`<!DOCTYPE html><html><head>${head}</head><body>${body}</body></html>`);
  const check = checks.get(id);
  if (!check) throw new Error(`check ${id} missing`);
  return check.pass;
}

const words = (n: number) => Array(n).fill("word").join(" ");

test("img_alt fails when an image lacks alt text and passes with none or all present", async () => {
  expect(await passes("img_alt", "<p>no images</p>")).toBe(true);
  expect(await passes("img_alt", '<img src="a.png" alt="A cat">')).toBe(true);
  expect(await passes("img_alt", '<img src="a.png">')).toBe(false);
  expect(await passes("img_alt", '<img src="a.png" alt="  ">')).toBe(false);
});

test("img_dimensions needs both width and height on every image", async () => {
  expect(await passes("img_dimensions", "<p>none</p>")).toBe(true);
  expect(await passes("img_dimensions", '<img src="a.png" width="10" height="10">')).toBe(true);
  expect(await passes("img_dimensions", '<img src="a.png" width="10">')).toBe(false);
});

test("heading_hierarchy fails when a level is skipped", async () => {
  expect(await passes("heading_hierarchy", "<h1>A</h1><h2>B</h2><h3>C</h3><h2>D</h2>")).toBe(true);
  expect(await passes("heading_hierarchy", "<h1>A</h1><h3>C</h3>")).toBe(false);
});

test("dom_size fails at 1500 elements", async () => {
  expect(await passes("dom_size", "<div></div>".repeat(100))).toBe(true);
  expect(await passes("dom_size", "<div></div>".repeat(1500))).toBe(false);
});

test("duplicate_ids fails when an id repeats", async () => {
  expect(await passes("duplicate_ids", '<div id="a"></div><div id="b"></div>')).toBe(true);
  expect(await passes("duplicate_ids", '<div id="a"></div><p id="a"></p>')).toBe(false);
});

test("empty_hrefs fails on missing, empty or # hrefs", async () => {
  expect(await passes("empty_hrefs", '<a href="/x">x</a>')).toBe(true);
  expect(await passes("empty_hrefs", '<a href="#">x</a>')).toBe(false);
  expect(await passes("empty_hrefs", '<a href="">x</a>')).toBe(false);
  expect(await passes("empty_hrefs", "<a>x</a>")).toBe(false);
});

test("link_noopener fails on target=_blank without noopener or noreferrer", async () => {
  expect(await passes("link_noopener", '<a href="https://o.com" target="_blank" rel="noopener">x</a>')).toBe(true);
  expect(await passes("link_noopener", '<a href="https://o.com" target="_blank" rel="noreferrer">x</a>')).toBe(true);
  expect(await passes("link_noopener", '<a href="https://o.com" target="_blank">x</a>')).toBe(false);
  expect(await passes("link_noopener", '<a href="https://o.com">x</a>')).toBe(true);
});

test("meta_refresh fails when a refresh redirect is declared", async () => {
  expect(await passes("meta_refresh", "<p>x</p>")).toBe(true);
  expect(await passes("meta_refresh", "<p>x</p>", '<meta http-equiv="refresh" content="0;url=/a">')).toBe(false);
});

test("text_html_ratio needs 15% visible text", async () => {
  expect(await passes("text_html_ratio", `<p>${words(50)}</p>`)).toBe(true);
  expect(await passes("text_html_ratio", `<div class="${words(100)}">hi</div>`)).toBe(false);
});

test("aria_landmarks needs both main and nav", async () => {
  expect(await passes("aria_landmarks", "<nav>n</nav><main>m</main>")).toBe(true);
  expect(await passes("aria_landmarks", "<main>m</main>")).toBe(false);
});

test("aeo_reading_level and aeo_short_sentences follow average sentence length", async () => {
  const short = `<p>${words(10)}. ${words(10)}.</p>`;
  const mid = `<p>${words(33)}. ${words(33)}.</p>`;
  const long = `<p>${words(60)}.</p>`;
  expect(await passes("aeo_reading_level", short)).toBe(true);
  expect(await passes("aeo_short_sentences", short)).toBe(true);
  expect(await passes("aeo_reading_level", mid)).toBe(true);
  expect(await passes("aeo_short_sentences", mid)).toBe(false);
  expect(await passes("aeo_reading_level", long)).toBe(false);
});

test("aeo_semantic_html needs four semantic elements", async () => {
  expect(await passes("aeo_semantic_html", "<header>h</header><nav>n</nav><main>m</main><footer>f</footer>")).toBe(
    true,
  );
  expect(await passes("aeo_semantic_html", "<header>h</header><main>m</main>")).toBe(false);
});

const speakableLd = (ld: string) => `<script type="application/ld+json">${ld}</script>`;

test("aeo_speakable passes only when JSON-LD declares speakable", async () => {
  const withIt = speakableLd('{"@type":"WebPage","speakable":{"@type":"SpeakableSpecification","cssSelector":["h1"]}}');
  const inGraph = speakableLd('{"@graph":[{"@type":"WebPage","speakable":{"cssSelector":["h1"]}}]}');
  const without = speakableLd('{"@type":"WebPage"}');
  expect(await passes("aeo_speakable", "<p>x</p>", withIt)).toBe(true);
  expect(await passes("aeo_speakable", "<p>x</p>", inGraph)).toBe(true);
  expect(await passes("aeo_speakable", "<p>x</p>", without)).toBe(false);
  expect(await passes("aeo_speakable", "<p>x</p>")).toBe(false);
});

test("geo_passage_citability needs a 40-120 word paragraph", async () => {
  expect(await passes("geo_passage_citability", `<p>${words(60)}</p>`)).toBe(true);
  expect(await passes("geo_passage_citability", `<p>${words(10)}</p><p>${words(130)}</p>`)).toBe(false);
});

test("geo_entity_markup needs two schema types", async () => {
  const two = speakableLd('[{"@type":"WebPage"},{"@type":"Organization"}]');
  const one = speakableLd('{"@type":"WebPage"}');
  expect(await passes("geo_entity_markup", "<p>x</p>", two)).toBe(true);
  expect(await passes("geo_entity_markup", "<p>x</p>", one)).toBe(false);
});

test("geo_intent_clarity needs two long title words in the first heading", async () => {
  const head = "<title>Rust Performance Guide</title>";
  expect(await passes("geo_intent_clarity", "<h1>Rust performance tuning</h1>", head)).toBe(true);
  expect(await passes("geo_intent_clarity", "<h1>Unrelated heading</h1>", head)).toBe(false);
  expect(await passes("geo_intent_clarity", "<h1>Rust performance</h1>")).toBe(false);
});

test("geo_format_variety needs three formatting types", async () => {
  expect(await passes("geo_format_variety", "<ul><li>a</li></ul><table></table><blockquote>q</blockquote>")).toBe(true);
  expect(await passes("geo_format_variety", "<ul><li>a</li></ul><table></table>")).toBe(false);
});

test("geo_snippet_control fails when nosnippet is set", async () => {
  expect(await passes("geo_snippet_control", "<p>x</p>")).toBe(true);
  expect(await passes("geo_snippet_control", "<p>x</p>", '<meta name="robots" content="nosnippet">')).toBe(false);
});

test("geo_passage_citability evaluates paragraphs on a page of 20k unclosed <p> within 1s", async () => {
  const body = `${"<p>x\n".repeat(20000)}<p>${words(60)}`;
  const start = performance.now();
  const ok = await passes("geo_passage_citability", body);
  expect(performance.now() - start).toBeLessThan(1000);
  expect(ok).toBe(true);
});

test("unquoted attributes are read for images, links and ids", async () => {
  expect(await passes("img_alt", "<img src=/x.png alt=Logo width=800 height=600>")).toBe(true);
  expect(await passes("img_dimensions", "<img src=/x.png alt=Logo width=800 height=600>")).toBe(true);
  expect(await passes("empty_hrefs", "<a href=/about/>About</a>")).toBe(true);
  expect(await passes("link_noopener", "<a href=/x target=_blank rel=noopener>x</a>")).toBe(true);
  expect(await passes("link_noopener", "<a href=/x target=_blank>x</a>")).toBe(false);
  expect(await passes("duplicate_ids", "<div id=foo></div><p id=foo></p>")).toBe(false);
});

test("tags inside noscript, script, style and comments are ignored", async () => {
  const pixel = '<noscript><img height="1" width="1" src="https://www.facebook.com/tr?id=1"/></noscript>';
  const good = '<img src="a.png" alt="A cat" width="1" height="1">';
  expect(await passes("img_alt", pixel + good)).toBe(true);
  expect(await passes("img_dimensions", "<!-- <img src=x> --><script>var a='<img src=y>'</script>" + good)).toBe(true);
});

test("text_html_ratio uses byte lengths so multibyte text matches the Go engine", async () => {
  // 20 Bengali chars = 60 bytes; html is ~ 70 bytes of markup. Char ratio < 15%, byte ratio > 15%.
  const bengali = "আমারসোনারবাংলাআমিতোমায়";
  const body = `<p>${bengali.repeat(3)}</p>`;
  const checks = await scanChecks(`<html><body>${body}</body></html>`);
  const label = checks.get("text_html_ratio")?.label ?? "";
  const bytes = new TextEncoder().encode(`<html><body>${body}</body></html>`).length;
  const textBytes = new TextEncoder().encode(bengali.repeat(3)).length;
  expect(label).toContain(`${((textBytes / bytes) * 100).toFixed(1)}%`);
});

test("custom elements do not count as semantic HTML or headings", async () => {
  const custom = "<details-modal></details-modal><header-drawer></header-drawer><nav-menu></nav-menu><main-x></main-x>";
  const checks = await scanChecks(`<html><body>${custom}</body></html>`);
  expect(checks.get("aeo_semantic_html")?.label).toContain("0 semantic");
  expect(await passes("aria_landmarks", "<main-x></main-x><nav-menu></nav-menu>")).toBe(false);
});

test("aeo_speakable ignores speakable nested under author", async () => {
  const ld = (o: object) => `<script type="application/ld+json">${JSON.stringify(o)}</script>`;
  expect(await passes("aeo_speakable", "<p>x</p>", ld({ "@type": "Article", speakable: {} }))).toBe(true);
  expect(
    await passes("aeo_speakable", "<p>x</p>", ld({ "@type": "Article", author: { "@type": "Person", speakable: {} } })),
  ).toBe(false);
});
