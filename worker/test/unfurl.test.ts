import { afterEach, expect, test } from "bun:test";
import worker, { type Env, prettyShareUrl } from "../src/index";
import { QUIPS, shareDescription } from "../src/quips";

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

const ENV = { SCAN_LIMITER: { limit: async () => ({ success: true }) } } as unknown as Env;
const BOT = { "User-Agent": "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)" };
const meta = (html: string, attr: string, key: string) =>
  html.match(new RegExp(`<meta ${attr}="${key}" content="([^"]*)"`))?.[1];

test("a link-preview bot fetching /?url= gets the page's scores in the share tags", async () => {
  globalThis.fetch = (async () =>
    new Response("<html><head><title>Hi</title></head><body></body></html>")) as unknown as typeof fetch;
  const scan = (await (
    await worker.fetch(
      new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url: "https://example.com/" }) }),
      ENV,
    )
  ).json()) as { overall: number; seo: { score: number }; aeo: { score: number }; geo: { score: number } };

  const html = await (
    await worker.fetch(new Request("https://w.test/?url=https%3A%2F%2Fexample.com%2F", { headers: BOT }), ENV)
  ).text();

  const title = meta(html, "property", "og:title") ?? "";
  expect(title.startsWith(`example.com scored ${scan.overall}/100. `)).toBe(true);
  expect(meta(html, "name", "twitter:title")).toBe(title);
  expect(html).toContain(`<title>${title}</title>`);
  expect(meta(html, "property", "og:description")).toBe(
    shareDescription({ seo: scan.seo.score, aeo: scan.aeo.score, geo: scan.geo.score }, "example.com"),
  );
  expect(meta(html, "property", "og:url")).toBe("https://findable.page/example.com");
});

test("people opening /?url= get the generic page without a server-side scan", async () => {
  let fetched = false;
  globalThis.fetch = (async () => {
    fetched = true;
    return new Response("");
  }) as unknown as typeof fetch;
  const html = await (
    await worker.fetch(
      new Request("https://w.test/?url=example.com", { headers: { "User-Agent": "Mozilla/5.0" } }),
      ENV,
    )
  ).text();
  expect(fetched).toBe(false);
  expect(meta(html, "property", "og:title")).toBe("findable: is your site findable by Google and AI?");
});

test("a bot asking about a blocked target still gets the generic page", async () => {
  const resp = await worker.fetch(new Request("https://w.test/?url=http%3A%2F%2F127.0.0.1%2F", { headers: BOT }), ENV);
  expect(resp.status).toBe(200);
  expect(meta(await resp.text(), "property", "og:title")).toBe("findable: is your site findable by Google and AI?");
});

test("the preview title jokes about a low score", async () => {
  globalThis.fetch = (async () => new Response("<html></html>")) as unknown as typeof fetch;
  const html = await (await worker.fetch(new Request("https://w.test/?url=example.com", { headers: BOT }), ENV)).text();
  const [, score, quip] = meta(html, "property", "og:title")?.match(/^example\.com scored (\d+)\/100\. (.+)$/) ?? [];
  expect(Number(score)).toBeLessThan(50);
  expect([...QUIPS[3], ...QUIPS[4]]).toContain(quip ?? "");
});

test("share tags escape the scanned host", async () => {
  globalThis.fetch = (async () => new Response("<html></html>")) as unknown as typeof fetch;
  const html = await (
    await worker.fetch(
      new Request(`https://w.test/?url=${encodeURIComponent('https://example.com/"><script>')}`, { headers: BOT }),
      ENV,
    )
  ).text();
  expect(html).not.toContain('"><script>');
});

test("a bot fetching the short /<host> link gets the scores and a short og:url", async () => {
  globalThis.fetch = (async () => new Response("<html></html>")) as unknown as typeof fetch;
  const resp = await worker.fetch(new Request("https://w.test/example.com", { headers: BOT }), ENV);
  const html = await resp.text();
  expect(resp.status).toBe(200);
  expect(meta(html, "property", "og:title")).toMatch(/^example\.com scored \d+\/100\. /);
  expect(meta(html, "property", "og:url")).toBe("https://findable.page/example.com");
});

test("people opening the short /<host> link get the page", async () => {
  const resp = await worker.fetch(new Request("https://w.test/example.com/blog/post"), ENV);
  expect(resp.status).toBe(200);
  expect(resp.headers.get("Content-Type")).toContain("text/html");
});

test("http-only targets keep the query-string share link", () => {
  expect(prettyShareUrl("http://example.com")).toBe("https://findable.page/?url=http%3A%2F%2Fexample.com");
  expect(prettyShareUrl("https://example.com/")).toBe("https://findable.page/example.com");
  expect(prettyShareUrl("example.com/a/b")).toBe("https://findable.page/example.com/a/b");
});

test("paths without a dot still 404", async () => {
  expect((await worker.fetch(new Request("https://w.test/nope"), ENV)).status).toBe(404);
});
