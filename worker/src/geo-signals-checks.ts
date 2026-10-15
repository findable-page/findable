import { metaProperty, metaTags, ok, type PageCtx, skip, tagsNamed } from "./extra-checks";
import type { Check } from "./index";
import { ISO_DATE, parseSchemaDoc, present, type SchemaDoc } from "./schema-checks";
import { resolveNavigable } from "./seo-page-checks";

const BYLINE_CLASS = /\b(byline|author)\b/;
const VISIBLE_DATE =
  /(updated|published|modified|posted|reviewed)\s*(on|:)?\s*(([a-z]{3,9}\.? \d{1,2}(st|nd|rd|th)?,? \d{4})|\d{4}-\d{2}-\d{2}|\d{1,2} [a-z]{3,9}\.? \d{4})/;
const EXPERIENCE =
  /\b(we tested|i tested|we found|i found|we measured|our (research|tests?|study|experience|data|analysis)|in my experience|hands-on|case study)\b/;
const EXPERTISE =
  /\b(ph\.?d|m\.?d|mba|cpa|certified|licensed|board[- ]certified|professor|years of experience|years experience)\b|\bdr\.?\s/;
const REVIEWED = /\b(reviewed by|fact[- ]checked|edited by|medically reviewed|peer[- ]reviewed)\b/;
const REFERENCES = /^(sources?|references?|citations?|further reading|bibliography|works cited)\b/;
const MAX_SNIPPET = /max-snippet:\s*(-?\d+)/g;
const HEADINGS = ["h1", "h2", "h3", "h4", "h5", "h6"];

const AUTHORITATIVE_SUFFIXES = [".gov", ".edu", ".mil", ".ac.uk", ".gov.uk", ".gc.ca"];
const AUTHORITATIVE_DOMAINS = [
  "wikipedia.org",
  "who.int",
  "un.org",
  "europa.eu",
  "doi.org",
  "arxiv.org",
  "nature.com",
  "science.org",
  "sciencedirect.com",
  "jstor.org",
  "springer.com",
  "ieee.org",
  "acm.org",
];

const lowerAttr = (value: string | undefined): string => (value ?? "").trim().toLowerCase();
const withoutWww = (host: string): string => host.toLowerCase().replace(/^www\./, "");

function schemaDate(doc: SchemaDoc, key: string): string {
  for (const node of doc.nodes) {
    const value = node[key];
    if (typeof value === "string" && ISO_DATE.test(value.trim())) return value.trim();
  }
  return "";
}

function publishingDateChecks(c: PageCtx): Check[] {
  const doc = parseSchemaDoc(c);
  const published = schemaDate(doc, "datePublished") || metaProperty(c, "article:published_time");
  const modified = schemaDate(doc, "dateModified") || metaProperty(c, "article:modified_time");
  const comparable = published.length >= 10 && modified.length >= 10;
  return [
    ok("geo_date_published", "Publish date declared in schema or article:published_time", published !== ""),
    ok("geo_date_modified", "Modified date declared in schema or article:modified_time", modified !== ""),
    comparable
      ? ok(
          "geo_modified_after_published",
          "Modified date is not earlier than the publish date",
          modified.slice(0, 10) >= published.slice(0, 10),
        )
      : skip("geo_modified_after_published", "Needs both a publish and a modified date"),
    ok(
      "geo_visible_date",
      "A visible published or updated date appears on the page",
      VISIBLE_DATE.test(c.text.toLowerCase()),
    ),
  ];
}

function externalHosts(c: PageCtx): string[] {
  const own = withoutWww(c.url.host);
  const hosts = new Set<string>();
  for (const a of tagsNamed(c, "a")) {
    const target = resolveNavigable(c, a.attrs.href ?? "");
    if (!target || (target.protocol !== "http:" && target.protocol !== "https:")) continue;
    const host = withoutWww(target.host);
    if (host && host !== own) hosts.add(host);
  }
  return [...hosts];
}

const isAuthoritativeHost = (host: string): boolean =>
  AUTHORITATIVE_SUFFIXES.some((suffix) => host.endsWith(suffix)) ||
  AUTHORITATIVE_DOMAINS.some((domain) => host === domain || host.endsWith(`.${domain}`));

function sourceQualityChecks(c: PageCtx): Check[] {
  const hosts = externalHosts(c);
  const authoritative = hosts.filter(isAuthoritativeHost).length;
  const citeTags = tagsNamed(c, "cite").length;
  const citedQuotes = ["blockquote", "q"].reduce(
    (sum, name) => sum + tagsNamed(c, name).filter((t) => (t.attrs.cite ?? "").trim() !== "").length,
    0,
  );
  const quotes = tagsNamed(c, "blockquote").length;
  const references = HEADINGS.some((name) =>
    tagsNamed(c, name).some((t) => REFERENCES.test(t.text.split(/\s+/).filter(Boolean).join(" ").toLowerCase())),
  );
  const attributionLabel = `${citeTags + citedQuotes} attributed source(s) for ${quotes} blockquote(s)`;
  return [
    ok(
      "geo_authoritative_links",
      `${authoritative} link(s) to government, academic or reference sources`,
      authoritative > 0,
    ),
    ok(
      "geo_source_diversity",
      `${hosts.length} distinct external source domain(s) (3 or more recommended)`,
      hosts.length >= 3,
    ),
    ok("geo_cite_markup", "Sources marked up with <cite> or a cite attribute", citeTags + citedQuotes > 0),
    quotes > 0
      ? ok("geo_quote_attribution", attributionLabel, citeTags + citedQuotes >= quotes)
      : skip("geo_quote_attribution", "No blockquotes to attribute"),
    ok("geo_references_section", "A Sources or References section is present", references),
  ];
}

function hasBylineElement(c: PageCtx): boolean {
  for (const [name, list] of c.tags) {
    if (name === "meta" || name === "link" || name === "html" || name === "head") continue;
    for (const t of list) {
      if (lowerAttr(t.attrs.itemprop) === "author" || BYLINE_CLASS.test((t.attrs.class ?? "").toLowerCase()))
        return true;
      if (name === "a" && (t.attrs.rel ?? "").toLowerCase().split(/\s+/).includes("author")) return true;
    }
  }
  return false;
}

function experienceChecks(c: PageCtx): Check[] {
  const lower = c.text.toLowerCase();
  return [
    ok(
      "geo_experience_signals",
      "First-hand experience wording (we tested, our data, case study)",
      EXPERIENCE.test(lower),
    ),
    ok(
      "geo_expertise_signals",
      "Credentials or experience stated (PhD, certified, years of experience)",
      EXPERTISE.test(lower),
    ),
    ok("geo_editorial_review", "Reviewed, fact-checked or edited-by statement present", REVIEWED.test(lower)),
    ok(
      "geo_author_byline",
      "Visible author byline element (rel=author, itemprop=author or byline class)",
      hasBylineElement(c),
    ),
  ];
}

function siteTrustChecks(c: PageCtx): Check[] {
  let policy = false;
  let llms = false;
  let contact = false;
  for (const a of tagsNamed(c, "a")) {
    const href = lowerAttr(a.attrs.href);
    const label = `${href} ${a.text.toLowerCase()}`;
    policy = policy || label.includes("privacy") || label.includes("terms");
    llms = llms || href.includes("llms.txt");
    contact = contact || href.startsWith("mailto:") || href.startsWith("tel:");
  }
  llms = llms || tagsNamed(c, "link").some((t) => (t.attrs.href ?? "").toLowerCase().includes("llms.txt"));
  const schemaContact = parseSchemaDoc(c).nodes.some(
    (n) => present(n.email) || present(n.telephone) || present(n.contactPoint),
  );
  return [
    ok("geo_policy_links", "Privacy or terms page linked", policy),
    ok("geo_llms_txt_link", "llms.txt linked for AI crawlers", llms),
    ok(
      "geo_contact_details",
      "Contact details present (mailto, tel, address or schema)",
      contact || schemaContact || tagsNamed(c, "address").length > 0,
    ),
    snippetLimitCheck(c),
  ];
}

function snippetLimitCheck(c: PageCtx): Check {
  const limits = ["robots", "googlebot", "bingbot"]
    .flatMap((name) => metaTags(c, name))
    .flatMap((content) => [...content.toLowerCase().matchAll(MAX_SNIPPET)].map((m) => Number(m[1])));
  if (limits.length === 0) return skip("geo_max_snippet", "No max-snippet directive to check");
  const tooSmall = limits.filter((n) => n >= 0 && n < 50).length;
  return ok("geo_max_snippet", `${tooSmall} max-snippet limit(s) below 50 characters`, tooSmall === 0);
}

export function geoSignalChecks(c: PageCtx): Check[] {
  return [...publishingDateChecks(c), ...sourceQualityChecks(c), ...experienceChecks(c), ...siteTrustChecks(c)];
}
