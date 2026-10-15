import { afterEach, expect, test } from "bun:test";
import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import worker, { type Check } from "../src/index";

interface CheckCase {
  url?: string;
  head?: string;
  body?: string;
  doc?: string;
  pad?: number;
  repeat?: number;
  headRepeat?: number;
}

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

function pageOf(c: CheckCase): { url: string; page: string } {
  const url = c.url ?? "https://example.com/page";
  if (c.doc) return { url, page: c.doc };
  const head = (c.pad ? `<title>${"x".repeat(c.pad)}</title>` : "") + (c.head ?? "").repeat(c.headRepeat ?? 1);
  return {
    url,
    page: `<!DOCTYPE html><html lang='en'><head>${head}</head><body>${(c.body ?? "").repeat(c.repeat ?? 1)}</body></html>`,
  };
}

async function scanChecks(url: string, html: string): Promise<Map<string, Check>> {
  globalThis.fetch = (async () => new Response(html)) as unknown as typeof fetch;
  const env = { SCAN_LIMITER: { limit: async () => ({ success: true }) } };
  const request = new Request("https://w.test/api/scan", { method: "POST", body: JSON.stringify({ url }) });
  const resp = await worker.fetch(request, env);
  const scan = (await resp.json()) as { seo: { checks: Check[] }; aeo: { checks: Check[] }; geo: { checks: Check[] } };
  return new Map([...scan.seo.checks, ...scan.aeo.checks, ...scan.geo.checks].map((c) => [c.id, c]));
}

const dataDir = join(import.meta.dir, "../../testdata");
const table: Record<string, { pass: CheckCase[]; fail: CheckCase[] }> = {};
for (const file of readdirSync(dataDir).filter((f) => /^check_cases.*\.json$/.test(f))) {
  Object.assign(table, JSON.parse(readFileSync(join(dataDir, file), "utf8")));
}

for (const [id, group] of Object.entries(table)) {
  test(`${id} matches the shared cases`, async () => {
    for (const [want, list] of [
      [true, group.pass],
      [false, group.fail],
    ] as const) {
      for (const [i, c] of list.entries()) {
        const { url, page } = pageOf(c);
        const check = (await scanChecks(url, page)).get(id);
        if (!check) throw new Error(`check ${id} missing`);
        expect({ id, i, pass: check.pass, page }).toEqual({ id, i, pass: want, page });
      }
    }
  });
}

test("every shared case id is a real check", async () => {
  const checks = await scanChecks("https://example.com/", "<html></html>");
  expect(Object.keys(table).filter((id) => !checks.has(id))).toEqual([]);
});

test("family counts", async () => {
  globalThis.fetch = (async () => new Response("<html></html>")) as unknown as typeof fetch;
  const env = { SCAN_LIMITER: { limit: async () => ({ success: true }) } };
  const request = new Request("https://w.test/api/scan", {
    method: "POST",
    body: JSON.stringify({ url: "https://example.com/" }),
  });
  const scan = (await (await worker.fetch(request, env)).json()) as Record<
    "seo" | "aeo" | "geo",
    { checks: unknown[] }
  >;
  expect([scan.seo.checks.length, scan.aeo.checks.length, scan.geo.checks.length]).toEqual([103, 37, 36]);
});
