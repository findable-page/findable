import type { Check } from "./index";

export interface Tag {
  attrs: Record<string, string>;
  text: string;
  media: boolean;
  inHead: boolean;
  pos: number;
}

export interface PageCtx {
  raw: string;
  url: URL;
  text: string;
  tags: Map<string, Tag[]>;
}

const TEXT_TAGS = new Set(["a", "button", "label", "h1", "h2", "h3", "h4", "h5", "h6", "p", "li"]);
const RAW_TEXT_TAGS = new Set(["script", "style", "noscript"]);
const P_ENDERS = new Set([
  "p",
  "li",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "div",
  "ul",
  "ol",
  "table",
  "section",
  "article",
  "blockquote",
  "dl",
  "pre",
  "form",
]);
const PARENT_CLOSERS = new Set(["ul", "ol", "table", "body", "html"]);
const TEXT_CAP = 8192;
const JSON_LD_CAP = 1 << 20;
const ATTR = /([^\s"'<>/=]+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>`]+)))?/g;
const NAME_CHAR = /[\w:-]/;
const isLetter = (code: number): boolean => (code >= 65 && code <= 90) || (code >= 97 && code <= 122);

function parseAttrs(source: string): Record<string, string> {
  const attrs: Record<string, string> = {};
  for (const m of source.matchAll(ATTR)) attrs[(m[1] ?? "").toLowerCase()] = m[2] ?? m[3] ?? m[4] ?? "";
  return attrs;
}

function stripTags(text: string): string {
  let out = "";
  let from = 0;
  for (;;) {
    const open = text.indexOf("<", from);
    if (open === -1) return out + text.slice(from);
    const close = text.indexOf(">", open);
    if (close === -1) return out + text.slice(from);
    out += text.slice(from, open);
    from = close + 1;
  }
}

function tagEnd(html: string, from: number): number {
  let i = from;
  let previous = "";
  while (i < html.length) {
    const ch = html[i] as string;
    if (ch === ">") return i;
    if ((ch === '"' || ch === "'") && previous === "=") {
      const closing = html.indexOf(ch, i + 1);
      if (closing !== -1) i = closing;
    }
    if (ch !== " " && ch !== "\t" && ch !== "\n" && ch !== "\r") previous = ch;
    i++;
  }
  return -1;
}

interface Pending {
  tag: Tag;
  start: number;
}

export function parseTags(html: string): Map<string, Tag[]> {
  const lower = html.toLowerCase();
  const tags = new Map<string, Tag[]>();
  const pending = new Map<string, Pending>();
  const rawClose = new Map<string, number>();
  let inHead = false;
  let pos = 0;

  const finish = (name: string, end: number) => {
    const open = pending.get(name);
    if (!open) return;
    pending.delete(name);
    const slice = html.slice(open.start, Math.min(end, open.start + TEXT_CAP));
    open.tag.media = /<(img|svg|picture)(?=[\s/>])/i.test(slice);
    open.tag.text = stripTags(slice);
  };
  const finishAll = (end: number) => {
    for (const name of [...pending.keys()]) finish(name, end);
  };
  const closeIndex = (name: string, from: number): number => {
    const cached = rawClose.get(name);
    if (cached !== undefined && (cached === -1 || cached >= from)) return cached;
    const found = lower.indexOf(`</${name}`, from);
    rawClose.set(name, found);
    return found;
  };

  let i = 0;
  for (;;) {
    const lt = html.indexOf("<", i);
    if (lt === -1) break;
    if (html.startsWith("<!--", lt)) {
      const commentEnd = html.indexOf("-->", lt + 4);
      if (commentEnd === -1) break;
      i = commentEnd + 3;
      continue;
    }
    const closing = html[lt + 1] === "/";
    const nameStart = lt + (closing ? 2 : 1);
    if (!isLetter(html.charCodeAt(nameStart))) {
      i = lt + 1;
      continue;
    }
    let nameEnd = nameStart + 1;
    while (nameEnd < html.length && NAME_CHAR.test(html[nameEnd] ?? "")) nameEnd++;
    const end = tagEnd(html, nameEnd);
    if (end === -1) break;
    const name = html.slice(nameStart, nameEnd).toLowerCase();
    i = end + 1;

    if (closing) {
      if (name === "head") inHead = false;
      if (PARENT_CLOSERS.has(name)) finishAll(lt);
      else finish(name, lt);
      continue;
    }

    if (name === "body") inHead = false;
    if (name === "head") inHead = true;
    if (P_ENDERS.has(name)) finish("p", lt);
    if (TEXT_TAGS.has(name)) finish(name, lt);

    const attrs = parseAttrs(html.slice(nameEnd, end));
    const tag: Tag = { pos: pos++, attrs, text: "", media: false, inHead };
    const list = tags.get(name) ?? [];
    list.push(tag);
    tags.set(name, list);

    if (TEXT_TAGS.has(name)) pending.set(name, { tag, start: end + 1 });
    else if (RAW_TEXT_TAGS.has(name)) {
      const close = closeIndex(name, end + 1);
      const stop = close === -1 ? html.length : close;
      const isJsonLd = name === "script" && (attrs.type ?? "").trim().toLowerCase() === "application/ld+json";
      if (isJsonLd) tag.text = html.slice(end + 1, Math.min(stop, end + 1 + JSON_LD_CAP));
      else if (name === "style") tag.text = html.slice(end + 1, stop);
      i = stop;
    }
  }
  finishAll(html.length);
  return tags;
}

export function buildCtx(raw: string, url: URL, text: string): PageCtx {
  return { raw, url, text, tags: parseTags(raw) };
}

export function tagsNamed(c: PageCtx, name: string): Tag[] {
  return c.tags.get(name) ?? [];
}

export function linkTags(c: PageCtx, rel: string): Tag[] {
  return tagsNamed(c, "link").filter((t) => (t.attrs.rel ?? "").toLowerCase().split(/\s+/).includes(rel));
}

export function metaTags(c: PageCtx, name: string): string[] {
  return tagsNamed(c, "meta")
    .filter((t) => (t.attrs.name ?? "").trim().toLowerCase() === name)
    .map((t) => (t.attrs.content ?? "").trim());
}

export function metaProperty(c: PageCtx, prop: string): string {
  for (const t of tagsNamed(c, "meta")) {
    for (const key of ["property", "name"]) {
      if ((t.attrs[key] ?? "").trim().toLowerCase() === prop && (t.attrs.content ?? "").trim()) {
        return (t.attrs.content ?? "").trim();
      }
    }
  }
  return "";
}

function firstHref(tags: Tag[]): string {
  return tags.map((t) => (t.attrs.href ?? "").trim()).find(Boolean) ?? "";
}

export const ok = (id: string, label: string, pass: boolean): Check => ({ id, label, pass });
export const skip = (id: string, label: string): Check => ({ id, label, pass: true, applicable: false });

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
      skip("canonical_self", "No canonical to compare with this page"),
      skip("canonical_absolute", "No canonical to validate"),
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
  return [ok("favicon", "Favicon link found", firstHref(linkTags(c, "icon")) !== "")];
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
      : skip("viewport_device_width", "No viewport to validate"),
    viewport
      ? ok("viewport_scalable", "Viewport does not block pinch zoom", !blocksZoom(viewport))
      : skip("viewport_scalable", "No viewport to validate"),
    lang
      ? ok("lang_valid", `HTML lang "${lang.slice(0, 35)}" is a valid language tag`, LANG_TAG.test(lang))
      : skip("lang_valid", "No lang attribute to validate"),
  ];
}

function hreflangChecks(c: PageCtx): Check[] {
  const alternates = linkTags(c, "alternate").filter((t) => (t.attrs.hreflang ?? "").trim() !== "");
  if (alternates.length === 0) {
    return [
      skip("hreflang_valid", "No hreflang annotations to validate"),
      skip("hreflang_absolute", "No hreflang annotations to validate"),
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
  if (!CHARSET_TAG.test(c.raw)) return skip("charset_early", "No charset declaration to place");
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
