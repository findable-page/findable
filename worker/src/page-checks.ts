import type { Check } from "./index";

const SEMANTIC_TAGS = ["article", "section", "main", "nav", "aside", "header", "footer", "time", "figure", "details"];

const NON_CONTENT = /<!--[\s\S]*?-->|<(script|style|noscript)\b[\s\S]*?<\/\1\s*>/gi;
const UNQUOTED = "([^\\s\"'=<>`]+)";

function stripNonContent(html: string): string {
  return html.replace(NON_CONTENT, " ");
}

function tagsOf(html: string, name: string): string[] {
  return html.match(new RegExp(`<${name}(?=[\\s/>])[^>]*>`, "gi")) ?? [];
}

function attrOf(tag: string, name: string): string | undefined {
  return tag
    .match(new RegExp(`\\s${name}\\s*=\\s*(?:"([^"]*)"|'([^']*)'|${UNQUOTED})`, "i"))
    ?.slice(1)
    .find((v) => v !== undefined);
}

const PARAGRAPH =
  /<p(?=[\s>])[^>]*>([\s\S]*?)(?=<\/p>|<p[\s>]|<\/?(?:div|section|article|ul|ol|li|h[1-6]|table|blockquote|main|header|footer)\b|$)/gi;

export function paragraphWordCounts(html: string): number[] {
  return [...html.matchAll(PARAGRAPH)].map(
    (m) =>
      (m[1] ?? "")
        .replace(/<[^>]+>/g, " ")
        .split(/\s+/)
        .filter(Boolean).length,
  );
}

function imageChecks(html: string): Check[] {
  const images = tagsOf(html, "img");
  const missingAlt = images.filter((t) => !(attrOf(t, "alt") ?? "").trim()).length;
  const missingDims = images.filter((t) => !attrOf(t, "width") || !attrOf(t, "height")).length;
  if (images.length === 0) {
    return [
      { id: "img_alt", label: "No images found on page", pass: true },
      { id: "img_dimensions", label: "No images to check for dimensions", pass: true },
    ];
  }
  return [
    { id: "img_alt", label: `${missingAlt} of ${images.length} images missing alt attributes`, pass: missingAlt === 0 },
    {
      id: "img_dimensions",
      label: `${missingDims} of ${images.length} images missing width/height`,
      pass: missingDims === 0,
    },
  ];
}

function headingsNeverSkipLevel(html: string): boolean {
  const levels = [...html.matchAll(/<h([1-6])(?=[\s/>])/gi)].map((m) => Number(m[1]));
  return levels.every((level, i) => i === 0 || level <= (levels[i - 1] ?? 0) + 1);
}

function duplicateIdCount(html: string): number {
  const counts = new Map<string, number>();
  for (const m of html.matchAll(new RegExp(`<[a-z][^>]*?\\sid\\s*=\\s*(?:"([^"]*)"|'([^']*)'|${UNQUOTED})`, "gi"))) {
    const id = (m[1] ?? m[2] ?? m[3] ?? "").trim();
    if (id) counts.set(id, (counts.get(id) ?? 0) + 1);
  }
  return [...counts.values()].filter((n) => n > 1).length;
}

function linkSafetyChecks(html: string): Check[] {
  const anchors = tagsOf(html, "a");
  const emptyHrefs = anchors.filter((t) => ["", "#"].includes((attrOf(t, "href") ?? "").trim())).length;
  const missingOpener = anchors.filter((t) => {
    if ((attrOf(t, "target") ?? "").trim().toLowerCase() !== "_blank") return false;
    return !/noopener|noreferrer/i.test(attrOf(t, "rel") ?? "");
  }).length;
  return [
    { id: "empty_hrefs", label: "No empty href attributes found", pass: emptyHrefs === 0 },
    {
      id: "link_noopener",
      label: 'External target="_blank" links include rel="noopener"',
      pass: missingOpener === 0,
    },
  ];
}

export function seoParityChecks(rawHtml: string, visibleText: string): Check[] {
  const html = stripNonContent(rawHtml);
  const encoder = new TextEncoder();
  const htmlBytes = encoder.encode(rawHtml).length;
  const ratio = htmlBytes ? (encoder.encode(visibleText).length / htmlBytes) * 100 : 0;
  const domSize = (html.match(/<[a-z][\w-]*/gi) ?? []).length;
  const duplicates = duplicateIdCount(html);
  const hasLandmarks = tagsOf(html, "main").length > 0 && tagsOf(html, "nav").length > 0;
  return [
    ...imageChecks(html),
    {
      id: "heading_hierarchy",
      label: "Heading hierarchy is correct (no skipped levels)",
      pass: headingsNeverSkipLevel(html),
    },
    {
      id: "text_html_ratio",
      label: `Text-to-HTML ratio: ${ratio.toFixed(1)}% (>=15% recommended)`,
      pass: ratio >= 15,
    },
    { id: "dom_size", label: `DOM size: ~${domSize} elements (<1500 good)`, pass: domSize < 1500 },
    ...linkSafetyChecks(html),
    {
      id: "meta_refresh",
      label: "No meta refresh redirect (good)",
      pass: !/<meta[^>]+http-equiv\s*=\s*["']?refresh/i.test(html),
    },
    { id: "aria_landmarks", label: "ARIA landmarks present", pass: hasLandmarks },
    {
      id: "duplicate_ids",
      label: `No duplicate IDs found (${duplicates} duplicates)`,
      pass: duplicates === 0,
    },
  ];
}

export function aeoParityChecks(
  rawHtml: string,
  visibleText: string,
  wordCount: number,
  hasSpeakable: boolean,
): Check[] {
  const html = stripNonContent(rawHtml);
  const avgSentenceLen = wordCount / visibleText.split(".").length;
  const semanticTotal = SEMANTIC_TAGS.reduce((sum, tag) => sum + tagsOf(html, tag).length, 0);
  return [
    { id: "aeo_speakable", label: "Speakable markup found for voice assistants", pass: hasSpeakable },
    {
      id: "aeo_reading_level",
      label: `Average sentence length: ${avgSentenceLen.toFixed(1)} words (<20 is AI-friendly)`,
      pass: avgSentenceLen <= 22,
    },
    { id: "aeo_short_sentences", label: "Concise sentence structures dominant", pass: avgSentenceLen <= 20 },
    {
      id: "aeo_semantic_html",
      label: `${semanticTotal} semantic HTML5 elements found`,
      pass: semanticTotal >= 4,
    },
  ];
}

interface GeoSignals {
  html: string;
  title: string;
  firstHeading: string | undefined;
  robotsMeta: string;
  schemaTypeCount: number;
}

function titleMatchesFirstHeading(title: string, firstHeading: string | undefined): boolean {
  if (!title || firstHeading === undefined) return false;
  const heading = firstHeading.toLowerCase();
  const matches = title
    .toLowerCase()
    .split(/\s+/)
    .filter((word) => word.length > 3 && heading.includes(word));
  return matches.length >= 2;
}

function formattingTypeCount(html: string): number {
  const present = [
    tagsOf(html, "ul").length + tagsOf(html, "ol").length > 0,
    tagsOf(html, "table").length > 0,
    tagsOf(html, "blockquote").length > 0,
    tagsOf(html, "code").length + tagsOf(html, "pre").length > 0,
    tagsOf(html, "dl").length > 0,
  ];
  return present.filter(Boolean).length;
}

export function geoParityChecks(g: GeoSignals): Check[] {
  const html = stripNonContent(g.html);
  const formats = formattingTypeCount(html);
  return [
    {
      id: "geo_passage_citability",
      label: "Passage-ready standalone answer blocks present",
      pass: paragraphWordCounts(html).some((n) => n >= 40 && n <= 120),
    },
    {
      id: "geo_entity_markup",
      label: `${g.schemaTypeCount} schema entity type(s) declared`,
      pass: g.schemaTypeCount >= 2,
    },
    {
      id: "geo_intent_clarity",
      label: "Title keywords match primary H1 heading",
      pass: titleMatchesFirstHeading(g.title, g.firstHeading),
    },
    {
      id: "geo_format_variety",
      label: `${formats} formatting types used (tables, lists, quotes, code, dl)`,
      pass: formats >= 3,
    },
    {
      id: "geo_snippet_control",
      label: "No restrictive nosnippet over-blocking",
      pass: !g.robotsMeta.includes("nosnippet"),
    },
  ];
}
