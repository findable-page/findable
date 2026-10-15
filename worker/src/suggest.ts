export const SUGGEST_MODEL = "@cf/meta/llama-3.1-8b-instruct-fp8";
const MAX_TITLE = 60;
const MAX_DESCRIPTION = 155;
const MAX_PAGE_TEXT = 2000;
const FAQ_COUNT = 3;
const MAX_FIELD = 300;
const MAX_HTML = 100 * 1024;
const MAX_FAQ_Q = 120;
const MAX_FAQ_A = 300;
export const MAX_KEYPHRASE = 100;
export const SUGGEST_MAX_TOKENS = 600;

export interface Suggestion {
  title: string;
  description: string;
  faq: { q: string; a: string }[];
}

function stripTags(html: string): string {
  return html
    .replace(/<[^>]*>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

function tagText(html: string, pattern: RegExp): string {
  return stripTags(pattern.exec(html)?.[1] ?? "").slice(0, MAX_FIELD);
}

function metaDescription(html: string): string {
  const tag = /<meta\b[^>]*\bname=["']description["'][^>]*>/i.exec(html)?.[0] ?? "";
  return /\bcontent=["']([^"']*)["']/i.exec(tag)?.[1]?.trim().slice(0, MAX_FIELD) ?? "";
}

export function suggestPrompt(fullHtml: string, keyphrase?: string): { role: string; content: string }[] {
  const html = fullHtml.slice(0, MAX_HTML);
  const body = html.replace(/<(script|style|noscript)\b[\s\S]*?<\/\1>/gi, " ").replace(/<head\b[\s\S]*?<\/head>/i, " ");
  const page = [
    `Current title: ${tagText(html, /<title[^>]*>([\s\S]*?)<\/title>/i)}`,
    `Current meta description: ${metaDescription(html)}`,
    `H1: ${tagText(html, /<h1[^>]*>([\s\S]*?)<\/h1>/i)}`,
    `Page text: ${stripTags(body).slice(0, MAX_PAGE_TEXT)}`,
  ].join("\n");
  const focus = keyphrase ? ` Work the focus keyphrase "${keyphrase}" in naturally.` : "";
  return [
    {
      role: "system",
      content: `You write SEO copy. Reply with only JSON: {"title": string (max ${MAX_TITLE} chars), "description": string (max ${MAX_DESCRIPTION} chars), "faq": [{"q": string, "a": string}] (exactly ${FAQ_COUNT} items)}.${focus}`,
    },
    { role: "user", content: page },
  ];
}

export function parseSuggestion(reply: string): Suggestion | null {
  const start = reply.indexOf("{");
  const end = reply.lastIndexOf("}");
  if (start < 0 || end < start) return null;
  let data: unknown;
  try {
    data = JSON.parse(reply.slice(start, end + 1));
  } catch {
    return null;
  }
  const { title, description, faq } = (data ?? {}) as Record<string, unknown>;
  if (typeof title !== "string" || typeof description !== "string" || !Array.isArray(faq)) return null;
  const items = faq
    .filter((f): f is { q: string; a: string } => typeof f?.q === "string" && typeof f?.a === "string")
    .slice(0, FAQ_COUNT);
  if (!title.trim() || !description.trim() || items.length === 0) return null;
  return {
    title: title.trim().slice(0, MAX_TITLE),
    description: description.trim().slice(0, MAX_DESCRIPTION),
    faq: items.map((f) => ({ q: f.q.trim().slice(0, MAX_FAQ_Q), a: f.a.trim().slice(0, MAX_FAQ_A) })),
  };
}
