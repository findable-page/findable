import UI_HTML from "../../internal/server/web/index.html";
import { aeoStructureChecks } from "./aeo-structure-checks";
import { pageSpeed } from "./cwv";
import { buildCtx, seoExtraChecks } from "./extra-checks";
import { geoSignalChecks } from "./geo-signals-checks";
import { BlockedTargetError, guardedFetch, type SelfRoute } from "./guard";
import { escapeHtml, LEARN_STYLE, renderLearnIndex, renderLearnNotFound, renderLearnPost } from "./learn";
import { learnPosts } from "./learn-content";
import { aeoParityChecks, geoParityChecks, paragraphWordCounts, seoParityChecks } from "./page-checks";
import { scoreQuip, shareDescription } from "./quips";
import { fixedWindowLimiter } from "./ratelimit";
import { sanitizeHtml } from "./sanitize";
import { seoSchemaChecks } from "./schema-checks";
import { seoPageChecks } from "./seo-page-checks";
import type { CardScores } from "./share-card";
import { type SiteCategory, siteChecks } from "./site";
import { headSlot, injectHead, llmsFullTxt, llmsTxt, type PageEntry, robotsTxt, sitemapXml } from "./site-files";
import { MAX_KEYPHRASE, parseSuggestion, SUGGEST_MAX_TOKENS, SUGGEST_MODEL, suggestPrompt } from "./suggest";
import { parseWatch, saveWatch } from "./watch";

export interface Check {
  id: string;
  label: string;
  pass: boolean;
  applicable?: boolean | null;
  severity?: string;
  category?: string;
}

export interface CategoryResult {
  score: number;
  passed: number;
  total: number;
  checks: Check[];
}

export interface ScanResponse {
  url: string;
  overall: number;
  seo: CategoryResult;
  aeo: CategoryResult;
  geo: CategoryResult;
  cwv: { score: null; checks: Check[]; passed: number; total: number };
  internalLinks: string[];
  scannedAt: string;
  site?: SiteCategory;
}

const CORS_HEADERS = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
  "Access-Control-Allow-Headers": "Content-Type",
};

export interface Env {
  SCAN_LIMITER?: RateLimit;
  WATCH_DB?: D1Database;
  PSI_KEY?: string;
  AI?: Ai;
}

const MAX_ANALYZED_HTML_CHARS = 1024 * 1024;

const fallbackAllowScan = fixedWindowLimiter(10, 60_000);

async function allowScan(env: Env, key: string): Promise<boolean> {
  if (!env.SCAN_LIMITER) return fallbackAllowScan(key);
  return (await env.SCAN_LIMITER.limit({ key })).success;
}

const PUBLIC_ORIGIN = "https://findable.page";

const PAGE_HTML = injectHead(
  UI_HTML as unknown as string,
  `${headSlot(PUBLIC_ORIGIN)}<meta name="findable-watch" content="on">`,
);

let uiPolicy: Promise<string> | undefined;

async function sha256Source(text: string): Promise<string> {
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(text)));
  return `'sha256-${btoa(String.fromCharCode(...digest))}'`;
}

async function hashesOf(tag: string): Promise<string> {
  const blocks = [...PAGE_HTML.matchAll(new RegExp(`<${tag}(?:\\s[^>]*)?>([\\s\\S]*?)</${tag}>`, "g"))];
  return (await Promise.all(blocks.map((m) => sha256Source(m[1] ?? "")))).join(" ");
}

async function buildUiPolicy(): Promise<string> {
  return [
    "default-src 'none'",
    `script-src ${await hashesOf("script")} https://static.cloudflareinsights.com`,
    "style-src 'unsafe-inline'",
    `style-src-elem ${await hashesOf("style")}`,
    "style-src-attr 'unsafe-inline'",
    "connect-src 'self' https://cloudflareinsights.com",
    "img-src 'self' data:",
    "manifest-src 'self'",
    "base-uri 'none'",
    "form-action 'self'",
    "frame-ancestors 'none'",
  ].join("; ");
}

function securityHeaders(csp: string): Record<string, string> {
  return {
    "Content-Security-Policy": csp,
    "X-Content-Type-Options": "nosniff",
    "Strict-Transport-Security": "max-age=31536000; includeSubDomains",
    "X-Frame-Options": "DENY",
    "Referrer-Policy": "strict-origin-when-cross-origin",
  };
}

const HOME_PAGE: PageEntry = {
  path: "/",
  title: "findable",
  description: "Scan any URL for SEO, AEO and GEO signals and get a score with the fixes that matter.",
};

const LEARN_PAGES: PageEntry[] = [
  {
    path: "/learn",
    title: "Learn findable",
    description: "Plain guides to SEO, AEO and GEO, and how to check them with findable.",
  },
  ...learnPosts.map((p) => ({ path: `/learn/${p.slug}`, title: p.title, description: p.description })),
];

const SITE_PAGES = [HOME_PAGE, ...LEARN_PAGES];

const SITE_FILES: Record<string, (origin: string) => { type: string; body: string }> = {
  "/robots.txt": (o) => ({ type: "text/plain", body: robotsTxt(o) }),
  "/sitemap.xml": (o) => ({
    type: "application/xml",
    body: sitemapXml(
      o,
      SITE_PAGES.map((p) => p.path),
    ),
  }),
  "/llms.txt": (o) => ({ type: "text/plain", body: llmsTxt(o, SITE_PAGES) }),
  "/llms-full.txt": (o) => ({ type: "text/plain", body: llmsFullTxt(o, SITE_PAGES) }),
};

let learnPolicy: Promise<string> | undefined;

async function buildLearnPolicy(): Promise<string> {
  return [
    "default-src 'none'",
    `style-src ${await sha256Source(LEARN_STYLE)}`,
    "img-src 'self'",
    "manifest-src 'self'",
    "base-uri 'none'",
    "form-action 'none'",
    "frame-ancestors 'none'",
  ].join("; ");
}

async function learnResponse(pathname: string, origin: string): Promise<Response> {
  const slug = pathname.slice("/learn".length).replace(/^\/|\/$/g, "");
  const html = slug ? renderLearnPost(slug, origin) : renderLearnIndex(origin);
  learnPolicy ??= buildLearnPolicy();
  return new Response(html ?? renderLearnNotFound(origin), {
    status: html ? 200 : 404,
    headers: { "Content-Type": "text/html; charset=utf-8", ...securityHeaders(await learnPolicy) },
  });
}

const OWN_HOSTS = ["findable.page", "seoscore.piply.xyz"];

const worker = {
  async fetch(request: Request, env: Env = {}): Promise<Response> {
    const url = new URL(request.url);

    if (request.method === "OPTIONS") {
      return new Response(null, { headers: CORS_HEADERS });
    }

    if (url.pathname === "/") {
      return pageResponse(url.searchParams.get("url"), request, env, url);
    }

    const siteFile = request.method === "GET" ? SITE_FILES[url.pathname] : undefined;
    if (siteFile) {
      const { type, body } = siteFile(url.origin);
      return new Response(body, { headers: { ...CORS_HEADERS, "Content-Type": `${type}; charset=utf-8` } });
    }

    if (request.method === "GET" && (url.pathname === "/learn" || url.pathname.startsWith("/learn/"))) {
      return learnResponse(url.pathname, url.origin);
    }

    if (url.pathname === "/og/card.png" && request.method === "GET") {
      return shareCardResponse(url, request);
    }

    if (url.pathname === "/health") {
      return new Response(JSON.stringify({ status: "ok", name: "findable", version: "0.3.0" }), {
        headers: { ...CORS_HEADERS, "Content-Type": "application/json" },
      });
    }

    if (url.pathname === "/api/scan" && request.method === "POST") {
      try {
        const key = rateLimitKey(request.headers.get("CF-Connecting-IP") ?? "unknown");
        if (!(await allowScan(env, key))) {
          return new Response(JSON.stringify({ error: "too many scans, try again in a minute" }), {
            status: 429,
            headers: { ...CORS_HEADERS, "Content-Type": "application/json" },
          });
        }
        const body = await readScanBody(request);
        if (typeof body === "string") return badRequest(body);

        const targetUrl = withScheme(body.url);
        const self: SelfRoute = { hosts: [url.hostname, ...OWN_HOSTS], handle: (r) => worker.fetch(r, env) };
        const scanRes = await auditUrl(targetUrl, self, body.keyphrase?.trim());
        if (body.site === true) scanRes.site = await siteChecks(targetUrl, self);

        return new Response(JSON.stringify(scanRes), {
          headers: { ...CORS_HEADERS, "Content-Type": "application/json" },
        });
      } catch (err) {
        const message = err instanceof Error && err.message ? err.message : "audit failed";
        return new Response(JSON.stringify({ error: message }), {
          status: err instanceof BlockedTargetError ? 400 : 500,
          headers: { ...CORS_HEADERS, "Content-Type": "application/json" },
        });
      }
    }

    if (url.pathname === "/api/cwv" && request.method === "POST") {
      return cwvResponse(request, env);
    }

    if (url.pathname === "/api/suggest" && request.method === "POST") {
      return suggestResponse(request, url, env);
    }

    if (url.pathname === "/api/watch" && request.method === "POST") {
      return watchResponse(request, url, env);
    }

    const pathTarget = url.pathname.slice(1);
    if (request.method === "GET" && BARE_HOST.test(pathTarget)) {
      return pageResponse(pathTarget, request, env, url);
    }

    return new Response("Not Found", { status: 404, headers: CORS_HEADERS });
  },
};

// Bots from many platforms unfurl a viral link at once; one scan per host per hour keeps them under the scan limit.
async function cachedSharePreview(
  target: string,
  env: Env,
  url: URL,
): Promise<{ html: string; cacheStatus: "hit" | "miss" }> {
  const cache = (globalThis as { caches?: { default: Cache } }).caches?.default;
  const key = new Request(`${PUBLIC_ORIGIN}/__share-preview/v2/${withScheme(target).toLowerCase()}`);
  const hit = await cache?.match(key);
  if (hit) return { html: await hit.text(), cacheStatus: "hit" };
  const html = await sharePreviewHtml(target, env, url);
  if (html !== PAGE_HTML) {
    await cache?.put(key, new Response(html, { headers: { "Cache-Control": "public, max-age=3600" } }));
  }
  return { html, cacheStatus: "miss" };
}

async function pageResponse(target: string | null, request: Request, env: Env, url: URL): Promise<Response> {
  uiPolicy ??= buildUiPolicy();
  const isPreviewBot = PREVIEW_BOTS.test(request.headers.get("User-Agent") ?? "");
  const preview = target && isPreviewBot ? await cachedSharePreview(target, env, url) : null;
  return new Response(preview?.html ?? PAGE_HTML, {
    headers: {
      ...(preview && { "X-Findable-Cache": preview.cacheStatus }),
      ...CORS_HEADERS,
      "Content-Type": "text/html; charset=utf-8",
      ...securityHeaders(await uiPolicy),
    },
  });
}

export function prettyShareUrl(target: string): string {
  const bare = target
    .trim()
    .replace(/^https:\/\//i, "")
    .replace(/^([^/?#]+)\/$/, "$1");
  return BARE_HOST.test(bare)
    ? `${PUBLIC_ORIGIN}/${bare}`
    : `${PUBLIC_ORIGIN}/?${new URLSearchParams({ url: target })}`;
}

export default worker;

const PREVIEW_BOTS =
  /facebookexternalhit|facebot|linkedinbot|twitterbot|slackbot|discordbot|telegrambot|whatsapp|skypeuripreview|redditbot|bluesky|mastodon|embedly|iframely/i;

function replaceMeta(html: string, attr: string, key: string, value: string): string {
  return html.replace(
    new RegExp(`(<meta ${attr}="${key}" content=")[^"]*(")`),
    (_m, open: string, close: string) => open + escapeHtml(value) + close,
  );
}

export function cardScores(params: URLSearchParams): CardScores | null {
  const host = params.get("h") ?? "";
  const score = (k: string) => {
    const n = Number(params.get(k));
    return Number.isInteger(n) && n >= 0 && n <= 100 ? n : null;
  };
  const [overall, seo, aeo, geo] = ["o", "s", "a", "g"].map(score);
  if (!/^[a-z0-9.-]{1,253}$/i.test(host) || overall == null || seo == null || aeo == null || geo == null) return null;
  return { host, overall, seo, aeo, geo };
}

async function shareCardResponse(url: URL, request: Request): Promise<Response> {
  const scores = cardScores(url.searchParams);
  if (!scores) return Response.redirect(`${url.origin}/og.png`, 302);
  const cache = (globalThis as { caches?: { default: Cache } }).caches?.default;
  const hit = await cache?.match(request);
  if (hit) {
    const resp = new Response(hit.body, hit);
    resp.headers.set("X-Findable-Cache", "hit");
    return resp;
  }
  try {
    const { renderShareCard } = await import("./og-render");
    const png = await (await renderShareCard(scores)).arrayBuffer();
    const headers = { "Content-Type": "image/png", "Cache-Control": "public, max-age=86400" };
    await cache?.put(request, new Response(png, { headers }));
    return new Response(png, { headers: { ...headers, "X-Findable-Cache": "miss" } });
  } catch {
    return Response.redirect(`${url.origin}/og.png`, 302);
  }
}

async function sharePreviewHtml(target: string, env: Env, requestUrl: URL): Promise<string> {
  try {
    if (!(await allowScan(env, "share-preview"))) return PAGE_HTML;
    const self: SelfRoute = { hosts: [requestUrl.hostname, ...OWN_HOSTS], handle: (r) => worker.fetch(r, env) };
    const scan = await auditUrl(withScheme(target), self);
    const host = new URL(scan.url).hostname;
    const title = `${host} scored ${scan.overall}/100. ${scoreQuip(scan.overall, host)}`;
    const description = shareDescription({ seo: scan.seo.score, aeo: scan.aeo.score, geo: scan.geo.score }, host);
    const shareUrl = prettyShareUrl(target);
    let html = PAGE_HTML.replace(/<title>[^<]*<\/title>/, `<title>${escapeHtml(title)}</title>`);
    html = replaceMeta(html, "property", "og:title", title);
    html = replaceMeta(html, "name", "twitter:title", title);
    html = replaceMeta(html, "property", "og:description", description);
    html = replaceMeta(html, "name", "twitter:description", description);
    const card = new URLSearchParams({
      h: host,
      o: String(scan.overall),
      s: String(scan.seo.score),
      a: String(scan.aeo.score),
      g: String(scan.geo.score),
    });
    const image = `${PUBLIC_ORIGIN}/og/card.png?${card}`;
    html = replaceMeta(html, "property", "og:image", image);
    html = replaceMeta(html, "name", "twitter:image", image);
    return replaceMeta(html, "property", "og:url", shareUrl);
  } catch {
    return PAGE_HTML;
  }
}

interface ScanBody {
  url: string;
  keyphrase?: string | undefined;
  site?: boolean;
}

const BARE_HOST = /^[a-z0-9-]+(\.[a-z0-9-]+)+(:\d+)?([/?#].*)?$/i;

function withScheme(target: string): string {
  const t = target.trim();
  return BARE_HOST.test(t) ? `https://${t}` : t;
}

function badRequest(error: string): Response {
  return new Response(JSON.stringify({ error }), {
    status: 400,
    headers: { ...CORS_HEADERS, "Content-Type": "application/json" },
  });
}

async function cwvResponse(request: Request, env: Env): Promise<Response> {
  const json = (status: number, body: object) =>
    new Response(JSON.stringify(body), { status, headers: { ...CORS_HEADERS, "Content-Type": "application/json" } });
  if (!env.PSI_KEY) return json(503, { error: "page speed is not set up" });
  const body = await readScanBody(request);
  if (typeof body === "string") return json(400, { error: body });
  const target = withScheme(body.url);
  if (!/^https?:\/\//i.test(target)) return json(400, { error: "url must be http or https" });
  const cacheKey = new Request(`https://findable.page/__cwv/v1/${encodeURIComponent(target)}`);
  const cache = (globalThis as { caches?: { default?: Cache } }).caches?.default;
  const hit = await cache?.match(cacheKey);
  if (hit) return json(200, await hit.json());
  const ip = rateLimitKey(request.headers.get("CF-Connecting-IP") ?? "unknown");
  if (!(await allowScan(env, `cwv:${ip}`))) return json(429, { error: "too many speed tests, try again in a minute" });
  try {
    const cwv = await pageSpeed(target, env.PSI_KEY);
    await cache?.put(cacheKey, new Response(JSON.stringify(cwv), { headers: { "Cache-Control": "max-age=86400" } }));
    return json(200, cwv);
  } catch (err) {
    console.error("cwv: pagespeed failed", err);
    return json(502, { error: "page speed test failed, try again later" });
  }
}

async function suggestResponse(request: Request, url: URL, env: Env): Promise<Response> {
  const json = (status: number, body: object) =>
    new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
  if (foreignOrigin(request, url)) return json(403, { error: "suggestions only from findable.page" });
  if (!env.AI) return json(503, { error: "AI suggestions are not set up" });
  const body = await readScanBody(request);
  if (typeof body === "string") return json(400, { error: body });
  const target = withScheme(body.url);
  if (!/^https?:\/\//i.test(target)) return json(400, { error: "url must be http or https" });
  const keyphrase = body.keyphrase?.trim() ?? "";
  if (keyphrase.length > MAX_KEYPHRASE)
    return json(400, { error: `keyphrase must be at most ${MAX_KEYPHRASE} characters` });
  try {
    const cacheKey = new Request(
      `https://findable.page/__suggest/v1/${encodeURIComponent(new URL(target).href)}/${encodeURIComponent(keyphrase)}`,
    );
    const cache = (globalThis as { caches?: { default?: Cache } }).caches?.default;
    const hit = await cache?.match(cacheKey);
    if (hit) return json(200, await hit.json());
    const ip = rateLimitKey(request.headers.get("CF-Connecting-IP") ?? "unknown");
    if (!(await allowScan(env, `suggest:${ip}`)))
      return json(429, { error: "too many requests, try again in a minute" });
    const self: SelfRoute = { hosts: [url.hostname, ...OWN_HOSTS], handle: (r) => worker.fetch(r, env) };
    const page = await guardedFetch(
      target,
      {
        "User-Agent": "Mozilla/5.0 (compatible; findable-worker/1.0; +https://findable.page)",
        Accept: "text/html,application/xhtml+xml",
      },
      self,
    );
    if (page.status >= 400 || !/html/i.test(page.headers.get("content-type") ?? ""))
      return json(422, { error: "that page is not an HTML page we can read" });
    const out = (await env.AI.run(SUGGEST_MODEL, {
      messages: suggestPrompt(page.body, keyphrase),
      max_tokens: SUGGEST_MAX_TOKENS,
    })) as { response?: string };
    const suggestion = parseSuggestion(out.response ?? "");
    if (!suggestion) return json(502, { error: "the AI returned an unusable answer, try again" });
    await cache?.put(
      cacheKey,
      new Response(JSON.stringify(suggestion), { headers: { "Cache-Control": "max-age=86400" } }),
    );
    return json(200, suggestion);
  } catch (err) {
    if (err instanceof BlockedTargetError) return json(400, { error: err.message });
    console.error("suggest: failed", err);
    return json(502, { error: "could not generate suggestions, try again later" });
  }
}

function foreignOrigin(request: Request, url: URL): boolean {
  const origin = request.headers.get("Origin");
  return !!origin && ![url.hostname, ...OWN_HOSTS].includes(new URL(origin).hostname);
}

function watchJson(status: number, body: object): Response {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

async function watchResponse(request: Request, url: URL, env: Env): Promise<Response> {
  if (foreignOrigin(request, url)) return watchJson(403, { error: "signups only from findable.page" });
  if (!request.headers.get("Content-Type")?.startsWith("application/json")) {
    return watchJson(415, { error: "send json" });
  }
  const ip = rateLimitKey(request.headers.get("CF-Connecting-IP") ?? "unknown");
  if (!(await allowScan(env, `watch:${ip}`)))
    return watchJson(429, { error: "too many signups, try again in a minute" });
  const signup = parseWatch(await request.json().catch(() => null));
  if (signup === "spam") {
    console.log("watch: honeypot hit");
    return watchJson(200, { ok: true });
  }
  if (typeof signup === "string") return watchJson(400, { error: signup });
  try {
    if (!env.WATCH_DB) throw new Error("WATCH_DB not bound");
    await saveWatch(env.WATCH_DB, signup);
  } catch (err) {
    console.error("watch: save failed", err);
    return watchJson(503, { error: "could not save, try again" });
  }
  return watchJson(200, { ok: true });
}

async function readScanBody(request: Request): Promise<ScanBody | string> {
  const body: unknown = await request.json().catch(() => null);
  if (typeof body !== "object" || body === null || Array.isArray(body)) return "body must be a JSON object";
  const { url, keyphrase, site } = body as Record<string, unknown>;
  if (typeof url !== "string" || !url.trim()) return "missing url";
  if (keyphrase !== undefined && typeof keyphrase !== "string") return "keyphrase must be a string";
  return { url, keyphrase, site: site === true };
}

function rateLimitKey(ip: string): string {
  if (!ip.includes(":")) return ip;
  const [head = "", tail] = ip.toLowerCase().split("::");
  const headGroups = head ? head.split(":") : [];
  const tailGroups = tail ? tail.split(":") : [];
  const zeros = tail === undefined ? [] : Array(8 - headGroups.length - tailGroups.length).fill("0");
  const prefix = [...headGroups, ...zeros, ...tailGroups].slice(0, 4);
  return `${prefix.map((g) => Number.parseInt(g, 16).toString(16)).join(":")}::/64`;
}

async function auditUrl(targetUrl: string, self: SelfRoute, keyphrase?: string): Promise<ScanResponse> {
  const page = await guardedFetch(
    targetUrl,
    {
      "User-Agent": "Mozilla/5.0 (compatible; findable-worker/1.0; +https://findable.page)",
      Accept: "text/html,application/xhtml+xml",
    },
    self,
  );
  const scan = analyzeHtml(page.url, page.body.slice(0, MAX_ANALYZED_HTML_CHARS), page.headers, keyphrase);
  return { ...scan, url: targetUrl };
}

function firstGroup(match: RegExpMatchArray | null): string {
  return match?.[1]?.trim() ?? "";
}

export function analyzeHtml(targetUrl: string, pageHtml: string, headers: Headers, keyphrase?: string): ScanResponse {
  const rawHtml = sanitizeHtml(pageHtml);
  const lowerHtml = rawHtml.toLowerCase();
  const parsedUrl = new URL(targetUrl);

  // Meta signals
  const titleMatch = rawHtml.match(/<title[^>]*>([^<]*)<\/title>/i);
  const title = firstGroup(titleMatch);

  const descMatch =
    rawHtml.match(/<meta[^>]+name=["']description["'][^>]+content=["']([^"']*)["']/i) ||
    rawHtml.match(/<meta[^>]+content=["']([^"']*)["'][^>]+name=["']description["']/i);
  const metaDesc = firstGroup(descMatch);

  const canonicalMatch =
    rawHtml.match(/<link[^>]+rel=["']canonical["'][^>]+href=["']([^"']*)["']/i) ||
    rawHtml.match(/<link[^>]+href=["']([^"']*)["'][^>]+rel=["']canonical["']/i);
  const canonical = firstGroup(canonicalMatch);

  const viewportMatch = rawHtml.match(/<meta[^>]+name=["']viewport["']/i);
  const charsetMatch = rawHtml.match(/<meta[^>]+charset=["']?utf-8/i) || lowerHtml.includes("charset=utf-8");
  const langMatch = rawHtml.match(/<html[^>]+lang=["']([^"']+)["']/i);
  const doctypeMatch = rawHtml.startsWith("<!DOCTYPE html") || rawHtml.startsWith("<!doctype html");
  const robotsMatch = rawHtml.match(/<meta[^>]+name=["']robots["'][^>]+content=["']([^"']*)["']/i);
  const robotsMeta = (robotsMatch?.[1] ?? "").toLowerCase();

  // Headings
  const h1Matches = [...rawHtml.matchAll(/<h1[^>]*>([\s\S]*?)<\/h1>/gi)];
  const h2Matches = [...rawHtml.matchAll(/<h2[^>]*>([\s\S]*?)<\/h2>/gi)];

  // Text content & word count
  const visibleText = rawHtml
    .replace(/<script[\s\S]*?<\/script>/gi, " ")
    .replace(/<style[\s\S]*?<\/style>/gi, " ")
    .replace(/<svg[\s\S]*?<\/svg>/gi, " ")
    .replace(/<[^>]+>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
  const words = visibleText.split(/\s+/).filter(Boolean);
  const wordCount = words.length;
  const hasDirectAnswer = paragraphWordCounts(rawHtml).some((n) => n >= 30 && n <= 80);

  // Open Graph & Twitter
  const ogTitle = rawHtml.match(/<meta[^>]+property=["']og:title["']/i);
  const ogDesc = rawHtml.match(/<meta[^>]+property=["']og:description["']/i);
  const ogImage = rawHtml.match(/<meta[^>]+property=["']og:image["']/i);
  const ogUrl = rawHtml.match(/<meta[^>]+property=["']og:url["']/i);
  const twitterCard = rawHtml.match(/<meta[^>]+name=["']twitter:card["']/i);

  // Links
  const linkMatches = [...rawHtml.matchAll(/<a[^>]+href=["']([^"']+)["']/gi)];
  const internalLinks: string[] = [];
  const externalLinks: string[] = [];
  for (const m of linkMatches) {
    const href = firstGroup(m);
    if (href.startsWith("#") || href.startsWith("mailto:") || href.startsWith("tel:") || href.startsWith("javascript:"))
      continue;
    try {
      const resolved = new URL(href, targetUrl);
      if (resolved.hostname === parsedUrl.hostname) {
        internalLinks.push(resolved.href);
      } else {
        externalLinks.push(resolved.href);
      }
    } catch {}
  }

  // Schema JSON-LD
  const jsonLdMatches = [
    ...rawHtml.matchAll(/<script[^>]+type=["']application\/ld\+json["'][^>]*>([\s\S]*?)<\/script>/gi),
  ];
  const schemaTypes: string[] = [];
  const sameAsLinks: string[] = [];
  let hasFAQ = false;
  let hasPerson = false;
  let hasOrg = false;
  let hasSpeakable = false;

  for (const jm of jsonLdMatches) {
    try {
      const parsed = JSON.parse(firstGroup(jm));
      const scanObj = (value: unknown, pageLevel: boolean, topLevel: boolean): void => {
        if (Array.isArray(value)) {
          for (const v of value) scanObj(v, pageLevel, topLevel);
          return;
        }
        if (!value || typeof value !== "object") return;
        const o = value as Record<string, unknown>;
        if (topLevel && "speakable" in o) hasSpeakable = true;
        const types = Array.isArray(o["@type"]) ? o["@type"] : o["@type"] ? [o["@type"]] : [];
        for (const raw of types) {
          const t = String(raw);
          schemaTypes.push(t);
          if (t.toLowerCase() === "faqpage") hasFAQ = true;
          if (pageLevel && t.toLowerCase() === "person") hasPerson = true;
          if (["organization", "bakery", "localbusiness"].includes(t.toLowerCase())) hasOrg = true;
        }
        if (pageLevel && o.sameAs) {
          if (Array.isArray(o.sameAs)) sameAsLinks.push(...o.sameAs.map(String));
          else sameAsLinks.push(String(o.sameAs));
        }
        for (const [key, child] of Object.entries(o)) {
          if (child && typeof child === "object")
            scanObj(child, pageLevel && (key === "@graph" || key === "author"), topLevel && key === "@graph");
        }
      };
      scanObj(parsed, true, true);
    } catch {}
  }

  // Formatting structures
  const hasDl = lowerHtml.includes("<dl") && lowerHtml.includes("<dt") && lowerHtml.includes("<dd");
  const hasOl = lowerHtml.includes("<ol");
  const hasUl = lowerHtml.includes("<ul");
  const hasTable = lowerHtml.includes("<table");
  const hasCode = lowerHtml.includes("<code") || lowerHtml.includes("<pre");
  const hasDetails = lowerHtml.includes("<details") && lowerHtml.includes("<summary");
  const hasBlockquote = lowerHtml.includes("<blockquote");
  const hasTime = lowerHtml.includes("<time");

  // Statistics & pronouns
  const statsMatches =
    visibleText.match(
      /(?:\$\s*\d+(?:\.\d+)?|\b\d+(?:\.\d+)?\s*(?:%|ms|s|MB|GB|KB|px|M|B|k|users|queries|hours|pages)\b)/g,
    ) || [];
  const lowerText = visibleText.toLowerCase();
  const youCount = (lowerText.match(/\b(you|your|we)\b/g) || []).length;

  const ctx = buildCtx(rawHtml, parsedUrl, visibleText);

  // --- SEO Checks ---
  const seoChecks: Check[] = [
    { id: "title_exists", label: title ? `Title tag found: "${title}"` : "No title tag found", pass: Boolean(title) },
    {
      id: "title_length",
      label: `Title length: ${title.length} characters (aim for 30-60)`,
      pass: title.length >= 30 && title.length <= 60,
    },
    {
      id: "desc_exists",
      label: metaDesc ? `Meta description found (${metaDesc.length} chars)` : "Missing meta description",
      pass: Boolean(metaDesc),
    },
    {
      id: "desc_length",
      label: `Description length: ${metaDesc.length} chars (aim for 120-160)`,
      pass: metaDesc.length >= 120 && metaDesc.length <= 160,
    },
    { id: "viewport", label: "Viewport meta tag found (mobile-friendly)", pass: Boolean(viewportMatch) },
    { id: "charset", label: "UTF-8 charset declared", pass: Boolean(charsetMatch) },
    { id: "lang_attr", label: "Language attribute set on HTML tag", pass: Boolean(langMatch) },
    { id: "html_doctype", label: "<!DOCTYPE html> declared", pass: Boolean(doctypeMatch) },
    { id: "robots_meta", label: "No robots meta tag blocking crawling", pass: !robotsMeta.includes("noindex") },
    {
      id: "h1_exists",
      label:
        h1Matches.length === 1 ? "Single H1 tag found" : `${h1Matches.length} H1 tags found (exactly 1 recommended)`,
      pass: h1Matches.length === 1,
    },
    { id: "h2_exists", label: `${h2Matches.length} H2 tags found`, pass: h2Matches.length > 0 },
    { id: "word_count", label: `${wordCount} words in body (300+ recommended)`, pass: wordCount >= 300 },
    { id: "internal_links", label: `${internalLinks.length} internal links found`, pass: internalLinks.length >= 2 },
    { id: "external_links", label: `${externalLinks.length} external links found`, pass: externalLinks.length >= 1 },
    { id: "canonical", label: "Canonical tag found", pass: Boolean(canonical) },
    { id: "og_title", label: "Open Graph title found", pass: Boolean(ogTitle) },
    { id: "og_desc", label: "Open Graph description found", pass: Boolean(ogDesc) },
    { id: "og_image", label: "Open Graph image found", pass: Boolean(ogImage) },
    { id: "og_url", label: "Open Graph URL found", pass: Boolean(ogUrl) },
    { id: "twitter_card", label: "Twitter card declared", pass: Boolean(twitterCard) },
    { id: "https", label: "HTTPS enabled", pass: parsedUrl.protocol === "https:" },
    {
      id: "hsts",
      label: "Strict-Transport-Security header enforced",
      pass: Boolean(headers.get("strict-transport-security")),
    },
    {
      id: "x_content_type",
      label: "X-Content-Type-Options: nosniff header present",
      pass: (headers.get("x-content-type-options") || "").includes("nosniff"),
    },
    { id: "x_frame_options", label: "X-Frame-Options header present", pass: Boolean(headers.get("x-frame-options")) },
    {
      id: "csp",
      label: "Content-Security-Policy header declared",
      pass: Boolean(headers.get("content-security-policy")),
    },
    { id: "jsonld", label: `${jsonLdMatches.length} JSON-LD schema block(s) found`, pass: jsonLdMatches.length > 0 },
    ...seoParityChecks(rawHtml, visibleText),
    ...seoExtraChecks(ctx),
    ...seoPageChecks(ctx),
    ...seoSchemaChecks(ctx),
  ];

  if (keyphrase) {
    const kp = keyphrase.toLowerCase();
    seoChecks.push({
      id: "keyphrase_in_title",
      label: `Focus keyphrase "${keyphrase}" in title tag`,
      pass: title.toLowerCase().includes(kp),
    });
    seoChecks.push({
      id: "keyphrase_in_desc",
      label: `Focus keyphrase "${keyphrase}" in meta description`,
      pass: metaDesc.toLowerCase().includes(kp),
    });
    seoChecks.push({
      id: "keyphrase_in_h1",
      label: `Focus keyphrase "${keyphrase}" in H1 heading`,
      pass: h1Matches.some((m) => (m[1] ?? "").toLowerCase().includes(kp)),
    });
    seoChecks.push({
      id: "keyphrase_in_content",
      label: `Focus keyphrase "${keyphrase}" appears in body copy`,
      pass: lowerText.includes(kp),
    });
  }

  // --- AEO Checks ---
  const aeoChecks: Check[] = [
    { id: "aeo_structured_data", label: "Structured data found for AI extraction", pass: jsonLdMatches.length > 0 },
    { id: "aeo_faq", label: "FAQ schema or Q&A section detected", pass: hasFAQ },
    { id: "aeo_direct_answers", label: "Concise answer paragraphs (30-80 words) found", pass: hasDirectAnswer },
    {
      id: "aeo_question_headings",
      label: "Question-style headings found (What/How/Why/?)",
      pass: h2Matches.some((m) => /^(what|how|why)|\?$/i.test((m[1] ?? "").trim())),
    },
    { id: "aeo_definition_lists", label: "Definition lists (<dl>) present for terminology", pass: hasDl },
    { id: "aeo_numbered_steps", label: "Numbered steps (<ol>) present for instructions", pass: hasOl },
    { id: "aeo_tables", label: "Data tables present for comparison", pass: hasTable },
    { id: "aeo_lists", label: "Lists (<ul>/<ol>) present for scannability", pass: hasUl || hasOl },
    { id: "aeo_code_blocks", label: "Code blocks (<code>/<pre>) present for technical clarity", pass: hasCode },
    { id: "aeo_collapsible", label: "<details>/<summary> elements present", pass: hasDetails },
    {
      id: "aeo_conversational",
      label: "Conversational tone with user pronouns ('you', 'your', 'we')",
      pass: youCount >= 3,
    },
    {
      id: "aeo_summary",
      label: "Summary or Key Takeaways section present",
      pass: /key takeaway|summary|conclusion/i.test(lowerText),
    },
    {
      id: "aeo_clean_url",
      label: "Clean lowercase URL structure",
      pass: !parsedUrl.pathname.includes("_") && parsedUrl.pathname === parsedUrl.pathname.toLowerCase(),
    },
    { id: "aeo_ai_robots", label: "No AI-blocking directives detected", pass: !robotsMeta.includes("noindex") },
    ...aeoParityChecks(rawHtml, visibleText, wordCount, hasSpeakable),
    ...aeoStructureChecks(ctx),
  ];

  // --- GEO Checks ---
  const geoChecks: Check[] = [
    {
      id: "geo_statistics_density",
      label: `${statsMatches.length} quantitative data points / metrics found`,
      pass: statsMatches.length >= 3,
    },
    { id: "geo_data_viz", label: "Structured data tables or charts found", pass: hasTable },
    {
      id: "geo_comparison",
      label: "Comparative sections or tables detected",
      pass: hasTable || /\bvs\b/i.test(lowerText),
    },
    { id: "geo_expert_quotes", label: "Expert quotes or blockquotes found", pass: hasBlockquote },
    {
      id: "geo_citations",
      label: `${externalLinks.length} external source citations found`,
      pass: externalLinks.length > 0,
    },
    {
      id: "geo_definitions",
      label: "Clear concept definitions found ('X is a...')",
      pass: /\bis (a|the|defined as)\b/i.test(lowerText),
    },
    { id: "geo_author_schema", label: "Author schema markup found", pass: hasPerson },
    {
      id: "geo_author_credentials",
      label: "Visible author byline with credentials or dates",
      pass: /\bby\s+[a-z]/i.test(lowerText) || hasPerson,
    },
    { id: "geo_org_schema", label: "Organization or Business schema markup declared", pass: hasOrg },
    {
      id: "geo_knowledge_graph",
      label: `${sameAsLinks.length} authoritative sameAs links in schema`,
      pass: sameAsLinks.length > 0,
    },
    {
      id: "geo_trust",
      label: "Trust signals (About and Contact links) present",
      pass: internalLinks.some((l) => /about/i.test(l)) && internalLinks.some((l) => /contact/i.test(l)),
    },
    {
      id: "geo_freshness",
      label: "Content freshness signals (dates/time elements) present",
      pass: hasTime || /\b202[4-9]\b/.test(lowerText),
    },
    { id: "geo_time_element", label: "<time> machine-readable tags present", pass: hasTime },
    { id: "geo_content_length", label: `${wordCount} words content depth (300+ recommended)`, pass: wordCount >= 300 },
    ...geoParityChecks({
      html: rawHtml,
      title,
      firstHeading: rawHtml
        .match(/<h[1-6]\b[^>]*>([\s\S]*?)<\/h[1-6]>/i)?.[1]
        ?.replace(/<[^>]+>/g, "")
        .trim(),
      robotsMeta,
      schemaTypeCount: schemaTypes.length,
    }),
    ...geoSignalChecks(ctx),
  ];

  const calcCat = (checks: Check[]): CategoryResult => {
    const counted = checks.filter((c) => c.applicable !== false);
    const passed = counted.filter((c) => c.pass).length;
    const total = counted.length;
    const score = total > 0 ? Math.round((passed / total) * 100) : 0;
    return { score, passed, total, checks };
  };

  const seo = calcCat(seoChecks);
  const aeo = calcCat(aeoChecks);
  const geo = calcCat(geoChecks);
  return {
    url: targetUrl,
    overall: Math.round(seo.score * 0.5 + aeo.score * 0.25 + geo.score * 0.25),
    seo,
    aeo,
    geo,
    cwv: { score: null, checks: [], passed: 0, total: 0 },
    internalLinks: [...new Set(internalLinks)],
    scannedAt: new Date().toISOString(),
  };
}
