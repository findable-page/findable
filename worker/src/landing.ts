import { AUTHOR, AUTHOR_URL, escapeHtml, faqLd, identityGraph, jsonLd, page } from "./learn";
import type { LearnFaq } from "./learn-content";
import type { PageEntry } from "./site-files";

interface LandingPage extends PageEntry {
  body: string;
  faq: LearnFaq[];
}

const UPDATED = "2026-10-15";

const CTA = `<p><a href="/"><strong>Scan a page free</strong></a>. Paste a URL, no signup.</p>`;

const COUNTS = "160 checks: 98 SEO, 35 AEO and 27 GEO";

const LANDING_PAGES: LandingPage[] = [
  {
    path: "/free-seo-scanner",
    title: "Free SEO scanner for SEO, AEO and GEO | findable",
    description:
      "Paste a URL and get a 0 to 100 score from 160 checks (98 SEO, 35 AEO, 27 GEO), plus the fixes worth the most points. Free, no signup.",
    body: `<h1>Free SEO scanner for any page</h1>
<p>findable fetches one public URL and runs ${COUNTS}. You get an overall score from 0 to 100, a score for each family and a list of failing checks ranked by the points they would recover.</p>
${CTA}
<h2 id="what-it-checks">What does it check?</h2>
<ul>
<li><strong>SEO (98 checks, 50% of the score)</strong>: title, meta description, canonical, headings, links, images, structured data and other signals a search crawler reads.</li>
<li><strong>AEO (35 checks, 25%)</strong>: whether an answer can be lifted from the page, such as FAQ markup, short direct answers, lists and tables.</li>
<li><strong>GEO (27 checks, 25%)</strong>: whether a generative model has facts, authorship, dates and sources to cite.</li>
</ul>
<h2 id="grades">What do the grades mean?</h2>
<p>The overall score maps to a grade: 90 to 100 is A+ Excellent, 80 to 89 is A Good, 70 to 79 is B Fair, 60 to 69 is C Needs work and below 60 is Critical issues.</p>
<h2 id="free">Is it really free?</h2>
<p>Yes. There is no account, no paywall and no trial. The code is MIT licensed on <a href="https://github.com/nadimtuhin/findable" rel="noopener">GitHub</a>. Scans are rate limited to keep the service available for everyone.</p>
<h2 id="limits">What are the limits?</h2>
<p>It scores one page at a time, so a good score on one URL says nothing about the rest of your site. It checks signals on the page. It doesn't ask Google Search or any AI product how it ranks or cites you, and a high score is not a promise about rankings or citations.</p>
<h2 id="more">Other ways to run it</h2>
<p>The same checks run in a <a href="/learn/findable-in-15-minutes">command line tool</a>, as an MCP server for AI agents and in <a href="/wordpress">WordPress</a>. See <a href="/learn">the guides</a> for what each check means.</p>`,
    faq: [
      {
        q: "Is the SEO scanner free?",
        a: "Yes. It costs nothing, needs no signup and the source code is MIT licensed.",
      },
      {
        q: "How many checks does it run?",
        a: "160 page checks: 98 for SEO, 35 for AEO and 27 for GEO. The hosted scan can also read robots.txt and llms.txt as separate site checks.",
      },
      {
        q: "What limits does a scan have?",
        a: "A scan reads up to 5 MB of a page and gives up after 10 seconds. It fetches public hosts on ports 80 and 443 only, and is limited to 10 scans per minute per IP address.",
      },
      {
        q: "Does it check robots.txt and AI crawlers?",
        a: "Yes, with the site option. It fetches robots.txt and llms.txt and reports whether GPTBot, ClaudeBot, PerplexityBot and Google-Extended are blocked. These are shown apart from the score. The optional speed test sends the URL to Google PageSpeed.",
      },
      {
        q: "Does it guarantee better rankings?",
        a: "No. The score is guidance about signals on one page, not a promise about rankings or AI citations.",
      },
    ],
  },
  {
    path: "/aeo",
    title: "AEO checker for answer engines | findable",
    description:
      "Answer engine optimization explained plainly, with a free checker that runs 35 AEO checks on any URL and lists what to fix. No signup.",
    body: `<h1>AEO checker for answer engines</h1>
<p>AEO means answer engine optimization. It is about shaping a page so an answer box or an AI chat tool can lift a short, correct answer from it. findable runs 35 AEO checks on a page, and AEO makes up 25% of the overall score.</p>
${CTA}
<h2 id="what-is-aeo">What is AEO?</h2>
<p>SEO gets a page found and ranked by a crawler. AEO gets a passage of it chosen as the answer. GEO, the third family, gives a generative model facts and sources it can cite. Read <a href="/learn/seo-vs-aeo-vs-geo">SEO vs AEO vs GEO</a> for the longer version.</p>
<h2 id="what-it-checks">What do the AEO checks look at?</h2>
<ul>
<li>FAQ schema or a visible Q&amp;A section (<code>aeo_faq</code>)</li>
<li>Concise answer paragraphs of 30 to 80 words (<code>aeo_direct_answers</code>)</li>
<li>Lists, numbered steps, tables and definition lists that are easy to lift</li>
<li>Structured data a machine can read (<code>aeo_structured_data</code>)</li>
<li>No directives that block AI crawlers (<code>aeo_ai_robots</code>)</li>
</ul>
<p>Every check has a stable id. The <a href="/learn/aeo-checks">AEO check guide</a> explains each one and how to fix it.</p>
<h2 id="limits">What it cannot tell you</h2>
<p>findable doesn't ask ChatGPT, Perplexity, Claude or Google AI about your page. It checks the signals those systems can use. Passing the checks makes a page easier to use as an answer and does not guarantee it will be chosen.</p>`,
    faq: [
      {
        q: "What is AEO?",
        a: "Answer engine optimization. It shapes a page so an answer box or chat tool can lift a short, correct answer.",
      },
      {
        q: "How many AEO checks does findable run?",
        a: "35 AEO checks. AEO is 25% of the overall score, SEO is 50% and GEO is 25%.",
      },
      {
        q: "Does findable query AI products?",
        a: "No. It checks signals on your page that answer engines can use. It doesn't ask any AI product about your page.",
      },
    ],
  },
  {
    path: "/wordpress",
    title: "WordPress SEO plugin with Fix All and undo | findable",
    description:
      "A free WordPress plugin that scans posts for SEO, AEO and GEO, then applies fixes to the page head with Fix All. Undo restores the previous state.",
    body: `<h1>findable for WordPress</h1>
<p>The findable plugin scans your posts and pages for SEO, AEO and GEO from inside WordPress. It is free and open source, and it uses the same ${COUNTS} as the web scanner.</p>
${CTA}
<h2 id="what-it-does">What does it do?</h2>
<ul>
<li>A box in the editor scans the published URL, with an optional focus keyphrase.</li>
<li>A score column appears in the Posts and Pages lists.</li>
<li>Tools &gt; findable bulk scans your latest posts, up to 20.</li>
</ul>
<h2 id="fix-all">What does Fix All do?</h2>
<p>Fix All prints a meta description, a canonical if WordPress doesn't already print one, Open Graph and twitter:card tags, Article and BreadcrumbList JSON-LD, and robots max-snippet and max-image-preview in the page head. Your post content is never edited. Undo fixes restores the previous state.</p>
<p>If an SEO plugin such as Yoast SEO or Rank Math is active, findable skips printing the tags so they are not duplicated.</p>
<h2 id="privacy">What is sent where?</h2>
<p>When you click a scan button, the plugin sends the public permalink of the post, and the keyphrase if you set one, to findable.page. Our server fetches that public URL, and the request comes from your site server's IP address. Nothing is sent until you click. See the <a href="/privacy">privacy page</a>.</p>
<h2 id="install">Where do I get it?</h2>
<p>The plugin source is in the <code>wp-plugin/findable</code> folder of the <a href="https://github.com/nadimtuhin/findable" rel="noopener">GitHub repository</a>. Scans share the web scanner's rate limit of 10 per minute per IP.</p>`,
    faq: [
      {
        q: "Is the WordPress plugin free?",
        a: "Yes. It is free and open source under GPLv2 or later. The rest of the repository is MIT.",
      },
      {
        q: "Does Fix All change my posts?",
        a: "No. It prints tags in the page head and never edits post content. Undo fixes restores the previous state.",
      },
      {
        q: "Does it work with Yoast or Rank Math?",
        a: "Yes. When a known SEO plugin is detected, findable skips printing its own tags so nothing is duplicated.",
      },
    ],
  },
  {
    path: "/compare",
    title: "findable vs seoscore.tools | findable",
    description:
      "A plain comparison of findable and seoscore.tools: checks, exports, WordPress plugin, CLI, MCP, API and licence. Includes where seoscore.tools is ahead.",
    body: `<h1>findable vs seoscore.tools</h1>
<p>Both tools scan a page for SEO, AEO and GEO and have a free scan. They differ in what surrounds the scan. This table lists what each one offers, as far as we know. We built findable, so check seoscore.tools yourself before you decide.</p>
${CTA}
<h2 id="table">How do they compare?</h2>
<table>
<thead><tr><th>Feature</th><th>findable</th><th>seoscore.tools</th></tr></thead>
<tbody>
<tr><td>Free scan</td><td>Yes</td><td>Yes</td></tr>
<tr><td>SEO, AEO and GEO</td><td>Yes</td><td>Yes</td></tr>
<tr><td>Number of checks</td><td>160</td><td>260+ (as claimed by the site)</td></tr>
<tr><td>Page speed</td><td>Optional speed test</td><td>Yes</td></tr>
<tr><td>Site crawl</td><td>Yes</td><td>20 pages</td></tr>
<tr><td>Export</td><td>CSV, PDF, JSON, Markdown, SARIF</td><td>CSV, PDF</td></tr>
<tr><td>Public report link</td><td>Yes, kept 90 days</td><td>Yes, kept 90 days</td></tr>
<tr><td>Languages</td><td>5</td><td>5</td></tr>
<tr><td>WordPress plugin</td><td>Free, with Fix All, undo and full-site scan</td><td>Auto-fixes, $79</td></tr>
<tr><td>AI mentions and AI content generator</td><td>No</td><td>Yes</td></tr>
<tr><td>Command line tool</td><td>Yes</td><td>No</td></tr>
<tr><td>MCP server</td><td>Yes</td><td>No</td></tr>
<tr><td>GitHub Action</td><td>Yes</td><td>No</td></tr>
<tr><td>Public API</td><td>Yes</td><td>No</td></tr>
<tr><td>Open source</td><td>Yes, MIT</td><td>No</td></tr>
</tbody>
</table>
<h2 id="ahead">Where is seoscore.tools ahead?</h2>
<p>seoscore.tools has AI features that findable does not: AI mentions and an AI content generator. It also claims more checks, 260+ against our 160. If you want those, it is the better fit.</p>
<h2 id="findable">Where is findable a better fit?</h2>
<p>findable suits you if you want to run scans from a terminal, a CI pipeline (<code>GitHub Action</code>), an AI agent over MCP or your own code through the API. You can read and change the source, which is MIT licensed, and the WordPress plugin is free. See <a href="/wordpress">findable for WordPress</a> and <a href="/learn/findable-in-15-minutes">the command line guide</a>.</p>`,
    faq: [
      {
        q: "Is findable a replacement for seoscore.tools?",
        a: "Only for some uses. findable has a CLI, MCP server, GitHub Action, public API and open source code. seoscore.tools has AI features and claims more checks.",
      },
      {
        q: "Does findable have AI features?",
        a: "No. findable has no AI mentions tracking and no AI content generator.",
      },
      {
        q: "How many checks does each tool run?",
        a: "findable runs 160 checks. seoscore.tools claims 260+.",
      },
    ],
  },
];

export const landingEntries: PageEntry[] = LANDING_PAGES.map(({ path, title, description }) => ({
  path,
  title,
  description,
}));

export function renderLanding(pathname: string, origin: string): string | null {
  const landing = LANDING_PAGES.find((p) => p.path === pathname.replace(/(.)\/$/, "$1"));
  if (!landing) return null;
  const byline = `<p class="byline">By <a href="${AUTHOR_URL}" rel="noopener">${AUTHOR}</a>. Updated <time datetime="${UPDATED}">${UPDATED}</time>.</p>`;
  const faq = landing.faq.map((f) => `<dt>${escapeHtml(f.q)}</dt><dd>${escapeHtml(f.a)}</dd>`).join("");
  return page({
    title: landing.title,
    description: landing.description,
    path: landing.path,
    origin,
    ogType: "website",
    structuredData: [
      jsonLd({
        "@context": "https://schema.org",
        "@type": "WebPage",
        name: landing.title,
        description: landing.description,
        url: origin + landing.path,
      }),
      identityGraph(origin),
      faqLd(landing.faq),
    ],
    body: `<article>\n${byline}\n${landing.body}\n<section>\n<h2 id="faq">FAQ</h2>\n<dl>${faq}</dl>\n</section>\n${CTA}\n</article>`,
  });
}
