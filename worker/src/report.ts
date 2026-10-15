import { escapeHtml, page } from "./learn";

const BASE62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
const ID_LENGTH = 12;
const REPORT_TTL_MS = 90 * 24 * 60 * 60 * 1000;
export const REPORT_ID = /^[0-9A-Za-z]{12}$/;

interface ReportCheck {
  id: string;
  label: string;
  pass: boolean;
  applicable?: boolean | null;
}

export interface StoredReport {
  url: string;
  scannedAt: string;
  overall: number;
  seo: { score: number; checks: ReportCheck[] };
  aeo: { score: number; checks: ReportCheck[] };
  geo: { score: number; checks: ReportCheck[] };
}

export class ReportTooLargeError extends Error {}

const MAX_LABEL = 200;
const MAX_JSON = 64_000;

function clamp(group: StoredReport["seo"]): StoredReport["seo"] {
  return { score: group.score, checks: group.checks.map((c) => ({ ...c, label: c.label.slice(0, MAX_LABEL) })) };
}

export async function deleteExpiredReports(db: D1Database, now = new Date()): Promise<void> {
  await db.prepare("DELETE FROM reports WHERE expires_at <= ?").bind(now.toISOString()).run();
}

export function newReportId(): string {
  const id: string[] = [];
  while (id.length < ID_LENGTH) {
    for (const byte of crypto.getRandomValues(new Uint8Array(ID_LENGTH))) {
      if (byte < 248 && id.length < ID_LENGTH) id.push(BASE62[byte % 62] as string);
    }
  }
  return id.join("");
}

export async function saveReport(db: D1Database, scan: StoredReport, now = new Date()): Promise<string> {
  const id = newReportId();
  const stored: StoredReport = {
    url: scan.url,
    scannedAt: scan.scannedAt,
    overall: scan.overall,
    seo: clamp(scan.seo),
    aeo: clamp(scan.aeo),
    geo: clamp(scan.geo),
  };
  const json = JSON.stringify(stored);
  if (json.length > MAX_JSON) throw new ReportTooLargeError("report too large");
  await deleteExpiredReports(db, now).catch(() => undefined);
  await db
    .prepare("INSERT INTO reports (id, host, url, created_at, expires_at, json) VALUES (?, ?, ?, ?, ?, ?)")
    .bind(
      id,
      new URL(scan.url).hostname,
      scan.url,
      now.toISOString(),
      new Date(now.getTime() + REPORT_TTL_MS).toISOString(),
      json,
    )
    .run();
  return id;
}

export async function loadReport(db: D1Database, id: string, now = new Date()): Promise<StoredReport | null> {
  const row = await db
    .prepare("SELECT json, expires_at FROM reports WHERE id = ?")
    .bind(id)
    .first<{ json: string; expires_at: string }>();
  if (!row || row.expires_at <= now.toISOString()) return null;
  return JSON.parse(row.json) as StoredReport;
}

// The English fix advice lives in the UI script; reading it here keeps one copy.
export function parseFixAdvice(uiHtml: string): Map<string, string> {
  const advice = new Map<string, string>();
  const block = uiHtml.match(/const fixAdvice = \{([\s\S]*?)\n\s*\};/)?.[1] ?? "";
  for (const m of block.matchAll(/^\s*(\w+): ("(?:[^"\\]|\\.)*"),?$/gm)) {
    advice.set(m[1] as string, JSON.parse(m[2] as string) as string);
  }
  return advice;
}

const DEFAULT_ADVICE = "Optimize this signal according to current search and AI answer guidelines.";

export function reportCardUrl(origin: string, report: StoredReport): string {
  const params = new URLSearchParams({
    h: new URL(report.url).hostname,
    o: String(report.overall),
    s: String(report.seo.score),
    a: String(report.aeo.score),
    g: String(report.geo.score),
  });
  return `${origin}/og/card.png?${params}`;
}

export function renderReport(id: string, report: StoredReport, origin: string, advice: Map<string, string>): string {
  const host = new URL(report.url).hostname;
  const failed = (["seo", "aeo", "geo"] as const).flatMap((family) =>
    report[family].checks
      .filter((c) => !c.pass && c.applicable !== false)
      .map((c) => ({ ...c, family: family.toUpperCase() })),
  );
  const rows = failed
    .map(
      (c) =>
        `<li><strong>${escapeHtml(c.label)}</strong> <span class="byline">(${c.family})</span><br>${escapeHtml(advice.get(c.id) ?? DEFAULT_ADVICE)}</li>`,
    )
    .join("\n");
  const html = page({
    title: `${host} scored ${report.overall}/100 | findable report`,
    description: `findable report for ${host}: SEO ${report.seo.score}, AEO ${report.aeo.score}, GEO ${report.geo.score}.`,
    path: `/r/${id}`,
    origin,
    ogType: "website",
    structuredData: ['<meta name="robots" content="noindex">'],
    body: `<h1>${escapeHtml(host)} scored ${report.overall}/100</h1>
<p class="byline">Scanned ${escapeHtml(report.scannedAt)} &middot; <a href="${escapeHtml(report.url)}" rel="noopener nofollow ugc">${escapeHtml(report.url)}</a></p>
<table><tr><th>SEO</th><th>AEO</th><th>GEO</th></tr><tr><td>${report.seo.score}</td><td>${report.aeo.score}</td><td>${report.geo.score}</td></tr></table>
<h2>${failed.length === 0 ? "Every applicable check passes" : `Fix these ${failed.length} checks`}</h2>
<ol>
${rows}
</ol>
<p><a href="/${escapeHtml(host)}">Scan ${escapeHtml(host)} again</a>. This report is kept for 90 days.</p>`,
  });
  const card = escapeHtml(reportCardUrl(origin, report));
  return html
    .replaceAll(`${escapeHtml(origin)}/og.jpg`, card)
    .replace(
      '<meta name="twitter:card" content="summary">',
      '<meta name="twitter:card" content="summary_large_image">',
    );
}
