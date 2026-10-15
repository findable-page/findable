import { expect, test } from "bun:test";
import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import worker, { analyzeHtml } from "../src/index";

const root = join(import.meta.dir, "../..");
const LANGS = ["de", "es", "tr", "ru"] as const;
type CheckText = Record<string, { name: string; fix: string }>;

const texts = Object.fromEntries(
  LANGS.map((l) => [
    l,
    JSON.parse(readFileSync(join(root, `internal/server/web/i18n/${l}.json`), "utf8")) as CheckText,
  ]),
) as Record<(typeof LANGS)[number], CheckText>;

function engineCheckIds(): Set<string> {
  const ids = new Set<string>();
  const run = (html: string) => {
    const scan = analyzeHtml("https://example.com/page", html, new Headers());
    for (const family of [scan.seo, scan.aeo, scan.geo]) for (const c of family.checks) ids.add(c.id);
  };
  run("<html></html>");
  const dir = join(root, "testdata");
  for (const file of readdirSync(dir).filter((f) => /^check_cases.*\.json$/.test(f))) {
    const table = JSON.parse(readFileSync(join(dir, file), "utf8")) as Record<
      string,
      {
        pass: { doc?: string; head?: string; body?: string; repeat?: number }[];
        fail: { doc?: string; head?: string; body?: string; repeat?: number }[];
      }
    >;
    for (const group of Object.values(table)) {
      for (const c of [...group.pass, ...group.fail]) {
        run(
          c.doc ??
            `<!DOCTYPE html><html lang='en'><head>${c.head ?? ""}</head><body>${(c.body ?? "").repeat(c.repeat ?? 1)}</body></html>`,
        );
      }
    }
  }
  return ids;
}

const placeholders = (s: string) => [...s.matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort();

test("every check id the engine produces has a name and fix in all four languages", () => {
  const ids = engineCheckIds();
  expect(ids.size).toBe(159);
  for (const lang of LANGS) {
    const missing = [...ids].filter((id) => !texts[lang][id]?.name?.trim() || !texts[lang][id]?.fix?.trim());
    expect({ lang, missing }).toEqual({ lang, missing: [] });
  }
});

test("language maps hold no ids the engine does not produce", () => {
  const ids = engineCheckIds();
  for (const lang of LANGS) {
    expect({ lang, extra: Object.keys(texts[lang]).filter((id) => !ids.has(id)) }).toEqual({ lang, extra: [] });
  }
});

test("placeholders agree across languages", () => {
  for (const id of Object.keys(texts.de)) {
    const want = placeholders(`${texts.de[id]?.name} ${texts.de[id]?.fix}`);
    for (const lang of LANGS) {
      const got = placeholders(`${texts[lang][id]?.name} ${texts[lang][id]?.fix}`);
      expect({ id, lang, got }).toEqual({ id, lang, got: want });
    }
  }
});

test("the worker serves each language map and 404s on unknown ones", async () => {
  for (const lang of LANGS) {
    const resp = await worker.fetch(new Request(`https://w.test/i18n/${lang}.json`));
    expect(resp.status).toBe(200);
    expect(await resp.json()).toEqual(texts[lang]);
  }
  for (const bad of ["en", "constructor", "xx"]) {
    expect((await worker.fetch(new Request(`https://w.test/i18n/${bad}.json`))).status).not.toBe(200);
  }
});
