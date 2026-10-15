import { linkTags, metaProperty, ok, type PageCtx, skip, tagsNamed } from "./extra-checks";
import type { Check } from "./index";

type Json = unknown;
export type Node = Record<string, Json>;

export const ISO_DATE = /^\d{4}-\d{2}-\d{2}([T ]\d{2}:\d{2}(:\d{2}(\.\d+)?)?(Z|[+-]\d{2}:?\d{2})?)?$/;
const HTTP_URL = /^https?:\/\//i;
const URL_KEYS = new Set(["url", "logo", "image", "sameas", "thumbnailurl", "contenturl", "embedurl"]);
export const ARTICLE_TYPES = ["article", "newsarticle", "blogposting", "techarticle", "scholarlyarticle"];
const ORG_TYPES = [
  "organization",
  "corporation",
  "ngo",
  "educationalorganization",
  "onlinestore",
  "newsmediaorganization",
];
const LOCAL_TYPES = [
  "localbusiness",
  "restaurant",
  "bakery",
  "store",
  "dentist",
  "hotel",
  "cafeorcoffeeshop",
  "barorpub",
  "medicalbusiness",
  "realestateagent",
  "autodealer",
  "healthandbeautybusiness",
];
const SCHEMA_CHECK_IDS = [
  "jsonld_valid",
  "jsonld_context",
  "schema_empty_values",
  "schema_urls_absolute",
  "org_name",
  "org_logo",
  "org_url",
  "website_name_url",
  "searchaction_valid",
  "breadcrumb_valid",
  "faq_valid",
  "article_headline",
  "article_author",
  "article_dates",
  "article_image",
  "product_offers",
  "product_price",
  "product_rating",
  "localbusiness_address",
  "localbusiness_phone",
  "video_fields",
];

export interface SchemaDoc {
  blocks: number;
  invalid: number;
  roots: Json[];
  nodes: Node[];
}

export const isNode = (v: Json): v is Node => typeof v === "object" && v !== null && !Array.isArray(v);

export function items(v: Json): Json[] {
  if (v === null || v === undefined) return [];
  return Array.isArray(v) ? v : [v];
}

export function present(v: Json): boolean {
  if (v === null || v === undefined) return false;
  if (typeof v === "string") return v.trim() !== "";
  if (Array.isArray(v)) return v.length > 0;
  if (isNode(v)) return Object.keys(v).length > 0;
  return true;
}

export function parseSchemaDoc(c: PageCtx): SchemaDoc {
  const doc: SchemaDoc = { blocks: 0, invalid: 0, roots: [], nodes: [] };
  for (const script of tagsNamed(c, "script")) {
    if ((script.attrs.type ?? "").trim().toLowerCase() !== "application/ld+json") continue;
    doc.blocks++;
    let root: Json;
    try {
      root = JSON.parse(script.text.trim());
    } catch {
      doc.invalid++;
      continue;
    }
    doc.roots.push(root);
    for (const item of items(root)) {
      if (!isNode(item)) continue;
      if ("@graph" in item) doc.nodes.push(...items(item["@graph"]).filter(isNode));
      if ("@type" in item) doc.nodes.push(item);
    }
  }
  return doc;
}

function typesOf(node: Node): string[] {
  return items(node["@type"])
    .filter((t): t is string => typeof t === "string")
    .map((t) => t.trim().toLowerCase());
}

export const ofType = (doc: SchemaDoc, ...wanted: string[]): Node[] =>
  doc.nodes.filter((n) => typesOf(n).some((t) => wanted.includes(t)));

function countStrings(root: Json, visit: (key: string, value: string) => boolean): number {
  let count = 0;
  const walk = (key: string, v: Json): void => {
    if (typeof v === "string") {
      if (visit(key, v)) count++;
    } else if (Array.isArray(v)) {
      for (const e of v) walk(key, e);
    } else if (isNode(v)) {
      for (const [k, e] of Object.entries(v)) walk(k, e);
    }
  };
  walk("", root);
  return count;
}

function contextOk(v: Json): boolean {
  if (typeof v === "string") return v.includes("schema.org");
  if (Array.isArray(v)) return v.some(contextOk);
  if (isNode(v)) return Object.values(v).some(contextOk);
  return false;
}

function numeric(v: Json): boolean {
  if (typeof v === "number") return true;
  return typeof v === "string" && v.trim() !== "" && Number.isFinite(Number(v));
}

const hasKey = (key: string) => (n: Node) => present(n[key]);

export function gated(nodes: unknown[], id: string, label: string, absent: string, pass: (n: Node) => boolean): Check {
  return nodes.length === 0 ? skip(id, absent) : ok(id, label, (nodes as Node[]).every(pass));
}

function contextChecks(doc: SchemaDoc): Check[] {
  const roots = doc.roots.flatMap(items).filter(isNode);
  const missing = roots.filter((r) => !contextOk(r["@context"])).length;
  const check =
    roots.length === 0
      ? skip("jsonld_context", "No valid JSON-LD to check")
      : ok("jsonld_context", `${missing} of ${roots.length} JSON-LD root(s) lack a schema.org @context`, missing === 0);
  const absent = "No valid JSON-LD to check";
  if (doc.roots.length === 0) {
    return [check, skip("schema_empty_values", absent), skip("schema_urls_absolute", absent)];
  }
  const empty = doc.roots.reduce<number>(
    (sum, root) => sum + countStrings(root, (key, value) => !key.startsWith("@") && value.trim() === ""),
    0,
  );
  const relative = doc.roots.reduce<number>(
    (sum, root) =>
      sum +
      countStrings(
        root,
        (key, value) => URL_KEYS.has(key.toLowerCase()) && value.trim() !== "" && !HTTP_URL.test(value.trim()),
      ),
    0,
  );
  return [
    check,
    ok("schema_empty_values", `${empty} empty string value(s) in structured data`, empty === 0),
    ok("schema_urls_absolute", `${relative} url/logo/image value(s) are not absolute http(s) URLs`, relative === 0),
  ];
}

function organizationChecks(doc: SchemaDoc): Check[] {
  const orgs = ofType(doc, ...ORG_TYPES, ...LOCAL_TYPES);
  const absent = "No Organization schema to validate";
  return [
    gated(orgs, "org_name", "Organization schema has a name", absent, hasKey("name")),
    gated(orgs, "org_logo", "Organization schema has a logo", absent, hasKey("logo")),
    gated(orgs, "org_url", "Organization schema has a url", absent, hasKey("url")),
  ];
}

function websiteChecks(doc: SchemaDoc): Check[] {
  const sites = ofType(doc, "website");
  const searches = sites.flatMap((site) =>
    items(site.potentialAction).filter((a): a is Node => isNode(a) && typesOf(a).includes("searchaction")),
  );
  return [
    gated(
      sites,
      "website_name_url",
      "WebSite schema has a name and url",
      "No WebSite schema to validate",
      (n) => present(n.name) && present(n.url),
    ),
    gated(
      searches,
      "searchaction_valid",
      "WebSite SearchAction has a target and query-input",
      "No SearchAction to validate",
      (n) => present(n.target) && present(n["query-input"]),
    ),
  ];
}

function breadcrumbAndFaqChecks(doc: SchemaDoc): Check[] {
  return [
    gated(
      ofType(doc, "breadcrumblist"),
      "breadcrumb_valid",
      "BreadcrumbList has 2+ positioned, named items",
      "No BreadcrumbList to validate",
      (n) => {
        const list = items(n.itemListElement);
        if (list.length < 2) return false;
        return list.every((item) => {
          if (!isNode(item) || !numeric(item.position)) return false;
          const target = isNode(item.item) ? item.item : {};
          return present(item.name) || present(target.name);
        });
      },
    ),
    gated(
      ofType(doc, "faqpage"),
      "faq_valid",
      "FAQPage questions each have a name and answer text",
      "No FAQPage to validate",
      (n) => {
        const questions = items(n.mainEntity);
        return (
          questions.length > 0 &&
          questions.every((q) => {
            if (!isNode(q) || !present(q.name)) return false;
            return isNode(q.acceptedAnswer) && present(q.acceptedAnswer.text);
          })
        );
      },
    ),
  ];
}

function articleChecks(doc: SchemaDoc): Check[] {
  const articles = ofType(doc, ...ARTICLE_TYPES);
  const absent = "No Article schema to validate";
  const isoString = (v: Json) => typeof v === "string" && ISO_DATE.test(v.trim());
  return [
    gated(articles, "article_headline", "Article headline is present and 110 characters or fewer", absent, (n) => {
      const headline = n.headline;
      return typeof headline === "string" && headline.trim() !== "" && [...headline].length <= 110;
    }),
    gated(articles, "article_author", "Article author is named", absent, (n) => {
      const authors = items(n.author);
      return authors.length > 0 && authors.every((a) => (isNode(a) ? present(a.name) : present(a)));
    }),
    gated(
      articles,
      "article_dates",
      "Article datePublished (and dateModified) are valid ISO dates",
      absent,
      (n) => isoString(n.datePublished) && (!("dateModified" in n) || isoString(n.dateModified)),
    ),
    gated(articles, "article_image", "Article schema has an image", absent, hasKey("image")),
  ];
}

function offerValid(v: Json): boolean {
  if (!isNode(v)) return false;
  const priced = present(v.price) || present(v.lowPrice) || present(v.priceSpecification);
  return priced && (present(v.priceCurrency) || present(v.priceSpecification));
}

function productChecks(doc: SchemaDoc): Check[] {
  const products = ofType(doc, "product");
  const absent = "No Product schema to validate";
  const ratings = doc.nodes.map((n) => n.aggregateRating).filter(isNode);
  return [
    gated(products, "product_offers", "Product schema has offers", absent, hasKey("offers")),
    gated(products, "product_price", "Product offers have a price and currency", absent, (n) => {
      const offers = items(n.offers);
      return offers.length > 0 && offers.every(offerValid);
    }),
    gated(
      ratings,
      "product_rating",
      "aggregateRating has a ratingValue and a review or rating count",
      "No aggregateRating to validate",
      (n) => present(n.ratingValue) && (present(n.reviewCount) || present(n.ratingCount)),
    ),
  ];
}

function businessAndVideoChecks(doc: SchemaDoc): Check[] {
  const local = ofType(doc, ...LOCAL_TYPES);
  const absent = "No LocalBusiness schema to validate";
  return [
    gated(local, "localbusiness_address", "LocalBusiness schema has an address", absent, hasKey("address")),
    gated(local, "localbusiness_phone", "LocalBusiness schema has a telephone", absent, hasKey("telephone")),
    gated(
      ofType(doc, "videoobject"),
      "video_fields",
      "VideoObject has a name, thumbnailUrl and uploadDate",
      "No VideoObject to validate",
      (n) => present(n.name) && present(n.thumbnailUrl) && present(n.uploadDate),
    ),
  ];
}

function schemaChecks(c: PageCtx): Check[] {
  const doc = parseSchemaDoc(c);
  if (doc.blocks === 0) {
    return SCHEMA_CHECK_IDS.map((id) => skip(id, "No JSON-LD structured data to validate"));
  }
  return [
    ok("jsonld_valid", `${doc.invalid} of ${doc.blocks} JSON-LD block(s) are not valid JSON`, doc.invalid === 0),
    ...contextChecks(doc),
    ...organizationChecks(doc),
    ...websiteChecks(doc),
    ...breadcrumbAndFaqChecks(doc),
    ...articleChecks(doc),
    ...productChecks(doc),
    ...businessAndVideoChecks(doc),
  ];
}

function samePage(c: PageCtx, a: string, b: string): boolean {
  try {
    const key = (href: string) => {
      const u = new URL(href, c.url);
      return u.host.toLowerCase() + u.pathname.replace(/\/$/, "");
    };
    return key(a) === key(b);
  } catch {
    return false;
  }
}

function socialCompletenessChecks(c: PageCtx): Check[] {
  const meta = (prop: string): string => metaProperty(c, prop);
  const image = meta("og:image");
  const ogUrl = meta("og:url");
  const canonical = linkTags(c, "canonical")
    .map((t) => (t.attrs.href ?? "").trim())
    .find(Boolean);
  return [
    ok("og_type", "Open Graph og:type is set", meta("og:type") !== ""),
    ok("og_site_name", "Open Graph og:site_name is set", meta("og:site_name") !== ""),
    image === ""
      ? skip("og_image_absolute", "No og:image to validate")
      : ok("og_image_absolute", "og:image is an absolute http(s) URL", HTTP_URL.test(image)),
    ogUrl === "" || !canonical
      ? skip("og_url_matches_canonical", "Needs both og:url and a canonical")
      : ok("og_url_matches_canonical", "og:url matches the canonical URL", samePage(c, ogUrl, canonical)),
  ];
}

export function seoSchemaChecks(c: PageCtx): Check[] {
  return [...schemaChecks(c), ...socialCompletenessChecks(c)];
}
