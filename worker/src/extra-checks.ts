import type { Check } from "./index";

export interface Tag {
  attrs: Record<string, string>;
  text: string;
  inHead: boolean;
}

export interface PageCtx {
  raw: string;
  url: URL;
  text: string;
  tags: Map<string, Tag[]>;
}

const TEXT_TAGS = new Set(["a", "button", "style", "label"]);
const RAW_TEXT_TAGS = new Set(["script", "style", "noscript"]);
const TAG = /<(\/?)([a-zA-Z][\w:-]*)((?:[^>"']|"[^"]*"|'[^']*')*)>/g;
const ATTR = /([^\s"'<>/=]+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>`]+)))?/g;

function parseAttrs(source: string): Record<string, string> {
  const attrs: Record<string, string> = {};
  for (const m of source.matchAll(ATTR)) attrs[(m[1] ?? "").toLowerCase()] = m[2] ?? m[3] ?? m[4] ?? "";
  return attrs;
}

export function parseTags(raw: string): Map<string, Tag[]> {
  const html = raw.replace(/<!--[\s\S]*?-->/g, " ");
  const lower = html.toLowerCase();
  const tags = new Map<string, Tag[]>();
  let inHead = false;
  TAG.lastIndex = 0;
  for (let m = TAG.exec(html); m; m = TAG.exec(html)) {
    const name = (m[2] ?? "").toLowerCase();
    if (m[1]) {
      if (name === "head") inHead = false;
      continue;
    }
    const bodyStart = TAG.lastIndex;
    let text = "";
    if (RAW_TEXT_TAGS.has(name) || TEXT_TAGS.has(name)) {
      const close = lower.indexOf(`</${name}`, bodyStart);
      const end = close === -1 ? html.length : close;
      text = html.slice(bodyStart, end);
      if (RAW_TEXT_TAGS.has(name)) TAG.lastIndex = end;
      else text = text.replace(/<[^>]+>/g, "");
    }
    if (name === "head") inHead = true;
    const list = tags.get(name) ?? [];
    list.push({ attrs: parseAttrs(m[3] ?? ""), text: name === "script" || name === "noscript" ? "" : text, inHead });
    tags.set(name, list);
  }
  return tags;
}

export function buildCtx(raw: string, url: URL, text: string): PageCtx {
  return { raw, url, text, tags: parseTags(raw) };
}

function tagsNamed(c: PageCtx, name: string): Tag[] {
  return c.tags.get(name) ?? [];
}

function linkTags(c: PageCtx, rel: string): Tag[] {
  return tagsNamed(c, "link").filter((t) => (t.attrs.rel ?? "").toLowerCase().split(/\s+/).includes(rel));
}

function metaTags(c: PageCtx, name: string): string[] {
  return tagsNamed(c, "meta")
    .filter((t) => (t.attrs.name ?? "").trim().toLowerCase() === name)
    .map((t) => (t.attrs.content ?? "").trim());
}

function firstHref(tags: Tag[]): string {
  return tags.map((t) => (t.attrs.href ?? "").trim()).find(Boolean) ?? "";
}

const ok = (id: string, label: string, pass: boolean): Check => ({ id, label, pass });

const LANG_TAG = /^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$/;
const HREFLANG = /^(x-default|[a-z]{2,3}(-[a-z]{4})?(-([a-z]{2}|\d{3}))?)$/i;
const CHARSET_TAG = /<meta[^>]+charset|http-equiv=["']?content-type/i;

function pagePathKey(u: URL): string {
  return u.host.toLowerCase() + u.pathname.replace(/\/$/, "");
}

function canonicalQuality(c: PageCtx): Check[] {
  const canonical = firstHref(linkTags(c, "canonical"));
  if (!canonical) {
    return [
      ok("canonical_self", "No canonical to compare with this page", true),
      ok("canonical_absolute", "No canonical to validate", true),
    ];
  }
  let self = false;
  try {
    self = pagePathKey(new URL(canonical, c.url)) === pagePathKey(c.url);
  } catch {}
  return [
    ok("canonical_self", self ? "Canonical points to this page" : `Canonical points elsewhere: ${canonical}`, self),
    ok("canonical_absolute", "Canonical is an absolute URL", /^[a-z][a-z0-9+.-]*:/i.test(canonical)),
  ];
}

function robotsChecks(c: PageCtx): Check[] {
  const tokens = new Set(
    metaTags(c, "robots")
      .flatMap((content) => content.toLowerCase().split(","))
      .map((t) => t.trim()),
  );
  const conflict = (tokens.has("index") && tokens.has("noindex")) || (tokens.has("follow") && tokens.has("nofollow"));
  return [
    ok("meta_robots_conflict", "Robots meta directives do not contradict each other", !conflict),
    ok(
      "robots_nofollow",
      "Page does not nofollow all its links via robots meta",
      !tokens.has("nofollow") && !tokens.has("none"),
    ),
  ];
}

function singletonChecks(c: PageCtx): Check[] {
  const titles = tagsNamed(c, "title").filter((t) => t.inHead).length;
  const descriptions = metaTags(c, "description").length;
  return [
    ok("title_single", `${titles} title tag(s) in head (exactly 1 expected)`, titles <= 1),
    ok("desc_single", `${descriptions} meta description tag(s) (exactly 1 expected)`, descriptions <= 1),
  ];
}

function urlShapeChecks(c: PageCtx): Check[] {
  const requestUri = c.url.pathname + c.url.search;
  const params = c.url.search.slice(1).split("&").filter(Boolean).length;
  const depth = c.url.pathname.split("/").filter(Boolean).length;
  return [
    ok(
      "url_length",
      `URL path and query are ${requestUri.length} characters (100 or fewer recommended)`,
      requestUri.length <= 100,
    ),
    ok("url_query_params", `${params} query parameter(s) in URL (2 or fewer recommended)`, params <= 2),
    ok("url_depth", `URL is ${depth} level(s) deep (5 or fewer recommended)`, depth <= 5),
  ];
}

function brandAssetChecks(c: PageCtx): Check[] {
  const themeColors = metaTags(c, "theme-color");
  const feed = linkTags(c, "alternate").some((t) => /rss|atom/i.test(t.attrs.type ?? ""));
  return [
    ok("favicon", "Favicon link found", firstHref(linkTags(c, "icon")) !== ""),
    ok("apple_touch_icon", "Apple touch icon declared", firstHref(linkTags(c, "apple-touch-icon")) !== ""),
    ok("theme_color", "theme-color meta tag found", themeColors.length > 0 && themeColors[0] !== ""),
    ok("web_manifest", "Web app manifest linked", firstHref(linkTags(c, "manifest")) !== ""),
    ok("rss_feed", "RSS or Atom feed advertised", feed),
  ];
}

function blocksZoom(viewport: string): boolean {
  if (viewport.includes("user-scalable=no") || viewport.includes("user-scalable=0")) return true;
  const scale = viewport.match(/maximum-scale=(\d+(?:\.\d+)?)/)?.[1];
  return scale !== undefined && Number(scale) < 2;
}

function viewportAndLangChecks(c: PageCtx): Check[] {
  const viewport = (metaTags(c, "viewport")[0] ?? "").toLowerCase().replaceAll(" ", "");
  const lang = (tagsNamed(c, "html")[0]?.attrs.lang ?? "").trim();
  return [
    viewport
      ? ok("viewport_device_width", "Viewport sets width=device-width", viewport.includes("width=device-width"))
      : ok("viewport_device_width", "No viewport to validate", true),
    viewport
      ? ok("viewport_scalable", "Viewport does not block pinch zoom", !blocksZoom(viewport))
      : ok("viewport_scalable", "No viewport to validate", true),
    lang
      ? ok("lang_valid", `HTML lang "${lang}" is a valid language tag`, LANG_TAG.test(lang))
      : ok("lang_valid", "No lang attribute to validate", true),
  ];
}

function hreflangChecks(c: PageCtx): Check[] {
  const alternates = linkTags(c, "alternate").filter((t) => (t.attrs.hreflang ?? "").trim() !== "");
  if (alternates.length === 0) {
    return [
      ok("hreflang_valid", "No hreflang annotations to validate", true),
      ok("hreflang_absolute", "No hreflang annotations to validate", true),
    ];
  }
  const invalid = alternates.filter((t) => !HREFLANG.test((t.attrs.hreflang ?? "").trim())).length;
  const relative = alternates.filter((t) => !/^https?:\/\//i.test((t.attrs.href ?? "").trim())).length;
  return [
    ok("hreflang_valid", `${invalid} of ${alternates.length} hreflang codes are invalid`, invalid === 0),
    ok("hreflang_absolute", `${relative} of ${alternates.length} hreflang URLs are not absolute`, relative === 0),
  ];
}

function charsetPlacementCheck(c: PageCtx): Check {
  if (!CHARSET_TAG.test(c.raw)) return ok("charset_early", "No charset declaration to place", true);
  return ok("charset_early", "Charset declared within the first 1024 bytes", CHARSET_TAG.test(c.raw.slice(0, 1024)));
}

export function seoExtraChecks(c: PageCtx): Check[] {
  return [
    ...canonicalQuality(c),
    ...robotsChecks(c),
    ...singletonChecks(c),
    ...urlShapeChecks(c),
    ...brandAssetChecks(c),
    ...viewportAndLangChecks(c),
    ...hreflangChecks(c),
    charsetPlacementCheck(c),
  ];
}
