import { type LearnPost, learnPosts } from "./learn-content";

const AUTHOR = "Nadim Tuhin";
const AUTHOR_URL = "https://github.com/nadimtuhin";
const REPO_URL = "https://github.com/nadimtuhin/findable";
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

const escapeHtml = (text: string): string =>
  text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");

const jsonLd = (data: unknown): string =>
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

function page(input: PageInput): string {
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
<meta property="og:type" content="${input.ogType}">
<meta property="og:site_name" content="${SITE_NAME}">
<meta property="og:title" content="${title}">
<meta property="og:description" content="${description}">
<meta property="og:url" content="${escapeHtml(url)}">
<meta name="twitter:card" content="summary">
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
<footer><p><a href="/">Audit your own page with ${SITE_NAME}</a> &middot; <a href="${REPO_URL}" rel="noopener">Source on GitHub</a></p></footer>
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
    (post) => `<li><a href="/learn/${post.slug}">${escapeHtml(post.title)}</a><br>${escapeHtml(post.description)}</li>`,
  );
  return `<ul>${items.join("")}</ul>`;
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
    ],
    body: `<h1>${escapeHtml(INDEX_TITLE)}</h1>\n<p>${escapeHtml(INDEX_DESCRIPTION)}</p>\n<h2>All guides</h2>\n${postList(learnPosts)}`,
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
<p>The hosted audit waits up to 10 s for a page and fetches at most 5 MB. It reads robots.txt and llms.txt up to 512 KB each.</p>
</section>`;

function related(post: LearnPost): string {
  const others = learnPosts.filter((p) => p.slug !== post.slug);
  return `<section>\n<h2>More guides</h2>\n${postList(others)}\n</section>`;
}

function postStructuredData(post: LearnPost, origin: string): string[] {
  const url = `${origin}/learn/${post.slug}`;
  const personId = `${origin}/#author`;
  const orgId = `${origin}/#organization`;
  const graph = jsonLd({
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "Article",
        headline: post.title,
        description: post.description,
        datePublished: post.datePublished,
        dateModified: post.dateModified,
        mainEntityOfPage: url,
        author: { "@id": personId },
        publisher: { "@id": orgId },
      },
      { "@type": "Person", "@id": personId, name: AUTHOR, url: AUTHOR_URL, sameAs: [AUTHOR_URL] },
      { "@type": "Organization", "@id": orgId, name: SITE_NAME, url: origin, sameAs: [REPO_URL] },
    ],
  });
  if (post.faq.length === 0) return [graph];
  const faq = jsonLd({
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: post.faq.map((f) => ({
      "@type": "Question",
      name: f.q,
      acceptedAnswer: { "@type": "Answer", text: f.a },
    })),
  });
  return [graph, faq];
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
