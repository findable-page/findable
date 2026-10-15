export interface PageEntry {
  path: string;
  title: string;
  description: string;
}

export interface FaqItem {
  question: string;
  answer: string;
}

export const HEAD_SLOT_MARKER = "<!--slot:head-->";

const AI_BOTS = ["GPTBot", "ClaudeBot", "PerplexityBot", "Google-Extended"];
const REPO = "https://github.com/nadimtuhin/findable";
const AUTHOR = { name: "Nadim Tuhin", url: "https://github.com/nadimtuhin" };
const TAGLINE = "findable: is your site findable by Google and by AI answer engines?";

const SUMMARY =
  "Open-source auditor that scores a page for traditional SEO, Answer Engine Optimization (AEO) and Generative Engine Optimization (GEO). " +
  "The web API and the Go CLI both run 160 page checks (98 SEO, 35 AEO, 27 GEO), plus 7 separate site checks (robots.txt, AI crawler access, llms.txt).";

const HOW_TO_USE = [
  "- CLI: `go install github.com/nadimtuhin/findable/cmd/findable@latest`, then `findable https://example.com`.",
  "- MCP: `findable -mcp`, or `npx -y findable-mcp`. Tools: audit_page, audit_html_content, crawl_site, generate_csv_report.",
  "- API: `POST /api/scan` with JSON body `{url, keyphrase?, site?}`; `site: true` adds the site checks.",
  `- Source and license (MIT): ${REPO}`,
];

export function robotsTxt(origin: string): string {
  return [
    ...["*", ...AI_BOTS].map((agent) => `User-agent: ${agent}`),
    "Allow: /",
    "Disallow: /api/",
    "",
    `Sitemap: ${origin}/sitemap.xml`,
    "",
  ].join("\n");
}

function escapeXml(text: string): string {
  return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

export function sitemapXml(origin: string, paths: string[]): string {
  const urls = paths.map((p) => `  <url><loc>${escapeXml(origin + p)}</loc></url>`).join("\n");
  return `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls}\n</urlset>\n`;
}

export function llmsTxt(origin: string, pages: PageEntry[]): string {
  const links = pages.map((p) => `- [${p.title}](${origin}${p.path}): ${p.description}`);
  return [
    "# findable",
    "",
    `> ${SUMMARY}`,
    "",
    TAGLINE,
    "",
    "## Pages",
    "",
    ...links,
    "",
    "## Optional",
    "",
    `- [Source code](${REPO}): Go CLI, MCP server and Cloudflare Worker`,
    `- [Full text](${origin}/llms-full.txt): every page described in full`,
    "",
  ].join("\n");
}

export function llmsFullTxt(origin: string, pages: PageEntry[]): string {
  const sections = pages.map((p) => `### ${p.title}\n\n${origin}${p.path}\n\n${p.description}\n`);
  return [
    "# findable",
    "",
    `> ${SUMMARY}`,
    "",
    TAGLINE,
    "",
    "## How to use it",
    "",
    ...HOW_TO_USE,
    "",
    "## Pages",
    "",
    ...sections,
  ].join("\n");
}

function jsonLdScript(data: object): string {
  const json = JSON.stringify(data).replace(/</g, "\\u003c");
  return `<script type="application/ld+json">${json}</script>`;
}

export function headSlot(origin: string, faq: FaqItem[] = []): string {
  const context = "https://schema.org";
  const blocks: object[] = [
    {
      "@context": context,
      "@type": "SoftwareApplication",
      name: "findable",
      applicationCategory: "DeveloperApplication",
      operatingSystem: "Any",
      url: origin,
      offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
      sameAs: [REPO],
      author: { "@type": "Person", name: AUTHOR.name, url: AUTHOR.url },
    },
    { "@context": context, "@type": "Person", name: AUTHOR.name, url: AUTHOR.url, sameAs: [AUTHOR.url] },
    {
      "@context": context,
      "@type": "Organization",
      name: "findable",
      url: origin,
      sameAs: [REPO],
      founder: { "@type": "Person", name: AUTHOR.name, url: AUTHOR.url },
    },
  ];
  if (faq.length > 0) {
    blocks.push({
      "@context": context,
      "@type": "FAQPage",
      mainEntity: faq.map((f) => ({
        "@type": "Question",
        name: f.question,
        acceptedAnswer: { "@type": "Answer", text: f.answer },
      })),
    });
  }
  return blocks.map(jsonLdScript).join("\n");
}

export function injectHead(html: string, slot: string): string {
  return html.replace(HEAD_SLOT_MARKER, () => slot);
}
