import { colorContrastCheck } from "./color-contrast";
import { linkTags, ok, type PageCtx, skip, type Tag, tagsNamed } from "./extra-checks";
import type { Check } from "./index";

const ALT_FILENAME =
  /\.(jpe?g|png|gif|webp|avif|svg|bmp)$|^(image|img|photo|picture|pic|untitled|screenshot)[-_ ]?\d*$|^(img|dsc|dscn|screenshot)[-_]\d+/i;
const IMAGE_EXT = /\.(jpe?g|png|gif|webp|avif|svg)$/i;
const MODERN_IMAGE_EXT = /\.(webp|avif|svg)$/i;
const GENERIC_ANCHORS = new Set([
  "click here",
  "here",
  "read more",
  "more",
  "learn more",
  "link",
  "this",
  "this link",
  "click",
]);
const NON_NAVIGABLE = ["#", "javascript:", "mailto:", "tel:"];
const MAX_EAGER_IMAGES = 5;
const byteLength = (text: string): number => new TextEncoder().encode(text).length;

function imageQualityChecks(c: PageCtx): Check[] {
  const images = tagsNamed(c, "img");
  if (images.length === 0) {
    return ["img_alt_quality", "img_alt_length", "img_lazy_loading", "img_modern_format"].map((id) =>
      skip(id, "No images to check"),
    );
  }
  let placeholder = 0;
  let long = 0;
  let modern = 0;
  let recognised = 0;
  let eager = 0;
  for (const img of images) {
    const alt = (img.attrs.alt ?? "").trim();
    if (alt && ALT_FILENAME.test(alt)) placeholder++;
    if (alt.length > 125) long++;
    const src = (img.attrs.src ?? "").split(/[?#]/)[0] ?? "";
    if (IMAGE_EXT.test(src)) {
      recognised++;
      if (MODERN_IMAGE_EXT.test(src)) modern++;
    }
    const lazy = (img.attrs.loading ?? "").trim().toLowerCase() === "lazy";
    const prioritised = (img.attrs.fetchpriority ?? "").trim().toLowerCase() === "high";
    if (!lazy && !prioritised) eager++;
  }
  return [
    ok("img_alt_quality", `${placeholder} image alt text(s) look like filenames or placeholders`, placeholder === 0),
    ok("img_alt_length", `${long} image alt text(s) longer than 125 characters`, long === 0),
    ok("img_lazy_loading", `${eager} image(s) load eagerly (limit ${MAX_EAGER_IMAGES})`, eager <= MAX_EAGER_IMAGES),
    ok(
      "img_modern_format",
      `${modern} of ${recognised} images use WebP, AVIF or SVG`,
      recognised === 0 || modern * 2 >= recognised,
    ),
  ];
}

export function resolveNavigable(c: PageCtx, href: string): URL | undefined {
  const trimmed = href.trim();
  if (!trimmed || NON_NAVIGABLE.some((prefix) => trimmed.startsWith(prefix))) return undefined;
  try {
    return new URL(trimmed, c.url);
  } catch {
    return undefined;
  }
}

function linkQualityChecks(c: PageCtx): Check[] {
  let nofollowInternal = 0;
  let generic = 0;
  let empty = 0;
  let withHref = 0;
  let javascript = 0;
  for (const a of tagsNamed(c, "a")) {
    const href = (a.attrs.href ?? "").trim();
    if (!href) continue;
    withHref++;
    if (href.toLowerCase().startsWith("javascript:")) javascript++;
    const target = resolveNavigable(c, href);
    if (target && target.host === c.url.host && /nofollow/i.test(a.attrs.rel ?? "")) nofollowInternal++;
    const text = a.text.trim().replace(/\s+/g, " ").toLowerCase();
    if (GENERIC_ANCHORS.has(text)) generic++;
    const unlabeled = !(a.attrs["aria-label"] ?? "").trim() && !(a.attrs.title ?? "").trim();
    if (href !== "#" && !text && !a.media && unlabeled) empty++;
  }
  const gated = (id: string, label: string, pass: boolean): Check =>
    withHref === 0 ? skip(id, "No links on the page") : ok(id, label, pass);
  return [
    gated("nofollow_internal", `${nofollowInternal} internal link(s) marked rel=nofollow`, nofollowInternal === 0),
    gated("generic_anchor_text", `${generic} link(s) use generic anchor text like "click here"`, generic === 0),
    gated("empty_anchor_text", `${empty} link(s) have no text, label or image`, empty === 0),
    gated("javascript_hrefs", `${javascript} link(s) use javascript: hrefs`, javascript === 0),
    gated("links_count", `${withHref} links on page (200 or fewer recommended)`, withHref <= 200),
    mixedContentCheck(c),
  ];
}

function mixedContentCheck(c: PageCtx): Check {
  if (c.url.protocol !== "https:") {
    return skip("mixed_content", "Page is not served over HTTPS, no mixed content to check");
  }
  const insecure = (tags: Tag[], attr: string) =>
    tags.filter((t) => (t.attrs[attr] ?? "").trim().toLowerCase().startsWith("http://")).length;
  const count =
    ["img", "script", "iframe", "video", "audio", "source"].reduce(
      (sum, name) => sum + insecure(tagsNamed(c, name), "src"),
      0,
    ) + insecure(linkTags(c, "stylesheet"), "href");
  return ok("mixed_content", `${count} insecure http:// resource(s) on an HTTPS page`, count === 0);
}

function hasResourceHints(c: PageCtx, stylesheets: Tag[]): boolean {
  const hinted = ["preconnect", "preload", "dns-prefetch", "prefetch", "modulepreload"];
  if (hinted.some((rel) => linkTags(c, rel).length > 0)) return true;
  const resources = [
    ...tagsNamed(c, "script").map((t) => t.attrs.src ?? ""),
    ...stylesheets.map((t) => t.attrs.href ?? ""),
  ];
  return !resources.some((ref) => {
    const target = resolveNavigable(c, ref);
    return target !== undefined && target.host !== c.url.host;
  });
}

function loadPerformanceChecks(c: PageCtx): Check[] {
  const scripts = tagsNamed(c, "script").filter((t) => (t.attrs.src ?? "").trim() !== "");
  const blocking = scripts.filter(
    (t) =>
      t.inHead &&
      !("async" in t.attrs) &&
      !("defer" in t.attrs) &&
      (t.attrs.type ?? "").trim().toLowerCase() !== "module",
  ).length;
  const stylesheets = linkTags(c, "stylesheet");
  const cssBytes = tagsNamed(c, "style").reduce((sum, t) => sum + byteLength(t.text), 0);
  const htmlBytes = byteLength(c.raw);
  return [
    scripts.length === 0
      ? skip("render_blocking_scripts", "No external scripts")
      : ok("render_blocking_scripts", `${blocking} render-blocking script(s) in head`, blocking === 0),
    scripts.length === 0
      ? skip("script_count", "No external scripts")
      : ok("script_count", `${scripts.length} external scripts (20 or fewer recommended)`, scripts.length <= 20),
    stylesheets.length === 0
      ? skip("stylesheet_count", "No stylesheets")
      : ok("stylesheet_count", `${stylesheets.length} stylesheets (6 or fewer recommended)`, stylesheets.length <= 6),
    ok("inline_css_size", `${cssBytes} bytes of inline CSS (20000 or fewer recommended)`, cssBytes <= 20000),
    ok("resource_hints", "Preconnect or preload hints cover third-party resources", hasResourceHints(c, stylesheets)),
    ok("html_weight", `HTML is ${htmlBytes} bytes (500000 or fewer recommended)`, htmlBytes <= 500000),
  ];
}

function accessibilityChecks(c: PageCtx): Check[] {
  const hasSkipLink = tagsNamed(c, "a").some((a) => a.text.toLowerCase().includes("skip to"));
  const buttons = tagsNamed(c, "button");
  const unnamed = buttons.filter(
    (b) => !b.text.trim() && !b.media && !(b.attrs["aria-label"] ?? "").trim() && !(b.attrs.title ?? "").trim(),
  ).length;
  const headers = tagsNamed(c, "th").length;
  const tables = tagsNamed(c, "table").length;
  const iframes = tagsNamed(c, "iframe");
  const untitled = iframes.filter((f) => !(f.attrs.title ?? "").trim()).length;
  return [
    ok("skip_link", "Skip-to-content link present", hasSkipLink),
    buttons.length === 0
      ? skip("button_names", "No buttons")
      : ok("button_names", `${unnamed} button(s) have no accessible name`, unnamed === 0),
    tables === 0
      ? skip("table_headers", "No tables")
      : ok("table_headers", `${headers} header cell(s) for ${tables} table(s)`, headers >= tables),
    iframes.length === 0
      ? skip("iframe_title", "No iframes")
      : ok("iframe_title", `${untitled} iframe(s) missing a title attribute`, untitled === 0),
    colorContrastCheck(c),
  ];
}

export function seoPageChecks(c: PageCtx): Check[] {
  return [...imageQualityChecks(c), ...linkQualityChecks(c), ...loadPerformanceChecks(c), ...accessibilityChecks(c)];
}
