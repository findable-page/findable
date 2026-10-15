import { AUTHOR, AUTHOR_URL, page, REPO_URL } from "./learn";

interface LegalPage {
  path: string;
  title: string;
  description: string;
  body: string;
}

const UPDATED = "2026-10-15";

const updated = `<p class="byline">Updated <time datetime="${UPDATED}">${UPDATED}</time>.</p>`;

const PAGES: LegalPage[] = [
  {
    path: "/about",
    title: "About findable",
    description: "findable is a free, open-source scanner that checks a page for SEO, AEO and GEO signals.",
    body: `<h1>About findable</h1>
${updated}
<p>findable is a free, open-source scanner. Paste a URL and it scores the page for SEO, AEO and GEO, then lists the fixes worth the most points.</p>
<p>It is built and maintained by <a href="${AUTHOR_URL}" rel="noopener">${AUTHOR}</a>. The code is on <a href="${REPO_URL}" rel="noopener">GitHub</a> under the MIT license. Bugs and ideas go in the <a href="${REPO_URL}/issues" rel="noopener">issue tracker</a>.</p>
<p>New to the terms? Read the <a href="/learn">plain-words guides</a>.</p>`,
  },
  {
    path: "/privacy",
    title: "Privacy at findable",
    description: "What findable fetches, what it keeps, and what it never does.",
    body: `<h1>Privacy</h1>
${updated}
<h2>What we fetch</h2>
<p>When you scan, our server fetches the public URL you submit, plus its robots.txt and llms.txt. Nothing else about you is needed to run a scan.</p>
<h2>What we keep</h2>
<ul>
<li>Scan results are not cached. Share previews are cached for 1 hour, and speed test results and share cards for up to 24 hours.</li>
<li>The optional speed test sends the URL you scan to Google PageSpeed.</li>
<li>Share links carry the score in the link itself. Anyone with the link can see it.</li>
<li>If you press Get permanent link, we re-run the scan and store the URL, host and check results in Cloudflare D1 for 90 days, then the link stops working. Anyone with the link can read the report, and search engines are asked not to index it. To have a report removed earlier, open a <a href="${REPO_URL}/issues" rel="noopener">GitHub issue</a>.</li>
<li>If you opt in to watch a site, we store your email and the site host in Cloudflare D1. We do not send emails yet. To have your entry removed, open a <a href="${REPO_URL}/issues" rel="noopener">GitHub issue</a>.</li>
</ul>
<h2>Logs</h2>
<p>The site runs on Cloudflare, which keeps standard request logs such as IP address and user agent. We use them only to run and protect the service, for example rate limits.</p>
<h2>Ads and trackers</h2>
<p>There are no ads, no analytics scripts and no tracking cookies on findable. Your language choice is kept in your own browser storage and never sent to us.</p>`,
  },
  {
    path: "/terms",
    title: "Terms of use for findable",
    description: "findable is provided as is, with no warranty. Scan only sites you are allowed to scan.",
    body: `<h1>Terms of use</h1>
${updated}
<ul>
<li>findable is provided as is, with no warranty of any kind. Scores and suggestions are guidance, not a promise about rankings or citations.</li>
<li>Scan only sites you own or have permission to scan.</li>
<li>Usage is subject to rate limits. We may slow or block requests that abuse the service.</li>
<li>The source code is MIT licensed. See the <a href="${REPO_URL}" rel="noopener">repository</a>.</li>
</ul>
<p>Questions go to the <a href="${REPO_URL}/issues" rel="noopener">issue tracker</a>.</p>`,
  },
];

export const LEGAL_PAGES = PAGES.map(({ path, title, description }) => ({ path, title, description }));

export function renderLegalPage(path: string, origin: string): string | null {
  const found = PAGES.find((p) => p.path === path.replace(/(.)\/$/, "$1"));
  if (!found) return null;
  return page({
    title: found.title,
    description: found.description,
    path: found.path,
    origin,
    ogType: "website",
    structuredData: [],
    body: found.body,
  });
}
