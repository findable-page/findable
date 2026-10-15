import { FETCH_TIMEOUT_MS, MAX_PAGE_BYTES } from "./guard";
import { type LearnPost, learnPosts } from "./learn-content";
import { MAX_SITE_FILE_BYTES } from "./site";

export const AUTHOR = "Nadim Tuhin";
export const AUTHOR_URL = "https://github.com/nadimtuhin";
export const REPO_URL = "https://github.com/nadimtuhin/findable";
const SITE_NAME = "findable";
const INDEX_TITLE = "Learn SEO, AEO and GEO with findable";
const INDEX_DESCRIPTION =
  "Plain guides to SEO, AEO and GEO and to every check findable runs, so you can see why a page is hard to find and fix it.";

export const LEARN_STYLE =
  "body{font:16px/1.6 system-ui,sans-serif;max-width:46rem;margin:0 auto;padding:1rem;color:#1a1a1a;background:#fff}" +
  "a{color:#0b5cad}nav,footer{font-size:.9rem}nav a{margin-right:1rem}h1{line-height:1.2}" +
  "table{border-collapse:collapse;display:block;overflow-x:auto}th,td{border:1px solid #ccc;padding:.3rem .5rem;text-align:left}" +
  "pre{overflow-x:auto;background:#f4f4f4;padding:.75rem}code{background:#f4f4f4;padding:0 .2rem}pre code{padding:0}" +
  ".byline{color:#555}details{margin:1rem 0}dt{font-weight:600;margin-top:1rem}dd{margin:.25rem 0 0}";

export const escapeHtml = (text: string): string =>
  text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");

export const jsonLd = (data: unknown): string =>
  `<script type="application/ld+json">${JSON.stringify(data).replace(/</g, "\\u003c")}</script>`;

export function learnPaths(): string[] {
  return ["/learn", ...learnPosts.map((post) => `/learn/${post.slug}`)];
}

interface PageInput {
  title: string;
  description: string;
  path: string;
  origin: string;
  ogType: "website" | "article";
  structuredData: string[];
  body: string;
}

export function page(input: PageInput): string {
  const url = input.origin + input.path;
  const title = escapeHtml(input.title);
  const description = escapeHtml(input.description);
  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>${title}</title>
<meta name="description" content="${description}">
<link rel="canonical" href="${escapeHtml(url)}">
<link rel="icon" href="/favicon.svg" type="image/svg+xml">
<link rel="icon" href="/favicon.ico" sizes="any">
<link rel="apple-touch-icon" href="/apple-touch-icon.png">
<link rel="manifest" href="/site.webmanifest">
<meta name="theme-color" content="#6d5efc">
<meta property="og:type" content="${input.ogType}">
<meta property="og:site_name" content="${SITE_NAME}">
<meta property="og:title" content="${title}">
<meta property="og:description" content="${description}">
<meta property="og:url" content="${escapeHtml(url)}">
<meta property="og:image" content="${escapeHtml(input.origin)}/og.jpg">
<meta name="twitter:card" content="summary">
<meta name="twitter:image" content="${escapeHtml(input.origin)}/og.jpg">
<meta name="twitter:title" content="${title}">
<meta name="twitter:description" content="${description}">
<style>${LEARN_STYLE}</style>
${input.structuredData.join("\n")}
</head>
<body>
<nav><a href="/">${SITE_NAME}</a><a href="/learn">Learn</a></nav>
<main>
${input.body}
</main>
<footer><p><a href="/">Audit your own page with ${SITE_NAME}</a> &middot; <a href="/compare">Compare</a> &middot; <a href="/about">About</a> &middot; <a href="/privacy">Privacy</a> &middot; <a href="/terms">Terms</a> &middot; <a href="${REPO_URL}" rel="noopener">Source on GitHub</a></p></footer>
</body>
</html>
`;
}

export function renderLearnNotFound(origin: string): string {
  return page({
    title: "Page not found | findable",
    description: "That learn page does not exist. Go back to the findable guides or run an audit on the home page.",
    path: "/learn",
    origin,
    ogType: "website",
    structuredData: [],
    body: `<h1>Page not found</h1>\n<p>There is no guide at this address. See <a href="/learn">all findable guides</a> or <a href="/">audit a page</a>.</p>`,
  });
}

function postList(posts: LearnPost[]): string {
  const items = posts.map(
    (post) =>
      `<dt><a href="/learn/${post.slug}">${escapeHtml(post.title)}</a></dt><dd>${escapeHtml(post.description)}</dd>`,
  );
  return `<dl>${items.join("")}</dl>`;
}

const INDEX_FAQ: LearnPost["faq"] = [
  {
    q: "What is findable?",
    a: "findable checks whether a page can be found by Google and by AI answer engines. It scores SEO, AEO and GEO and lists every failing check by id so you know what to fix.",
  },
  {
    q: "Where should I start?",
    a: "Read the comparison of SEO, AEO and GEO first, then run an audit on your own page and open the check guide for each failing id.",
  },
  {
    q: "Who writes these guides?",
    a: `${AUTHOR} writes them. ${AUTHOR} builds findable, so every guide describes checks that exist in the code.`,
  },
];

function latestModified(): string {
  return learnPosts.map((p) => p.dateModified).sort()[learnPosts.length - 1] ?? "";
}

function indexBody(): string {
  const updated = latestModified();
  const start = ["seo-vs-aeo-vs-geo", "findable-in-15-minutes", "why-ai-answer-engines-skip-your-site"];
  const steps = start
    .map((slug) => learnPosts.find((p) => p.slug === slug))
    .filter((p): p is LearnPost => p !== undefined)
    .map((p) => `<li><a href="/learn/${p.slug}">${escapeHtml(p.title)}</a></li>`);
  return [
    `<h1>${escapeHtml(INDEX_TITLE)}</h1>`,
    `<p class="byline">By <a href="${AUTHOR_URL}" rel="noopener">${AUTHOR}</a>. Updated <time datetime="${updated}">${updated}</time>.</p>`,
    `<p>${escapeHtml(INDEX_DESCRIPTION)} findable is a tool that audits one page for SEO, AEO and GEO. You can read the guides in any order, and each one links to the others.</p>`,
    `<p>On this page: <a href="#start-here">start here</a>, <a href="#all-guides">all guides</a>, <a href="#quick-start">command line</a>, <a href="#summary">summary</a>.</p>`,
    `<h2 id="start-here">Where should you start?</h2>\n<ol>${steps.join("")}</ol>`,
    `<h2 id="all-guides">All guides</h2>\n${postList(learnPosts)}`,
    `<h2 id="quick-start">How do you run an audit from the command line?</h2>\n<p>Pass a URL to the seoscore command. The <code>-fail-under</code> flag fails the run when the score drops below a number, which is useful in CI.</p>\n<pre><code>seoscore -fail-under 80 https://example.com</code></pre>`,
    SCORING_NOTE,
    `<h2 id="summary">Summary</h2>\n<ul><li>SEO is how crawlers find and index a page.</li><li>AEO is how easily an answer can be lifted from it.</li><li>GEO is how likely a generative model is to cite it.</li><li>Every check has a stable id, so you can look up a failure and fix it.</li></ul>`,
    faqSection(INDEX_FAQ)
      .replace("<dl>", "<details open><summary>Questions and answers</summary><dl>")
      .replace("</dl>", "</dl></details>"),
  ].join("\n");
}

export function renderLearnIndex(origin: string): string {
  return page({
    title: INDEX_TITLE,
    description: INDEX_DESCRIPTION,
    path: "/learn",
    origin,
    ogType: "website",
    structuredData: [
      jsonLd({
        "@context": "https://schema.org",
        "@type": "CollectionPage",
        name: INDEX_TITLE,
        description: INDEX_DESCRIPTION,
        url: `${origin}/learn`,
      }),
      identityGraph(origin),
      faqLd(INDEX_FAQ),
    ],
    body: indexBody(),
  });
}

function tableOfContents(html: string): string {
  const entries = [...html.matchAll(/<h2 id="([^"]+)">([\s\S]*?)<\/h2>/g)];
  if (entries.length === 0) return "";
  const items = entries.map((m) => `<li><a href="#${m[1]}">${m[2]}</a></li>`);
  return `<details><summary>Contents</summary><ol>${items.join("")}</ol></details>`;
}

function faqSection(faq: LearnPost["faq"]): string {
  if (faq.length === 0) return "";
  const entries = faq.map((f) => `<dt>${escapeHtml(f.q)}</dt><dd>${escapeHtml(f.a)}</dd>`);
  return `<section>\n<h2 id="faq">FAQ</h2>\n<dl>${entries.join("")}</dl>\n</section>`;
}

function byline(post: LearnPost): string {
  const published = `<time datetime="${post.datePublished}">${post.datePublished}</time>`;
  const modified = `<time datetime="${post.dateModified}">${post.dateModified}</time>`;
  return `<p class="byline">By <a href="${AUTHOR_URL}" rel="noopener">${AUTHOR}</a>. Published ${published}. Updated ${modified}.</p>`;
}

const SCORING_NOTE = `<section>
<h2 id="how-findable-scores">How does findable score a page?</h2>
<table><thead><tr><th>Family</th><th>Weight</th><th>Question it asks</th></tr></thead><tbody>
<tr><td>SEO</td><td>50%</td><td>Can crawlers read and index this page?</td></tr>
<tr><td>AEO</td><td>25%</td><td>Is it easy to lift an answer from this page?</td></tr>
<tr><td>GEO</td><td>25%</td><td>Does the page give a generative model a reason to cite it?</td></tr>
</tbody></table>
<p>The hosted audit waits up to ${FETCH_TIMEOUT_MS / 1000} s for a page and fetches at most ${MAX_PAGE_BYTES / (1024 * 1024)} MB. It reads robots.txt and llms.txt up to ${MAX_SITE_FILE_BYTES / 1024} KB each.</p>
</section>`;

function related(post: LearnPost): string {
  const others = learnPosts.filter((p) => p.slug !== post.slug);
  return `<section>\n<h2>More guides</h2>\n${postList(others)}\n</section>`;
}

const personId = (origin: string) => `${origin}/#author`;
const orgId = (origin: string) => `${origin}/#organization`;

const identityEntities = (origin: string) => [
  { "@type": "Person", "@id": personId(origin), name: AUTHOR, url: AUTHOR_URL, sameAs: [AUTHOR_URL] },
  { "@type": "Organization", "@id": orgId(origin), name: SITE_NAME, url: origin, sameAs: [REPO_URL] },
];

export const identityGraph = (origin: string): string =>
  jsonLd({ "@context": "https://schema.org", "@graph": identityEntities(origin) });

export const faqLd = (faq: LearnPost["faq"]): string =>
  jsonLd({
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: faq.map((f) => ({
      "@type": "Question",
      name: f.q,
      acceptedAnswer: { "@type": "Answer", text: f.a },
    })),
  });

function postStructuredData(post: LearnPost, origin: string): string[] {
  const graph = jsonLd({
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "Article",
        headline: post.title,
        description: post.description,
        datePublished: post.datePublished,
        dateModified: post.dateModified,
        mainEntityOfPage: `${origin}/learn/${post.slug}`,
        author: { "@id": personId(origin) },
        publisher: { "@id": orgId(origin) },
      },
      ...identityEntities(origin),
    ],
  });
  return post.faq.length === 0 ? [graph] : [graph, faqLd(post.faq)];
}

export function renderLearnPost(slug: string, origin: string): string | null {
  const post = learnPosts.find((p) => p.slug === slug);
  if (!post) return null;
  const [heading = "", ...rest] = post.html.split("\n");
  const body = [
    heading,
    byline(post),
    tableOfContents(post.html),
    rest.join("\n"),
    faqSection(post.faq),
    SCORING_NOTE,
    related(post),
  ];
  return page({
    title: post.title,
    description: post.description,
    path: `/learn/${post.slug}`,
    origin,
    ogType: "article",
    structuredData: postStructuredData(post, origin),
    body: `<article>\n${body.filter(Boolean).join("\n")}\n</article>`,
  });
}
