#!/usr/bin/env node

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod/v3";

const API_ENDPOINT = process.env.SEOSCORE_API_URL || "https://seoscore-api.nadimtuhin.workers.dev/api/scan";

interface CategoryPayload {
  score?: number | null;
  checks?: { pass: boolean }[];
}

interface ScanPayload {
  seo?: CategoryPayload;
  aeo?: CategoryPayload;
  geo?: CategoryPayload;
  internalLinks?: string[];
}

interface PageSummary {
  url: string;
  overall: number;
  seo: number;
  aeo: number;
  geo: number;
  failedChecks: number;
}

const server = new McpServer({
  name: "seoscore-mcp",
  version: "1.0.0",
});

function textResult(text: string, isError = false) {
  return isError
    ? { content: [{ type: "text" as const, text }], isError }
    : { content: [{ type: "text" as const, text }] };
}

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

function postScan(url: string, keyphrase: string | undefined): Promise<Response> {
  return fetch(API_ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ url, keyphrase }),
  });
}

function failedCount(page: ScanPayload): number {
  return [
    ...(page.seo?.checks || []).filter((c) => !c.pass),
    ...(page.aeo?.checks || []).filter((c) => !c.pass),
    ...(page.geo?.checks || []).filter((c) => !c.pass),
  ].length;
}

function summarize(url: string, page: ScanPayload): PageSummary {
  const seo = page.seo?.score ?? 0;
  const aeo = page.aeo?.score ?? 0;
  const geo = page.geo?.score ?? 0;
  const overall = Math.round(seo * 0.5 + aeo * 0.25 + geo * 0.25);
  return { url, overall, seo, aeo, geo, failedChecks: failedCount(page) };
}

server.tool(
  "audit_page",
  "Evaluates 80+ factors across Traditional SEO (50%), Answer Engine Optimization (25%), and Generative Engine Optimization (25%). Returns 0-100 scores and prioritized diagnostic checks.",
  {
    url: z.string().describe("The full web URL (e.g. https://example.com) to audit."),
    keyphrase: z
      .string()
      .optional()
      .describe("Optional focus keyphrase to evaluate in title, meta description, and headings."),
  },
  async ({ url, keyphrase }) => {
    try {
      const resp = await postScan(url, keyphrase);
      if (!resp.ok) {
        return textResult(`Audit API answered HTTP ${resp.status}: ${await resp.text()}`, true);
      }
      const data: unknown = await resp.json();
      return textResult(JSON.stringify(data, null, 2));
    } catch (err) {
      return textResult(`Failed to audit ${url}: ${errorMessage(err)}`, true);
    }
  },
);

server.tool(
  "crawl_site",
  "Crawls a website up to a specified page limit and audits each discovered internal page for SEO, AEO, and GEO scores.",
  {
    url: z.string().describe("Root URL to begin crawling."),
    limit: z.number().optional().default(5).describe("Maximum number of internal pages to crawl (1-25)."),
    keyphrase: z.string().optional().describe("Optional focus keyphrase."),
  },
  async ({ url, limit, keyphrase }) => {
    try {
      const startHost = new URL(url).hostname;
      const queue: string[] = [url];
      const visited = new Set<string>();
      const results: PageSummary[] = [];
      const maxPages = limit || 5;

      while (results.length < maxPages) {
        const current = queue.shift();
        if (current === undefined) break;
        if (visited.has(current)) continue;
        visited.add(current);

        const resp = await postScan(current, keyphrase);
        if (!resp.ok) continue;
        const pageData = (await resp.json()) as ScanPayload;
        results.push(summarize(current, pageData));

        for (const link of pageData.internalLinks || []) {
          if (
            URL.canParse(link) &&
            new URL(link).hostname === startHost &&
            !visited.has(link) &&
            queue.length < maxPages * 2
          ) {
            queue.push(link);
          }
        }
      }

      return textResult(JSON.stringify({ rootUrl: url, totalScanned: results.length, pages: results }, null, 2));
    } catch (err) {
      return textResult(`Crawl failed: ${errorMessage(err)}`, true);
    }
  },
);

server.prompt(
  "seo_audit_prompt",
  "Generates an expert executive audit recommendation report prioritizing highest-impact fixes for Google, Perplexity, and ChatGPT.",
  {
    url: z.string().describe("Website URL to audit and plan fixes for"),
  },
  async ({ url }) => ({
    messages: [
      {
        role: "user",
        content: {
          type: "text",
          text: `Please audit ${url} using the 'audit_page' tool and produce:
1. High-level score card (Overall, Traditional SEO, Answer Engine AEO, Generative Engine GEO).
2. Priority 1 (Quick Wins): Immediate HTML tag and metadata changes.
3. Priority 2 (Answer Engines): Content structure, FAQ schema, definition lists, and direct answer formatting for ChatGPT and Perplexity.
4. Priority 3 (Generative Search): Empirical data tables, citations, and author/entity schema markup for Google AI Overviews.`,
        },
      },
    ],
  }),
);

async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("seoscore-mcp server running on stdio");
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
