import { describe, expect, test } from "bun:test";
import worker from "../src/index";

const ORIGIN = "https://findable.page";
const get = (path: string) => worker.fetch(new Request(`${ORIGIN}${path}`));
const PATHS = ["/free-seo-scanner", "/aeo", "/wordpress", "/compare"];

function ldBlocks(
  html: string,
): { "@type"?: string; mainEntity?: { name: string; acceptedAnswer: { text: string } }[] }[] {
  return [...html.matchAll(/<script type="application\/ld\+json">([\s\S]*?)<\/script>/g)].map((m) =>
    JSON.parse(m[1] ?? ""),
  );
}

describe("landing pages", () => {
  for (const path of PATHS) {
    test(`${path} is html with its own canonical, title and h1`, async () => {
      const resp = await get(path);
      const html = await resp.text();
      expect(resp.status).toBe(200);
      expect(resp.headers.get("Content-Type")).toBe("text/html; charset=utf-8");
      expect(html).toContain(`<link rel="canonical" href="${ORIGIN}${path}">`);
      expect(html).toMatch(/<h1>[^<]+<\/h1>/);
      expect(html).toContain('href="/"');
      expect(ldBlocks(html).some((b) => b["@type"] === "FAQPage")).toBe(true);
    });
  }

  test("titles and descriptions are unique", async () => {
    const metas = await Promise.all(
      PATHS.map(async (p) => {
        const html = await (await get(p)).text();
        return [html.match(/<title>(.*?)<\/title>/)?.[1], html.match(/name="description" content="(.*?)"/)?.[1]];
      }),
    );
    expect(new Set(metas.map((m) => m[0])).size).toBe(PATHS.length);
    expect(new Set(metas.map((m) => m[1])).size).toBe(PATHS.length);
  });

  test("sitemap.xml and llms.txt list them", async () => {
    const sitemap = await (await get("/sitemap.xml")).text();
    const llms = await (await get("/llms.txt")).text();
    for (const path of PATHS) {
      expect(sitemap).toContain(`<loc>${ORIGIN}${path}</loc>`);
      expect(llms).toContain(`(${ORIGIN}${path})`);
    }
  });
});

describe("homepage", () => {
  test("FAQPage JSON-LD parses and matches the visible questions", async () => {
    const html = await (await get("/")).text();
    const faq = ldBlocks(html).find((b) => b["@type"] === "FAQPage");
    expect(faq?.mainEntity?.length).toBeGreaterThanOrEqual(6);
    const visible = [...html.matchAll(/<details class="faq"><summary[^>]*>(.*?)<\/summary><p[^>]*>(.*?)<\/p>/g)].map(
      (m) => [m[1] ?? "", m[2] ?? ""],
    );
    expect(faq?.mainEntity?.map((e) => [e.name, e.acceptedAnswer.text])).toEqual(visible);
  });

  test("every inline script, JSON-LD included, is allowed by the CSP hashes", async () => {
    const resp = await get("/");
    const html = await resp.text();
    const csp = resp.headers.get("Content-Security-Policy") ?? "";
    for (const m of html.matchAll(/<script(?:\s[^>]*)?>([\s\S]*?)<\/script>/g)) {
      const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(m[1] ?? "")));
      expect(csp).toContain(`'sha256-${btoa(String.fromCharCode(...digest))}'`);
    }
  });
});

describe("self-scan", () => {
  for (const [i, path] of ["/", ...PATHS].entries()) {
    test(`${path} scores well with findable's own engine`, async () => {
      const resp = await worker.fetch(
        new Request(`${ORIGIN}/api/scan`, {
          method: "POST",
          headers: { "CF-Connecting-IP": `198.51.100.${i + 1}` },
          body: JSON.stringify({ url: `${ORIGIN}${path}` }),
        }),
      );
      const result = (await resp.json()) as { overall: number };
      expect(result.overall).toBeGreaterThanOrEqual(75);
    });
  }
});

describe("compare page", () => {
  test("is titled, neutral about where the other tool leads, and in the sitemap", async () => {
    const html = await (await get("/compare")).text();
    expect(html).toContain("<title>findable vs seoscore.tools");
    expect(html).toContain("<table>");
    expect(html).toContain("AI content generator");
    expect(html).toContain("260+");
    expect((await (await get("/sitemap.xml")).text()).includes(`${ORIGIN}/compare`)).toBe(true);
    expect(await (await get("/about")).text()).toContain('href="/compare"');
  });
});
