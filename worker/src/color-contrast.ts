import { ok, type PageCtx, skip, tagsNamed } from "./extra-checks";
import type { Check } from "./index";

type Rgb = [number, number, number];

const MIN_CONTRAST_RATIO = 4.5;
const MIN_LARGE_TEXT_RATIO = 3;
const LARGE_TEXT_TAGS = new Set(["h1", "h2", "h3"]);
const CONTRAST_TAGS = ["p", "a", "li", "button", "label", "h1", "h2", "h3", "h4", "h5", "h6"];
const NAMED_ENTITIES: Record<string, string> = {
  amp: "&",
  quot: '"',
  apos: "'",
  lt: "<",
  gt: ">",
  colon: ":",
  semi: ";",
  num: "#",
  lpar: "(",
  rpar: ")",
  comma: ",",
};
const ENTITY = /&(?:#(\d{1,7})|#[xX]([0-9a-fA-F]{1,6})|([a-zA-Z]+));?/g;
const trimAscii = (s: string): string => s.replace(/^[ \t\n\f\r]+|[ \t\n\f\r]+$/g, "");

function decodeEntities(s: string): string {
  return s.replace(ENTITY, (whole, dec, hex, name) => {
    if (name) return NAMED_ENTITIES[name.toLowerCase()] ?? whole;
    const code = dec ? Number(dec) : Number.parseInt(hex, 16);
    return code > 0 && code <= 0x10ffff ? String.fromCodePoint(code) : whole;
  });
}
const NAMED_COLORS: Record<string, Rgb> = { white: [255, 255, 255], black: [0, 0, 0] };
const HEX_COLOR = /^#([0-9a-f]{3}|[0-9a-f]{6})$/;
const RGB_COLOR =
  /^rgb\([ \t\n\f\r]*(\d{1,3})[ \t\n\f\r]*,[ \t\n\f\r]*(\d{1,3})[ \t\n\f\r]*,[ \t\n\f\r]*(\d{1,3})[ \t\n\f\r]*\)$/;

function parseCssColor(value: string): Rgb | null {
  const named = NAMED_COLORS[value];
  if (named) return named;
  const hex = HEX_COLOR.exec(value)?.[1];
  if (hex) {
    const full = hex.length === 3 ? [...hex].map((c) => c + c).join("") : hex;
    return [0, 2, 4].map((i) => Number.parseInt(full.slice(i, i + 2), 16)) as Rgb;
  }
  const rgb = RGB_COLOR.exec(value);
  if (rgb) {
    const parts = [rgb[1], rgb[2], rgb[3]].map(Number);
    if (parts.every((n) => n <= 255)) return parts as Rgb;
  }
  return null;
}

function inlineColorPair(style: string): [Rgb, Rgb] | null {
  let fg: Rgb | null = null;
  let bg: Rgb | null = null;
  for (const decl of decodeEntities(style).split(";")) {
    const colon = decl.indexOf(":");
    if (colon === -1) continue;
    const prop = trimAscii(decl.slice(0, colon)).toLowerCase();
    const color = parseCssColor(trimAscii(decl.slice(colon + 1)).toLowerCase());
    if (prop === "color") fg = color;
    else if (prop === "background-color") bg = color;
  }
  return fg && bg ? [fg, bg] : null;
}

function luminance(color: Rgb): number {
  const [r, g, b] = color.map((v) => {
    const c = v / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  }) as Rgb;
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrastRatio(a: Rgb, b: Rgb): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x) as [number, number];
  return (hi + 0.05) / (lo + 0.05);
}

export function colorContrastCheck(c: PageCtx): Check {
  let checked = 0;
  let low = 0;
  for (const name of CONTRAST_TAGS) {
    for (const tag of tagsNamed(c, name)) {
      if (!trimAscii(tag.text)) continue;
      const pair = inlineColorPair(tag.attrs.style ?? "");
      if (!pair) continue;
      checked++;
      const min = LARGE_TEXT_TAGS.has(name) ? MIN_LARGE_TEXT_RATIO : MIN_CONTRAST_RATIO;
      if (contrastRatio(pair[0], pair[1]) < min) low++;
    }
  }
  return checked === 0
    ? skip("color_contrast", "No inline color pairs to check")
    : ok(
        "color_contrast",
        `${low} of ${checked} inline-colored text element(s) fall below WCAG AA contrast (4.5:1, 3:1 for h1-h3)`,
        low === 0,
      );
}
