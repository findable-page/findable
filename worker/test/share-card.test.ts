import { expect, test } from "bun:test";
import worker, { cardScores, type Env } from "../src/index";
import { shareCardHtml } from "../src/share-card";

const ENV = { SCAN_LIMITER: { limit: async () => ({ success: true }) } } as unknown as Env;
const BOT = { "User-Agent": "Twitterbot/1.0" };

test("the preview page points og:image at a card carrying the scan's scores", async () => {
  const realFetch = globalThis.fetch;
  globalThis.fetch = (async () => new Response("<html></html>")) as unknown as typeof fetch;
  try {
    const html = await (await worker.fetch(new Request("https://w.test/example.com", { headers: BOT }), ENV)).text();
    const image = html.match(/<meta property="og:image" content="([^"]*)"/)?.[1] ?? "";
    const params = new URL(image.replaceAll("&amp;", "&")).searchParams;
    expect(image.startsWith("https://findable.page/og/card.png?h=example.com&amp;o=")).toBe(true);
    expect(cardScores(params)?.host).toBe("example.com");
    expect(html).toContain(`<meta name="twitter:image" content="${image}"`);
  } finally {
    globalThis.fetch = realFetch;
  }
});

test("card params outside 0-100 or with a strange host are rejected", () => {
  const ok = { h: "example.com", o: "62", s: "70", a: "50", g: "40" };
  expect(cardScores(new URLSearchParams(ok))).toEqual({ host: "example.com", overall: 62, seo: 70, aeo: 50, geo: 40 });
  expect(cardScores(new URLSearchParams({ ...ok, o: "101" }))).toBeNull();
  expect(cardScores(new URLSearchParams({ ...ok, s: "4.5" }))).toBeNull();
  expect(cardScores(new URLSearchParams({ ...ok, h: "<b>x</b>" }))).toBeNull();
});

test("a bad card request falls back to the static image", async () => {
  const resp = await worker.fetch(new Request("https://w.test/og/card.png?h=x&o=999"), ENV);
  expect(resp.status).toBe(302);
  expect(resp.headers.get("Location")).toBe("https://w.test/og.png");
});

test("the card shows the host and every score", () => {
  const html = shareCardHtml({ host: "example.com", overall: 62, seo: 71, aeo: 55, geo: 43 });
  for (const s of ["example.com", ">62<", ">71<", ">55<", ">43<"]) expect(html).toContain(s);
});

test("a second bot asking about the same host within the hour reuses the first scan", async () => {
  const store = new Map<string, Response>();
  const g = globalThis as { caches?: unknown };
  g.caches = {
    default: {
      match: async (r: Request) => store.get(r.url)?.clone(),
      put: async (r: Request, resp: Response) => void store.set(r.url, resp),
    },
  };
  const realFetch = globalThis.fetch;
  let fetches = 0;
  globalThis.fetch = (async () => {
    fetches++;
    return new Response("<html></html>");
  }) as unknown as typeof fetch;
  try {
    const firstResp = await worker.fetch(new Request("https://w.test/example.com", { headers: BOT }), ENV);
    const first = await firstResp.text();
    const afterFirst = fetches;
    const secondResp = await worker.fetch(new Request("https://w.test/?url=example.com", { headers: BOT }), ENV);
    const second = await secondResp.text();
    expect(afterFirst).toBeGreaterThan(0);
    expect(fetches).toBe(afterFirst);
    expect(second).toBe(first);
    expect(firstResp.headers.get("X-Findable-Cache")).toBe("miss");
    expect(secondResp.headers.get("X-Findable-Cache")).toBe("hit");
  } finally {
    globalThis.fetch = realFetch;
    delete g.caches;
  }
});
