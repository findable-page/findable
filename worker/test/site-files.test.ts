import { expect, test } from "bun:test";
import worker from "../src/index";
import { robotsAllows } from "../src/site";

const ORIGIN = "https://w.test";
const get = (path: string) => worker.fetch(new Request(`${ORIGIN}${path}`));

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

test("site files answer GET only", async () => {
  const resp = await worker.fetch(new Request(`${ORIGIN}/robots.txt`, { method: "POST" }));
  expect(resp.status).toBe(404);
});
