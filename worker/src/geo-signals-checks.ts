import { metaProperty, ok, type PageCtx, skip, tagsNamed } from "./extra-checks";
import type { Check } from "./index";
import { ARTICLE_TYPES, ISO_DATE, ofType, parseSchemaDoc, type SchemaDoc } from "./schema-checks";
import { resolveNavigable } from "./seo-page-checks";

const VISIBLE_DATE =
  /(updated|published|modified|posted|reviewed)\s*(on|:)?\s*(([a-z]{3,9}\.? \d{1,2}(st|nd|rd|th)?,? \d{4})|\d{4}-\d{2}-\d{2}|\d{1,2} [a-z]{3,9}\.? \d{4})/;
const NOT_ARTICLE = "Not an article page";

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

const withoutWww = (host: string): string => host.toLowerCase().replace(/^www\./, "");

function schemaDate(doc: SchemaDoc, key: string): string {
  for (const node of doc.nodes) {
    const value = node[key];
    if (typeof value === "string" && ISO_DATE.test(value.trim())) return value.trim();
  }
  return "";
}

function isArticlePage(c: PageCtx): boolean {
  return (
    ofType(parseSchemaDoc(c), ...ARTICLE_TYPES).length > 0 ||
    metaProperty(c, "article:published_time") !== "" ||
    articleContainsHeading(c)
  );
}

function articleContainsHeading(c: PageCtx): boolean {
  if (tagsNamed(c, "article").length !== 1) return false;
  const lower = c.raw.toLowerCase();
  const start = lower.indexOf("<article");
  if (start === -1) return false;
  const close = lower.indexOf("</article", start);
  return lower.slice(start, close === -1 ? undefined : close).includes("<h1");
}

function publishingDateChecks(c: PageCtx, article: boolean): Check[] {
  const doc = parseSchemaDoc(c);
  const published = schemaDate(doc, "datePublished") || metaProperty(c, "article:published_time");
  const modified = schemaDate(doc, "dateModified") || metaProperty(c, "article:modified_time");
  const comparable = published.length >= 10 && modified.length >= 10;
  return [
    article
      ? ok("geo_date_published", "Publish date declared in schema or article:published_time", published !== "")
      : skip("geo_date_published", NOT_ARTICLE),
    article
      ? ok("geo_date_modified", "Modified date declared in schema or article:modified_time", modified !== "")
      : skip("geo_date_modified", NOT_ARTICLE),
    comparable
      ? ok(
          "geo_modified_after_published",
          "Modified date is not earlier than the publish date",
          modified.slice(0, 10) >= published.slice(0, 10),
        )
      : skip("geo_modified_after_published", "Needs both a publish and a modified date"),
    article
      ? ok(
          "geo_visible_date",
          "A visible published or updated date appears on the page",
          VISIBLE_DATE.test(c.text.toLowerCase()),
        )
      : skip("geo_visible_date", NOT_ARTICLE),
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

function sourceQualityChecks(c: PageCtx, article: boolean): Check[] {
  const hosts = externalHosts(c);
  const authoritative = hosts.filter(isAuthoritativeHost).length;
  const citeTags = tagsNamed(c, "cite").length;
  const citedQuotes = ["blockquote", "q"].reduce(
    (sum, name) => sum + tagsNamed(c, name).filter((t) => (t.attrs.cite ?? "").trim() !== "").length,
    0,
  );
  const quotes = tagsNamed(c, "blockquote").length;
  const attributionLabel = `${citeTags + citedQuotes} attributed source(s) for ${quotes} blockquote(s)`;
  return [
    article
      ? ok(
          "geo_authoritative_links",
          `${authoritative} link(s) to government, academic or reference sources`,
          authoritative > 0,
        )
      : skip("geo_authoritative_links", NOT_ARTICLE),
    article
      ? ok(
          "geo_source_diversity",
          `${hosts.length} distinct external source domain(s) (3 or more recommended)`,
          hosts.length >= 3,
        )
      : skip("geo_source_diversity", NOT_ARTICLE),
    quotes > 0
      ? ok("geo_quote_attribution", attributionLabel, citeTags + citedQuotes >= quotes)
      : skip("geo_quote_attribution", "No blockquotes to attribute"),
  ];
}

function policyLinkCheck(c: PageCtx): Check {
  const policy = tagsNamed(c, "a").some((a) => {
    const label = `${(a.attrs.href ?? "").trim().toLowerCase()} ${a.text.toLowerCase()}`;
    return label.includes("privacy") || label.includes("terms");
  });
  return ok("geo_policy_links", "Privacy or terms page linked", policy);
}

export function geoSignalChecks(c: PageCtx): Check[] {
  const article = isArticlePage(c);
  return [...publishingDateChecks(c, article), ...sourceQualityChecks(c, article), policyLinkCheck(c)];
}
