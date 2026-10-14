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
