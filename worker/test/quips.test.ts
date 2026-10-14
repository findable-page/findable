import { expect, test } from "bun:test";
import { cardTheme, QUIPS, scoreQuip, shareDescription } from "../src/quips";

test("the same site always gets the same line", () => {
  expect(scoreQuip(62, "example.com")).toBe(scoreQuip(62, "example.com"));
});

test("sites with the same score mostly get different lines", () => {
  const lines = new Set(Array.from({ length: 40 }, (_, i) => scoreQuip(62, `site${i}.com`)));
  expect(lines.size).toBeGreaterThanOrEqual(6);
});

test("each line comes from its score band", () => {
  for (const [score, band] of [
    [95, 0],
    [75, 1],
    [55, 2],
    [35, 3],
    [10, 4],
  ] as const) {
    expect(QUIPS[band]).toContain(scoreQuip(score, "example.com"));
  }
});

test("every line fits a share title next to a long host", () => {
  for (const line of QUIPS.flat()) expect(line.length).toBeLessThanOrEqual(48);
});

test("descriptions vary by site and always carry the three scores", () => {
  const scores = { seo: 71, aeo: 55, geo: 43 };
  const texts = new Set(Array.from({ length: 40 }, (_, i) => shareDescription(scores, `site${i}.com`)));
  expect(texts.size).toBeGreaterThanOrEqual(5);
  for (const t of texts) expect(t.startsWith("SEO 71 · AEO 55 · GEO 43. ")).toBe(true);
  expect(shareDescription(scores, "a.com")).toBe(shareDescription(scores, "a.com"));
});

test("the card theme varies by site", () => {
  const themes = new Set(Array.from({ length: 40 }, (_, i) => cardTheme(`site${i}.com`).background));
  expect(themes.size).toBeGreaterThanOrEqual(4);
});
