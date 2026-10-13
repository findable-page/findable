#!/usr/bin/env node

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";

const API_ENDPOINT = process.env.SEOSCORE_API_URL || "https://seoscore-api.nadimtuhin.workers.dev/api/scan";

const server = new McpServer({
  name: "seoscore-mcp",
  version: "1.0.0",
});

const serverAny = server as any;

serverAny.tool(
  "audit_page",
  "Evaluates 80+ factors across Traditional SEO (50%), Answer Engine Optimization (25%), and Generative Engine Optimization (25%). Returns 0-100 scores and prioritized diagnostic checks.",
  {
    url: z.string().describe("The full web URL (e.g. https://example.com) to audit."),
    keyphrase: z.string().optional().describe("Optional focus keyphrase to evaluate in title, meta description, and headings."),
  },
  async ({ url, keyphrase }: { url: string; keyphrase?: string }) => {
    try {
      const resp = await fetch(API_ENDPOINT, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url, keyphrase }),
      });

      if (!resp.ok) {
        return {
          content: [{ type: "text", text: `Audit API answered HTTP ${resp.status}: ${await resp.text()}` }],
          isError: true,
        };
      }

      const data = await resp.json();
      return {
        content: [{ type: "text", text: JSON.stringify(data, null, 2) }],
      };
    } catch (err: any) {
      return {
        content: [{ type: "text", text: `Failed to audit ${url}: ${err.message}` }],
        isError: true,
      };
    }
  }
);

serverAny.tool(
  "crawl_site",
  "Crawls a website up to a specified page limit and audits each discovered internal page for SEO, AEO, and GEO scores.",
  {
    url: z.string().describe("Root URL to begin crawling."),
    limit: z.number().optional().default(5).describe("Maximum number of internal pages to crawl (1-25)."),
    keyphrase: z.string().optional().describe("Optional focus keyphrase."),
  },
  async ({ url, limit, keyphrase }: { url: string; limit?: number; keyphrase?: string }) => {
    try {
      const startHost = new URL(url).hostname;
      const queue: string[] = [url];
      const visited = new Set<string>();
      const results: any[] = [];
      const maxPages = limit || 5;

      while (queue.length > 0 && results.length < maxPages) {
        const current = queue.shift()!;
        if (visited.has(current)) continue;
        visited.add(current);

        const resp = await fetch(API_ENDPOINT, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ url: current, keyphrase }),
        });

        if (!resp.ok) continue;
        const pageData = (await resp.json()) as any;
        const seo = pageData.seo?.score ?? 0;
        const aeo = pageData.aeo?.score ?? 0;
        const geo = pageData.geo?.score ?? 0;
        const overall = Math.round(seo * 0.5 + aeo * 0.25 + geo * 0.25);

        results.push({
          url: current,
          overall,
          seo,
          aeo,
          geo,
          failedChecks: [
            ...(pageData.seo?.checks || []).filter((c: any) => !c.pass),
            ...(pageData.aeo?.checks || []).filter((c: any) => !c.pass),
            ...(pageData.geo?.checks || []).filter((c: any) => !c.pass),
          ].length,
        });

        for (const link of pageData.internalLinks || []) {
          try {
            const parsed = new URL(link);
            if (parsed.hostname === startHost && !visited.has(link) && queue.length < maxPages * 2) {
              queue.push(link);
            }
          } catch {}
        }
      }

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify({ rootUrl: url, totalScanned: results.length, pages: results }, null, 2),
          },
        ],
      };
    } catch (err: any) {
      return {
        content: [{ type: "text", text: `Crawl failed: ${err.message}` }],
        isError: true,
      };
    }
  }
);

serverAny.prompt(
  "seo_audit_prompt",
  "Generates an expert executive audit recommendation report prioritizing highest-impact fixes for Google, Perplexity, and ChatGPT.",
  {
    url: z.string().describe("Website URL to audit and plan fixes for"),
  },
  async ({ url }: { url: string }) => ({
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
  })
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
