var __defProp = Object.defineProperty;
var __name = (target, value) => __defProp(target, "name", { value, configurable: true });

// src/index.ts
import UI_HTML from "./9e6dfde93e7bf4747b857397052f4269280c44e3-index.html";

// src/guard.ts
var MAX_PAGE_BYTES = 5 * 1024 * 1024;
var FETCH_TIMEOUT_MS = 1e4;
var MAX_REDIRECTS = 5;
var BLOCKED_SUFFIXES = [".localhost", ".local", ".internal", ".home.arpa"];
var BLOCKED_V4 = [
  [0, 8],
  [167772160, 8],
  [1681915904, 10],
  [2130706432, 8],
  [2851995648, 16],
  [2886729728, 12],
  [3221225472, 24],
  [3232235520, 16],
  [3323068416, 15],
  [3758096384, 3]
];
var BlockedTargetError = class extends Error {
  static {
    __name(this, "BlockedTargetError");
  }
  constructor(target) {
    super(`blocked address: only public internet hosts can be scanned (${target})`);
    this.name = "BlockedTargetError";
  }
};
function ipv4Number(host) {
  const parts = host.split(".");
  if (parts.length !== 4 || !parts.every((p) => /^\d{1,3}$/.test(p) && Number(p) <= 255)) return null;
  return parts.reduce((n, p) => n * 256 + Number(p), 0);
}
__name(ipv4Number, "ipv4Number");
function publicV4(n) {
  return !BLOCKED_V4.some(([base, bits]) => Math.floor(n / 2 ** (32 - bits)) === Math.floor(base / 2 ** (32 - bits)));
}
__name(publicV4, "publicV4");
function publicV6(host) {
  const mapped = /^::ffff:([0-9a-f]{1,4}):([0-9a-f]{1,4})$/.exec(host);
  if (mapped) return publicV4(Number.parseInt(mapped[1] ?? "0", 16) * 65536 + Number.parseInt(mapped[2] ?? "0", 16));
  const head = (host.split("::")[0] ?? "").split(":");
  const [first = 0, second = 0] = [0, 1].map((i) => Number.parseInt(head[i] || "0", 16));
  if (first === 8194 || first === 8193 && second === 0) return false;
  if (first === 16383 && second < 4096) return false;
  return first >= 8192 && first <= 16383;
}
__name(publicV6, "publicV6");
function parseURL(target) {
  try {
    return new URL(target);
  } catch {
    return null;
  }
}
__name(parseURL, "parseURL");
function publicTarget(target) {
  const u = parseURL(target);
  if (!u) return false;
  if (u.protocol !== "http:" && u.protocol !== "https:") return false;
  if (u.port !== "" && u.port !== "80" && u.port !== "443") return false;
  const host = u.hostname.toLowerCase().replace(/\.+$/, "");
  if (host.startsWith("[")) return publicV6(host.slice(1, -1));
  const v4 = ipv4Number(host);
  if (v4 !== null) return publicV4(v4);
  if (host === "localhost" || !host.includes(".")) return false;
  return !BLOCKED_SUFFIXES.some((s) => host.endsWith(s));
}
__name(publicTarget, "publicTarget");
async function readCapped(resp) {
  if (!resp.body) return "";
  const reader = resp.body.getReader();
  const chunks = [];
  let size = 0;
  for (; ; ) {
    const { done, value } = await reader.read();
    if (done) break;
    size += value.byteLength;
    if (size > MAX_PAGE_BYTES) {
      await reader.cancel();
      throw new Error(`page is larger than ${MAX_PAGE_BYTES} bytes`);
    }
    chunks.push(value);
  }
  const all = new Uint8Array(size);
  let offset = 0;
  for (const c of chunks) {
    all.set(c, offset);
    offset += c.byteLength;
  }
  return new TextDecoder().decode(all);
}
__name(readCapped, "readCapped");
async function fetchHop(current, headers, signal, self) {
  const u = new URL(current);
  if (!self?.hosts.includes(u.hostname.replace(/\.+$/, "").toLowerCase()))
    return fetch(current, { headers, redirect: "manual", signal });
  if (u.pathname.startsWith("/api/")) throw new BlockedTargetError(current);
  return self.handle(new Request(current, { method: "GET", headers }));
}
__name(fetchHop, "fetchHop");
async function guardedFetch(target, headers = {}, self) {
  const signal = AbortSignal.timeout(FETCH_TIMEOUT_MS);
  let current = target;
  for (let hop = 0; hop <= MAX_REDIRECTS; hop++) {
    if (!publicTarget(current)) throw new BlockedTargetError(current);
    const resp = await fetchHop(current, headers, signal, self);
    const location = resp.headers.get("location");
    if (resp.status >= 300 && resp.status < 400 && location) {
      await resp.body?.cancel();
      current = new URL(location, current).toString();
      continue;
    }
    return { url: current, status: resp.status, headers: resp.headers, body: await readCapped(resp) };
  }
  throw new Error("too many redirects");
}
__name(guardedFetch, "guardedFetch");

// src/learn-content.ts
var learnPosts = [
  {
    "slug": "aeo-checks",
    "title": "Every findable AEO Check ID and How to Fix It",
    "description": "All 17 AEO checks in findable explained, from aeo_faq to aeo_summary. See what each id tests, how loose the test is and how to fix it. AEO is 25% of the score.",
    "datePublished": "2026-10-14",
    "dateModified": "2026-10-14",
    "html": `<h1>Every AEO check findable runs, and how to fix each one</h1>
<p>AEO is about making your page easy to answer from. People ask a question. A search feature or an assistant looks for a clean passage that answers it. These 17 checks test whether your page offers that. Most are simple pattern matches on your HTML or text.</p>
<h2 id="what-does-the-aeo-family-test-in-findable">What does the AEO family test in findable?</h2>
<p>AEO is 25% of your findable score. It has 17 checks on answer shape: structured data, FAQ markup, short direct paragraphs, question headings, lists, tables, plain language and clean markup. Every check is a pattern test on your HTML and text, so each one is only a proxy.</p>
<table><thead><tr><th>Check id</th><th>What it checks</th><th>Fix in one line</th></tr></thead><tbody><tr><td><code>aeo_structured_data</code></td><td>Any JSON-LD is present</td><td>Add a JSON-LD block</td></tr><tr><td><code>aeo_faq</code></td><td>FAQPage schema type</td><td>Add FAQPage markup</td></tr><tr><td><code>aeo_direct_answers</code></td><td>A paragraph of 30 to 80 words</td><td>Answer a question in one tight paragraph</td></tr><tr><td><code>aeo_question_headings</code></td><td>A heading starts with What, How or Why, or ends with ?</td><td>Phrase an h2 as a question</td></tr><tr><td><code>aeo_definition_lists</code></td><td>A <code>&lt;dl&gt;</code> is present</td><td>Use a definition list for terms</td></tr><tr><td><code>aeo_numbered_steps</code></td><td>An <code>&lt;ol&gt;</code> is present</td><td>Write steps as an ordered list</td></tr><tr><td><code>aeo_tables</code></td><td>A <code>&lt;table&gt;</code> is present</td><td>Put comparisons in a table</td></tr><tr><td><code>aeo_lists</code></td><td>A <code>&lt;ul&gt;</code> or <code>&lt;ol&gt;</code> is present</td><td>Use a list for items</td></tr><tr><td><code>aeo_code_blocks</code></td><td><code>&lt;code&gt;</code> or <code>&lt;pre&gt;</code> is present</td><td>Mark code as code</td></tr><tr><td><code>aeo_collapsible</code></td><td>A <code>&lt;details&gt;</code> is present</td><td>Use details for extras</td></tr><tr><td><code>aeo_conversational</code></td><td>&quot;you&quot;, &quot;your&quot; or &quot;we&quot; 3 or more times</td><td>Address the reader</td></tr><tr><td><code>aeo_reading_level</code></td><td>Average sentence is 22 words or fewer</td><td>Shorten sentences</td></tr><tr><td><code>aeo_short_sentences</code></td><td>Average sentence is 20 words or fewer</td><td>Shorten them more</td></tr><tr><td><code>aeo_summary</code></td><td>Text has &quot;key takeaway&quot;, &quot;summary&quot; or &quot;conclusion&quot;</td><td>Add a summary section</td></tr><tr><td><code>aeo_clean_url</code></td><td>Path is lowercase with no underscores</td><td>Use lowercase hyphen URLs</td></tr><tr><td><code>aeo_semantic_html</code></td><td>4 or more semantic HTML5 elements</td><td>Use header, main, article, nav</td></tr><tr><td><code>aeo_ai_robots</code></td><td>No <code>noindex</code> in the robots meta</td><td>Remove <code>noindex</code></td></tr></tbody></table>
<h2 id="how-do-i-fix-the-structured-data-and-faq-checks">How do I fix the structured data and FAQ checks?</h2>
<p>Two checks cover markup: <code>aeo_structured_data</code> passes with any JSON-LD block, and <code>aeo_faq</code> passes only with a FAQPage type. Add JSON-LD that matches content people can see on the page. Do not mark up questions that are not visible. A block that is not valid JSON is dropped, so a trailing comma fails both checks at once.</p>
<pre><code>&lt;script type=&quot;application/ld+json&quot;&gt;
{
  &quot;@context&quot;: &quot;https://schema.org&quot;,
  &quot;@type&quot;: &quot;FAQPage&quot;,
  &quot;mainEntity&quot;: [{
    &quot;@type&quot;: &quot;Question&quot;,
    &quot;name&quot;: &quot;What is AEO?&quot;,
    &quot;acceptedAnswer&quot;: {&quot;@type&quot;: &quot;Answer&quot;, &quot;text&quot;: &quot;Making a page easy to answer from.&quot;}
  }]
}
&lt;/script&gt;</code></pre>
<p>See <a href="https://schema.org/FAQPage" rel="noopener">schema.org FAQPage</a> and Google's <a href="https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data" rel="noopener">structured data intro</a>. Markup helps machines read the page. It does not force a rich result.</p>
<h2 id="how-do-i-fix-the-direct-answer-and-question-heading-checks">How do I fix the direct answer and question heading checks?</h2>
<p>Two checks cover answer shape: <code>aeo_direct_answers</code> wants one paragraph of 30 to 80 words, and <code>aeo_question_headings</code> wants a heading that starts with What, How or Why, or ends with a question mark. Put the answer right under the heading. Only <code>&lt;p&gt;</code> elements count as paragraphs, so a 50 word answer inside a <code>&lt;div&gt;</code> is invisible to the check.</p>
<pre><code>&lt;h2&gt;How do I add a canonical link?&lt;/h2&gt;
&lt;p&gt;Put a link tag in the head with rel set to canonical and your preferred URL.
Every duplicate version of the page should point at that same URL. Search engines
then treat that address as the main one.&lt;/p&gt;</code></pre>
<p>The word range is a guess at a good snippet size. It is not a published rule.</p>
<h2 id="how-do-i-fix-the-list-table-and-code-checks">How do I fix the list, table and code checks?</h2>
<p>Six checks reward structure: <code>aeo_definition_lists</code>, <code>aeo_numbered_steps</code>, <code>aeo_tables</code>, <code>aeo_lists</code>, <code>aeo_code_blocks</code> and <code>aeo_collapsible</code>. Each one passes when the element exists. Use the right element for the job and these pass on their own. An <code>&lt;ol&gt;</code> satisfies both <code>aeo_numbered_steps</code> and <code>aeo_lists</code>, so a page with one ordered list and no bullets still passes both.</p>
<pre><code>&lt;ol&gt;&lt;li&gt;Open the file.&lt;/li&gt;&lt;li&gt;Add the tag.&lt;/li&gt;&lt;li&gt;Save.&lt;/li&gt;&lt;/ol&gt;
&lt;dl&gt;&lt;dt&gt;AEO&lt;/dt&gt;&lt;dd&gt;Answer engine optimization.&lt;/dd&gt;&lt;/dl&gt;
&lt;details&gt;&lt;summary&gt;More detail&lt;/summary&gt;&lt;p&gt;Extra text.&lt;/p&gt;&lt;/details&gt;</code></pre>
<p>Do not add an empty table just to pass. These checks only see that the tag exists. They cannot tell if it helps the reader. A pageful of prose with no steps or data may not need all six.</p>
<h2 id="how-do-i-fix-the-readability-checks">How do I fix the readability checks?</h2>
<p>Three checks cover plain writing: <code>aeo_conversational</code>, <code>aeo_reading_level</code> and <code>aeo_short_sentences</code>. Use &quot;you&quot; or &quot;we&quot; at least 3 times. Keep the average sentence at 22 words or fewer for reading level, and 20 or fewer for the short sentence check. The pronoun count needs a space on each side, so &quot;you,&quot; and &quot;you.&quot; are not counted.</p>
<p>Cut filler. Split long sentences at &quot;and&quot; or &quot;but&quot;. Prefer plain words. The average is computed by splitting text on periods, so a page full of bullet points with no periods looks like one giant sentence and fails. End list items with a period if that bites you.</p>
<h2 id="how-do-i-fix-the-summary-check">How do I fix the summary check?</h2>
<p>The <code>aeo_summary</code> check is simple. It passes when your text contains the phrase &quot;key takeaway&quot;, &quot;summary&quot; or &quot;conclusion&quot;. That is all. It does not read your summary or judge it. A page that says &quot;no summary&quot; would pass. The match is case-insensitive and runs over the whole page text, so a nav link called &quot;Summary&quot; passes it too.</p>
<p>Do the real thing anyway. End the page with a &quot;Key takeaways&quot; section of 3 to 6 bullets. Readers who skim land there first. The check is a nudge to write one, not proof that you wrote a good one.</p>
<h2 id="how-do-i-fix-the-url-semantic-html-and-robots-checks">How do I fix the URL, semantic HTML and robots checks?</h2>
<p>Three checks cover housekeeping: <code>aeo_clean_url</code> wants a lowercase path with no underscores, <code>aeo_semantic_html</code> wants 4 or more semantic HTML5 elements, and <code>aeo_ai_robots</code> fails if the robots meta contains <code>noindex</code>. Use hyphens, landmark elements and no noindex. The elements counted are article, section, main, nav, aside, header, footer, time, figure and details.</p>
<pre><code>&lt;!-- good path: /learn/aeo-checks  bad path: /Learn/AEO_Checks --&gt;
&lt;header&gt;...&lt;/header&gt;
&lt;nav&gt;...&lt;/nav&gt;
&lt;main&gt;&lt;article&gt;...&lt;/article&gt;&lt;/main&gt;
&lt;footer&gt;...&lt;/footer&gt;</code></pre>
<p>Despite the name, <code>aeo_ai_robots</code> only checks the robots meta tag for <code>noindex</code>. It does not read robots.txt, and it ignores <code>nofollow</code> and <code>nosnippet</code>. For AI crawler rules in robots.txt, run with <code>-site</code> and read the unscored Site checks in the GEO guide.</p>
<h2 id="key-takeaways">Key takeaways</h2>
<ul><li>AEO is 25% of the findable score and has 17 checks.</li><li>Most checks only test that an element or phrase exists. They are proxies.</li><li>Put a 30 to 80 word answer right under a question heading.</li><li>Use lists, tables, steps and code blocks where the content calls for them.</li><li><code>aeo_summary</code> only looks for the words summary, key takeaway or conclusion. Write a real one.</li><li>Keep sentences short and speak to the reader as &quot;you&quot;.</li><li>A good score helps, but it does not guarantee a citation anywhere.</li></ul>`,
    "faq": [
      {
        "q": "What is AEO in findable?",
        "a": "AEO means answer engine optimization. The checks test whether your page is easy to lift an answer from: clear questions, short answers, lists, tables, structured data and plain sentences."
      },
      {
        "q": "How much of the score is AEO?",
        "a": "AEO is 25% of the findable score. SEO is 50% and GEO is 25%."
      },
      {
        "q": "Is aeo_summary a smart check?",
        "a": "No. It only looks for the words summary, key takeaway or conclusion in your text. It cannot judge whether your summary is good, so write a real one."
      },
      {
        "q": "Does passing AEO checks guarantee I get quoted by an AI answer?",
        "a": "No. The checks measure structure that helps, not a promise. Nobody outside the search companies can guarantee a citation, and findable does not claim to."
      }
    ]
  },
  {
    "slug": "findable-in-15-minutes",
    "title": "Get Findable in 15 Minutes: A Step by Step Checklist",
    "description": "Install findable, audit a page and your site, fix the top failures, then add MCP, the web UI, the API and a CI gate. Every step has a copy-paste command.",
    "datePublished": "2026-10-14",
    "dateModified": "2026-10-14",
    "html": `<h1>Get Findable in 15 Minutes: A Step by Step Checklist</h1>
<p>You can go from nothing to a working audit, an editor integration and a CI gate in about 15 minutes. Follow the steps in order. Each one is a command you can paste.</p>
<h2 id="how-do-you-install-findable">How do you install findable?</h2>
<p>You install it with one Go command. It puts a binary called <code>seoscore</code> on your path. The product is called findable, but the command line tool keeps the older name. You need a recent Go toolchain, and this takes about one minute.</p>
<pre><code>go install github.com/nadimtuhin/seoscore-cli/cmd/seoscore@latest</code></pre>
<h2 id="how-do-you-run-your-first-audit">How do you run your first audit?</h2>
<p>Point <code>seoscore</code> at a URL and pass your target phrase with <code>-k</code>. It fetches the page, runs every check and prints a score. You can also audit an HTML file on disk before you deploy it. Expect this to take about one minute.</p>
<pre><code>seoscore https://example.com -k &quot;your keyphrase&quot;
seoscore ./page.html</code></pre>
<h2 id="how-do-you-check-robots-txt-and-llms-txt">How do you check robots.txt and llms.txt?</h2>
<p>Add the <code>-site</code> flag. It also fetches <code>/robots.txt</code> and <code>/llms.txt</code> and reports whether GPTBot, ClaudeBot, PerplexityBot and Google-Extended are allowed. These site results are unscored, so they never change your grade. The flag needs a real URL and refuses a local file.</p>
<pre><code>seoscore https://example.com -k &quot;your keyphrase&quot; -site</code></pre>
<h2 id="how-do-you-read-the-output">How do you read the output?</h2>
<p>The score is SEO 50%, AEO 25% and GEO 25%. Each line is a check with a stable id, such as <code>title_length</code> or <code>aeo_faq</code>. Failures matter most. Read the failing ids first, then the site section if you used <code>-site</code>.</p>
<p>For a machine readable report, choose a format. You can write <code>csv</code>, <code>json</code> or <code>sarif</code>, send files to a folder with <code>-o</code>, and quiet the console with <code>-q</code>. These are handy when you want to diff two runs, or feed the findings into another tool.</p>
<pre><code>seoscore https://example.com -format json -o ./reports -q</code></pre>
<h2 id="how-do-you-fix-the-top-failures">How do you fix the top failures?</h2>
<p>Take the first three failing ids and fix them. Quick wins are usually <code>title_length</code> (30 to 60 characters), <code>desc_length</code> (120 to 160), <code>h1_exists</code> (exactly one h1) and <code>canonical</code>. Then re-run the same command and watch the score rise. Repeat until the top fails are gone.</p>
<pre><code>seoscore https://example.com -k &quot;your keyphrase&quot; -site</code></pre>
<p>To audit more than one page, crawl. Use <code>-crawl</code> with <code>-limit</code>, or feed a sitemap with <code>-sitemap</code>.</p>
<pre><code>seoscore https://example.com -crawl -limit 25
seoscore https://example.com -sitemap https://example.com/sitemap.xml</code></pre>
<h2 id="how-do-you-set-up-mcp">How do you set up MCP?</h2>
<p>Run findable as an MCP server so your editor agent can audit pages for you. Add this to your MCP config. You can also launch it with <code>npx -y seoscore-mcp</code> if you prefer not to install Go. The tools are <code>audit_page</code>, <code>audit_html_content</code>, <code>crawl_site</code> and <code>generate_csv_report</code>.</p>
<pre><code>{&quot;mcpServers&quot;:{&quot;seoscore&quot;:{&quot;command&quot;:&quot;seoscore&quot;,&quot;args&quot;:[&quot;-mcp&quot;]}}}</code></pre>
<p>The server also offers a prompt called <code>seo_audit_prompt</code>. It speaks over stdio, so there is no port to open and nothing is exposed to your network. Ask your agent to audit a page, and it calls these tools for you.</p>
<h2 id="how-do-you-use-the-web-ui">How do you use the web UI?</h2>
<p>Open <a href="https://findable.piply.xyz" rel="noopener">findable.piply.xyz</a> and paste a URL to get a scored report in your browser, with no install at all. If you want it local, serve it yourself. The same binary hosts the web UI and the API on a port you choose.</p>
<pre><code>seoscore -serve 127.0.0.1:8080</code></pre>
<h2 id="how-do-you-call-the-api-with-curl">How do you call the API with curl?</h2>
<p>Send a POST to <code>/api/scan</code> with a JSON body. Set <code>site</code> to true for the robots.txt and llms.txt checks, and set <code>format</code> to <code>json</code> so you can parse the response in a script or a dashboard. The response uses the same check ids as the CLI.</p>
<pre><code>curl -X POST https://findable.piply.xyz/api/scan \\
  -H &quot;Content-Type: application/json&quot; \\
  -d '{&quot;url&quot;:&quot;https://example.com&quot;,&quot;keyphrase&quot;:&quot;your keyphrase&quot;,&quot;site&quot;:true,&quot;format&quot;:&quot;json&quot;}'</code></pre>
<h2 id="how-do-you-gate-ci-with-the-github-action">How do you gate CI with the GitHub Action?</h2>
<p>Add the action to a workflow. Set <code>fail-under</code> to the lowest score you accept. The job fails below that number, and the action uploads SARIF so each failing check shows up as a code scanning alert in the repository. Set <code>upload-sarif</code> to false if you only want the exit code.</p>
<pre><code>name: findable
on: [push]
jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: nadimtuhin/findable@v0.2.0
        with:
          url: https://example.com
          fail-under: 80
          args: -site</code></pre>
<p>You can get the same gate locally with <code>seoscore https://example.com -fail-under 80</code>. It exits 1 below 80. Start with a threshold near your current score, then raise it as you fix more checks. The Site checks are unscored, so <code>-site</code> in <code>args</code> adds rows to the report but never changes whether the gate passes.</p>
<h2 id="key-takeaways">Key takeaways</h2>
<ul><li>Install with one <code>go install</code> line. The binary is <code>seoscore</code>.</li><li>Run <code>seoscore &lt;url&gt; -k &quot;phrase&quot; -site</code> and read the failing check ids first.</li><li>Fix the top three failures, then re-run to confirm.</li><li>Use <code>-mcp</code> for editors, <code>-serve 127.0.0.1:8080</code> for a local UI, and POST <code>/api/scan</code> for scripts.</li><li>Gate CI with <code>-fail-under N</code> or the <code>nadimtuhin/findable@v0.2.0</code> action.</li></ul>`,
    "faq": [
      {
        "q": "Does findable need an API key?",
        "a": "No. The CLI runs locally and audits any URL or local HTML file. You only need Go to install it."
      },
      {
        "q": "What is the binary called?",
        "a": "The product is findable, but the command line binary is still named seoscore. Use seoscore in every command."
      },
      {
        "q": "How do I fail a build on a low score?",
        "a": "Pass -fail-under N. The CLI exits with code 1 when the score is below N, which stops your CI job."
      },
      {
        "q": "Can I run it on a local file?",
        "a": "Yes. Run seoscore ./page.html to audit a file on disk with no server."
      }
    ]
  },
  {
    "slug": "geo-checks",
    "title": "Every findable GEO Check ID, Plus AI Crawler Checks",
    "description": "All 19 GEO checks in findable explained, plus the unscored Site checks for GPTBot, ClaudeBot and llms.txt. GEO is 25% of the score. Fixes included.",
    "datePublished": "2026-10-14",
    "dateModified": "2026-10-14",
    "html": `<h1>Every GEO check findable runs, and how to fix each one</h1>
<p>AI tools now write answers and cite pages. GEO is the work of making your page worth citing. findable runs 19 GEO checks, and a short set of unscored Site checks for AI crawlers. This guide covers both. Remember that these are heuristics. They test signals, not quality.</p>
<h2 id="what-does-the-geo-family-test-in-findable">What does the GEO family test in findable?</h2>
<p>GEO is 25% of your findable score. It has 19 checks on citability: numbers, tables, quotes, outbound sources, clear definitions, author and organization markup, trust pages, fresh dates and content length. Each check is a pattern match, so a pass shows the signal exists, not that it is good.</p>
<table><thead><tr><th>Check id</th><th>What it checks</th><th>Fix in one line</th></tr></thead><tbody><tr><td><code>geo_statistics_density</code></td><td>3 or more numbers with units</td><td>State real figures with units</td></tr><tr><td><code>geo_data_viz</code></td><td>A table is present</td><td>Show data in a table</td></tr><tr><td><code>geo_comparison</code></td><td>A table or &quot; vs &quot; in text</td><td>Compare options side by side</td></tr><tr><td><code>geo_expert_quotes</code></td><td>A blockquote or expert quote</td><td>Quote a named source</td></tr><tr><td><code>geo_citations</code></td><td>1 or more external link</td><td>Link to your sources</td></tr><tr><td><code>geo_definitions</code></td><td>&quot; is a &quot;, &quot; is the &quot; or &quot; defined as &quot;</td><td>Define key terms plainly</td></tr><tr><td><code>geo_passage_citability</code></td><td>A paragraph of 40 to 120 words</td><td>Write self-contained paragraphs</td></tr><tr><td><code>geo_author_schema</code></td><td>Person schema or author in schema</td><td>Add author to JSON-LD</td></tr><tr><td><code>geo_author_credentials</code></td><td>A visible &quot;by &quot; byline or schema author</td><td>Show who wrote it</td></tr><tr><td><code>geo_org_schema</code></td><td>Organization or LocalBusiness schema</td><td>Add Organization JSON-LD</td></tr><tr><td><code>geo_knowledge_graph</code></td><td><code>sameAs</code> links in schema</td><td>Link your profiles</td></tr><tr><td><code>geo_trust</code></td><td>Internal links to about and contact</td><td>Link both pages</td></tr><tr><td><code>geo_entity_markup</code></td><td>2 or more schema types</td><td>Mark up more entities</td></tr><tr><td><code>geo_freshness</code></td><td><code>&lt;time&gt;</code>, a year in text, or <code>dateModified</code></td><td>Show a date</td></tr><tr><td><code>geo_time_element</code></td><td>A <code>&lt;time&gt;</code> element</td><td>Wrap dates in <code>&lt;time&gt;</code></td></tr><tr><td><code>geo_content_length</code></td><td>300 or more words</td><td>Write enough to be useful</td></tr><tr><td><code>geo_intent_clarity</code></td><td>2 or more title words (over 3 letters) in the first heading</td><td>Match h1 to title</td></tr><tr><td><code>geo_format_variety</code></td><td>3 or more of lists, tables, quotes, code, dl</td><td>Mix formats</td></tr><tr><td><code>geo_snippet_control</code></td><td>No <code>nosnippet</code> in the robots meta</td><td>Remove <code>nosnippet</code></td></tr></tbody></table>
<h2 id="how-do-i-fix-the-evidence-checks">How do I fix the evidence checks?</h2>
<p>Five checks reward evidence: <code>geo_statistics_density</code>, <code>geo_data_viz</code>, <code>geo_comparison</code>, <code>geo_expert_quotes</code> and <code>geo_citations</code>. Add real numbers with units, a table, a named quote and links to sources. Never invent a figure to pass. A fake number hurts trust. The stat matcher wants a unit it knows, such as <code>%</code>, <code>ms</code>, <code>MB</code>, <code>px</code>, <code>hours</code> or a <code>$</code> amount; &quot;300 words&quot; does not count.</p>
<pre><code>&lt;blockquote&gt;
  &lt;p&gt;Quote text from a named person.&lt;/p&gt;
  &lt;footer&gt;Name, role, &lt;a href=&quot;https://example.org/source&quot;&gt;source&lt;/a&gt;&lt;/footer&gt;
&lt;/blockquote&gt;
&lt;p&gt;Source: &lt;a href=&quot;https://arxiv.org/abs/2311.09735&quot;&gt;GEO: Generative Engine Optimization&lt;/a&gt;.&lt;/p&gt;</code></pre>
<p>The <a href="https://arxiv.org/abs/2311.09735" rel="noopener">GEO paper</a> by Aggarwal et al. tested ideas like adding citations, quotations and statistics. In their tests, these helped visibility. Results will vary by engine and topic. <code>geo_expert_quotes</code> is satisfied by any <code>&lt;blockquote&gt;</code> with text in it; nothing checks that a named person said it.</p>
<h2 id="how-do-i-fix-the-passage-and-definition-checks">How do I fix the passage and definition checks?</h2>
<p>Three checks cover passages: <code>geo_definitions</code>, <code>geo_passage_citability</code> and <code>geo_content_length</code>. State what a thing is using &quot;is a&quot; or &quot;is the&quot;. Write at least one paragraph of 40 to 120 words that stands alone. Reach 300 words in total. Only <code>&lt;p&gt;</code> text counts toward the passage check, so an answer written as a list or inside a <code>&lt;div&gt;</code> is never seen.</p>
<p>A passage is citable when you can lift it out and it still makes sense. Name the subject in the first sentence. Avoid &quot;this&quot; and &quot;it&quot; with no clear referent. The <code>geo_definitions</code> test only looks for the phrases &quot; is a &quot;, &quot; is the &quot; or &quot; defined as &quot;, so it is easy to pass and easy to fool.</p>
<h2 id="how-do-i-fix-the-author-and-organization-checks">How do I fix the author and organization checks?</h2>
<p>Six checks cover who is behind the page: <code>geo_author_schema</code>, <code>geo_author_credentials</code>, <code>geo_org_schema</code>, <code>geo_knowledge_graph</code>, <code>geo_trust</code> and <code>geo_entity_markup</code>. Show a byline, add Person and Organization JSON-LD with <code>sameAs</code> links, link to your about and contact pages, and use at least 2 schema types.</p>
<pre><code>&lt;p&gt;By Nadim Tuhin&lt;/p&gt;
&lt;script type=&quot;application/ld+json&quot;&gt;
{
  &quot;@context&quot;: &quot;https://schema.org&quot;,
  &quot;@graph&quot;: [
    {&quot;@type&quot;: &quot;Organization&quot;, &quot;name&quot;: &quot;Example&quot;, &quot;url&quot;: &quot;https://example.com&quot;,
     &quot;sameAs&quot;: [&quot;https://github.com/example&quot;]},
    {&quot;@type&quot;: &quot;Person&quot;, &quot;name&quot;: &quot;Nadim Tuhin&quot;}
  ]
}
&lt;/script&gt;</code></pre>
<p>The <code>geo_trust</code> check looks for the strings &quot;about&quot; and &quot;contact&quot; in your internal link URLs, so <code>/about-us</code> and <code>/contact-sales</code> both count. The schema author only counts when <code>author</code> is an object with a <code>name</code>; a bare string is ignored. See <a href="https://schema.org/" rel="noopener">schema.org</a> for types, and Google's <a href="https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data" rel="noopener">structured data intro</a>.</p>
<h2 id="how-do-i-fix-the-freshness-intent-and-format-checks">How do I fix the freshness, intent and format checks?</h2>
<p>Five checks cover shape and dates: <code>geo_freshness</code>, <code>geo_time_element</code>, <code>geo_intent_clarity</code>, <code>geo_format_variety</code> and <code>geo_snippet_control</code>. Show a date in a <code>&lt;time&gt;</code> element, share words between title and first heading, mix 3 formats, and avoid <code>nosnippet</code> in the robots meta. The five formats counted are lists, tables, blockquotes, code and definition lists.</p>
<pre><code>&lt;time datetime=&quot;2026-10-14&quot;&gt;October 14, 2026&lt;/time&gt;</code></pre>
<p><code>geo_freshness</code> passes on a <code>&lt;time&gt;</code> element, a year from 2024 to 2029 in the text, or a <code>dateModified</code> in schema. A year in the text is a weak signal, and a 2023 date does not count. <code>geo_intent_clarity</code> counts shared words longer than 3 letters between the title and the first heading. <code>geo_snippet_control</code> fails if <code>nosnippet</code> limits what can be shown.</p>
<h2 id="what-are-the-site-checks-and-why-do-ai-crawlers-care">What are the Site checks and why do AI crawlers care?</h2>
<p>The Site checks run only with the <code>-site</code> flag and are unscored. They fetch <code>/robots.txt</code> and <code>/llms.txt</code> and test whether named AI bots may read your page. They matter because a blocked crawler cannot read even a perfect page. <code>-site</code> needs an http or https URL; it refuses a local HTML file.</p>
<pre><code>seoscore -site https://example.com</code></pre>
<table><thead><tr><th>Check id</th><th>What it checks</th></tr></thead><tbody><tr><td><code>robots-txt</code></td><td><code>/robots.txt</code> exists</td></tr><tr><td><code>robots-page</code></td><td>robots.txt allows <code>*</code> to fetch the page</td></tr><tr><td><code>robots-gptbot</code></td><td>robots.txt allows GPTBot</td></tr><tr><td><code>robots-claudebot</code></td><td>robots.txt allows ClaudeBot</td></tr><tr><td><code>robots-perplexitybot</code></td><td>robots.txt allows PerplexityBot</td></tr><tr><td><code>robots-google-extended</code></td><td>robots.txt allows Google-Extended</td></tr><tr><td><code>llms-txt</code></td><td><code>/llms.txt</code> exists</td></tr></tbody></table>
<h2 id="how-do-i-let-ai-bots-in-or-keep-them-out">How do I let AI bots in, or keep them out?</h2>
<p>Edit <code>/robots.txt</code>. A bot you name with <code>Disallow: /</code> is blocked. A bot you do not mention follows the <code>*</code> rules. The choice is yours. A bot that cannot fetch the page cannot cite it. Blocking one keeps your text out of that vendor's crawl. findable only reports the current state.</p>
<pre><code>User-agent: GPTBot
Allow: /

User-agent: ClaudeBot
Allow: /

User-agent: PerplexityBot
Allow: /

User-agent: Google-Extended
Allow: /</code></pre>
<p>Official docs: <a href="https://platform.openai.com/docs/bots" rel="noopener">OpenAI crawlers</a>, <a href="https://support.anthropic.com/en/articles/8896518-does-anthropic-crawl-data-from-the-web-and-how-can-site-owners-block-the-crawler" rel="noopener">Anthropic ClaudeBot</a>, <a href="https://docs.perplexity.ai/guides/bots" rel="noopener">Perplexity bots</a>, <a href="https://developers.google.com/search/docs/crawling-indexing/google-common-crawlers" rel="noopener">Google common crawlers</a> and the <a href="https://developers.google.com/search/docs/crawling-indexing/robots/intro" rel="noopener">robots.txt intro</a>. Note that robots.txt is a request, not a lock.</p>
<p>For <code>llms-txt</code>, add a plain Markdown file at <code>/llms.txt</code> that lists your key pages. See the <a href="https://llmstxt.org/" rel="noopener">llms.txt proposal</a>. It is a proposal, and no one promises that every AI tool reads it.</p>
<h2 id="key-takeaways">Key takeaways</h2>
<ul><li>GEO is 25% of the findable score and has 19 checks.</li><li>Real sources, quotes, tables and numbers with units make a page easier to cite.</li><li>Never invent a statistic to pass <code>geo_statistics_density</code>.</li><li>Show a byline, add Person and Organization JSON-LD, and link about and contact pages.</li><li>Site checks are unscored. Run them with <code>-site</code>.</li><li>Check <code>robots-gptbot</code>, <code>robots-claudebot</code>, <code>robots-perplexitybot</code> and <code>robots-google-extended</code>, then decide who you want in.</li><li>Add <code>/llms.txt</code> if you want to, knowing it is only a proposal.</li></ul>`,
    "faq": [
      {
        "q": "What is GEO in findable?",
        "a": "GEO means generative engine optimization. The checks look for things that make a page easy to cite in a generated answer: sources, numbers, quotes, authorship, entity markup and fresh dates."
      },
      {
        "q": "How much of the score is GEO?",
        "a": "GEO is 25% of the findable score. SEO is 50% and AEO is 25%."
      },
      {
        "q": "Are the Site checks part of the score?",
        "a": "No. The Site checks run only with the -site flag and are unscored. They test robots.txt, AI bot access and llms.txt, which still decide whether AI crawlers can read you."
      },
      {
        "q": "Does llms.txt make AI tools cite me?",
        "a": "No promise. llms.txt is a proposal, not a standard every crawler follows. findable only checks that the file exists."
      }
    ]
  },
  {
    "slug": "seo-checks",
    "title": "Every findable SEO Check ID and How to Fix It",
    "description": "A plain guide to all 36 SEO checks in findable, from title_length to csp. See what each id tests, when it passes and the fix. SEO is 50% of your score.",
    "datePublished": "2026-10-14",
    "dateModified": "2026-10-14",
    "html": `<h1>Every SEO check findable runs, and how to fix each one</h1>
<p>You ran an audit and got a list of ids like <code>title_length</code> and <code>img_dimensions</code>. This guide explains every one of them. Each check is a simple test on your HTML or response headers. Fix the failing ones and your score goes up.</p>
<h2 id="what-does-the-seo-family-cover-in-findable">What does the SEO family cover in findable?</h2>
<p>SEO is 50% of your findable score, the biggest share. It has 36 checks on the basics: metadata, page setup, headings, content size, links, images, social tags, security headers and markup. Each check has a fixed id, so you can search for it in any report.</p>
<table><thead><tr><th>Check id</th><th>What it checks</th><th>Fix in one line</th></tr></thead><tbody><tr><td><code>title_exists</code></td><td>A title tag is present</td><td>Add <code>&lt;title&gt;</code> in the head</td></tr><tr><td><code>title_length</code></td><td>Title is 30 to 60 characters</td><td>Rewrite the title to fit</td></tr><tr><td><code>desc_exists</code></td><td>A meta description is present</td><td>Add <code>&lt;meta name=&quot;description&quot;&gt;</code></td></tr><tr><td><code>desc_length</code></td><td>Description is 120 to 160 characters</td><td>Trim or expand it</td></tr><tr><td><code>viewport</code></td><td>A viewport meta tag is present</td><td>Add the viewport tag</td></tr><tr><td><code>charset</code></td><td>UTF-8 is declared</td><td>Add <code>&lt;meta charset=&quot;utf-8&quot;&gt;</code></td></tr><tr><td><code>lang_attr</code></td><td><code>&lt;html&gt;</code> has a lang attribute</td><td>Set <code>lang=&quot;en&quot;</code> or your language</td></tr><tr><td><code>html_doctype</code></td><td><code>&lt;!DOCTYPE html&gt;</code> is present</td><td>Put it on line one</td></tr><tr><td><code>meta_refresh</code></td><td>No meta refresh tag</td><td>Use a server redirect instead</td></tr><tr><td><code>robots_meta</code></td><td>No <code>noindex</code> in the robots meta</td><td>Remove <code>noindex</code></td></tr><tr><td><code>h1_exists</code></td><td>Exactly one h1</td><td>Keep one h1 per page</td></tr><tr><td><code>h2_exists</code></td><td>At least one h2</td><td>Split content into sections</td></tr><tr><td><code>heading_hierarchy</code></td><td>No skipped heading levels</td><td>Go h1, h2, h3 in order</td></tr><tr><td><code>word_count</code></td><td>300 or more words</td><td>Add useful body text</td></tr><tr><td><code>text_html_ratio</code></td><td>Text is at least 15% of the HTML</td><td>Cut markup bloat</td></tr><tr><td><code>dom_size</code></td><td>Fewer than 1500 elements</td><td>Simplify the markup</td></tr><tr><td><code>internal_links</code></td><td>2 or more internal links</td><td>Link to related pages</td></tr><tr><td><code>external_links</code></td><td>1 or more external link</td><td>Cite a source</td></tr><tr><td><code>empty_hrefs</code></td><td>No empty href values</td><td>Fix or remove the link</td></tr><tr><td><code>link_noopener</code></td><td><code>target=&quot;_blank&quot;</code> links have <code>rel=&quot;noopener&quot;</code></td><td>Add the rel attribute</td></tr><tr><td><code>img_alt</code></td><td>Every image has alt text</td><td>Add alt to each img</td></tr><tr><td><code>img_dimensions</code></td><td>Every image has width and height</td><td>Add both attributes</td></tr><tr><td><code>canonical</code></td><td>A canonical link is present</td><td>Add <code>&lt;link rel=&quot;canonical&quot;&gt;</code></td></tr><tr><td><code>og_title</code></td><td>Open Graph title is present</td><td>Add <code>og:title</code></td></tr><tr><td><code>og_desc</code></td><td>Open Graph description is present</td><td>Add <code>og:description</code></td></tr><tr><td><code>og_image</code></td><td>Open Graph image is present</td><td>Add <code>og:image</code></td></tr><tr><td><code>og_url</code></td><td>Open Graph URL is present</td><td>Add <code>og:url</code></td></tr><tr><td><code>twitter_card</code></td><td>A <code>twitter:card</code> meta is present</td><td>Add the card tag</td></tr><tr><td><code>https</code></td><td>Page is served over HTTPS</td><td>Install a certificate</td></tr><tr><td><code>hsts</code></td><td>Strict-Transport-Security header is set</td><td>Send the header</td></tr><tr><td><code>x_content_type</code></td><td><code>X-Content-Type-Options: nosniff</code> is set</td><td>Send the header</td></tr><tr><td><code>x_frame_options</code></td><td>X-Frame-Options header is set</td><td>Send the header</td></tr><tr><td><code>csp</code></td><td>Content-Security-Policy header is set</td><td>Send a policy</td></tr><tr><td><code>jsonld</code></td><td>At least one JSON-LD block</td><td>Add structured data</td></tr><tr><td><code>aria_landmarks</code></td><td><code>&lt;main&gt;</code> and <code>&lt;nav&gt;</code> are present</td><td>Use landmark elements</td></tr><tr><td><code>duplicate_ids</code></td><td>No duplicate id attributes</td><td>Make each id unique</td></tr></tbody></table>
<h2 id="how-do-i-fix-the-title-and-description-checks">How do I fix the title and description checks?</h2>
<p>Four checks cover this: <code>title_exists</code>, <code>title_length</code>, <code>desc_exists</code> and <code>desc_length</code>. The title must exist and be 30 to 60 characters. The description must exist and be 120 to 160 characters. Search engines use both to build your result snippet, so write them for people.</p>
<pre><code>&lt;title&gt;Every SEO Check ID and How to Fix It&lt;/title&gt;
&lt;meta name=&quot;description&quot; content=&quot;A plain guide to the SEO checks, with the fix for each one.&quot;&gt;</code></pre>
<p>The length ranges are common guidance, not a hard rule from Google. Google may rewrite your snippet anyway. See the <a href="https://developers.google.com/search/docs/fundamentals/seo-starter-guide" rel="noopener">SEO starter guide</a> for what it says about titles and snippets.</p>
<h2 id="how-do-i-fix-the-page-setup-checks">How do I fix the page setup checks?</h2>
<p>Six checks cover basic page setup: <code>viewport</code>, <code>charset</code>, <code>lang_attr</code>, <code>html_doctype</code>, <code>meta_refresh</code> and <code>robots_meta</code>. Most are one-line additions to the head. The last two are about things you must not have: a meta refresh and a <code>noindex</code> directive. <code>html_doctype</code> ignores leading whitespace but fails if a byte order mark comes before the doctype.</p>
<pre><code>&lt;!DOCTYPE html&gt;
&lt;html lang=&quot;en&quot;&gt;
&lt;head&gt;
  &lt;meta charset=&quot;utf-8&quot;&gt;
  &lt;meta name=&quot;viewport&quot; content=&quot;width=device-width, initial-scale=1&quot;&gt;</code></pre>
<p>If <code>robots_meta</code> fails, look for <code>noindex</code> in <code>&lt;meta name=&quot;robots&quot;&gt;</code>. It tells crawlers to skip the page. If <code>meta_refresh</code> fails, replace the tag with a 301 redirect on the server.</p>
<h2 id="how-do-i-fix-the-heading-and-content-checks">How do I fix the heading and content checks?</h2>
<p>Six checks look at structure and size: <code>h1_exists</code>, <code>h2_exists</code>, <code>heading_hierarchy</code>, <code>word_count</code>, <code>text_html_ratio</code> and <code>dom_size</code>. Use one h1, then h2 sections, then h3 inside them. Never jump from h2 to h4. Write at least 300 words of real text. A page with two h1 tags fails <code>h1_exists</code> just like a page with none.</p>
<p>Be honest about three of these. <code>word_count</code>, <code>text_html_ratio</code> (15%) and <code>dom_size</code> (1500 elements) are heuristics. Google does not publish these cutoffs. A short page can still be great. Use them as a smell test for thin or bloated pages.</p>
<h2 id="how-do-i-fix-the-link-checks">How do I fix the link checks?</h2>
<p>Four checks cover links: <code>internal_links</code>, <code>external_links</code>, <code>empty_hrefs</code> and <code>link_noopener</code>. You need at least 2 internal links and 1 external link. No link may have an empty href. Every <code>target=&quot;_blank&quot;</code> link needs <code>rel=&quot;noopener&quot;</code>. A link to another host counts as external, so a link to your own subdomain is external too.</p>
<pre><code>&lt;a href=&quot;https://example.org/source&quot; target=&quot;_blank&quot; rel=&quot;noopener&quot;&gt;Source&lt;/a&gt;</code></pre>
<p>The <code>noopener</code> value stops the new page from reaching your page through <code>window.opener</code>. <code>rel=&quot;noreferrer&quot;</code> also passes the check, since it implies noopener. Read the <a href="https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/rel/noopener" rel="noopener">MDN page on rel=noopener</a> for details.</p>
<h2 id="how-do-i-fix-the-image-checks">How do I fix the image checks?</h2>
<p>Two checks cover images: <code>img_alt</code> and <code>img_dimensions</code>. Every image needs an alt attribute and explicit width and height. The alt check passes when the page has no images. The size attributes let the browser reserve space, so the layout does not jump.</p>
<pre><code>&lt;img src=&quot;/chart.png&quot; alt=&quot;Bar chart of audit scores by family&quot; width=&quot;800&quot; height=&quot;450&quot;&gt;</code></pre>
<p>One catch: findable treats <code>alt=&quot;&quot;</code> as missing, because the check trims whitespace and tests for empty. Decorative images with an empty alt are correct HTML, but they will still fail <code>img_alt</code>. Decide which matters more to you. Do not stuff keywords into alt text. Describe the image.</p>
<h2 id="how-do-i-fix-the-canonical-and-social-tag-checks">How do I fix the canonical and social tag checks?</h2>
<p>Six checks cover sharing and duplicates: <code>canonical</code>, <code>og_title</code>, <code>og_desc</code>, <code>og_image</code>, <code>og_url</code> and <code>twitter_card</code>. The canonical link names the preferred URL. The Open Graph and Twitter tags control how your link looks when shared. See the <a href="https://ogp.me/" rel="noopener">Open Graph protocol</a>. Each check only tests that the tag has a non-empty value, so a broken <code>og:image</code> URL still passes.</p>
<pre><code>&lt;link rel=&quot;canonical&quot; href=&quot;https://example.com/page&quot;&gt;
&lt;meta property=&quot;og:title&quot; content=&quot;Page title&quot;&gt;
&lt;meta property=&quot;og:description&quot; content=&quot;Short summary.&quot;&gt;
&lt;meta property=&quot;og:image&quot; content=&quot;https://example.com/card.png&quot;&gt;
&lt;meta property=&quot;og:url&quot; content=&quot;https://example.com/page&quot;&gt;
&lt;meta name=&quot;twitter:card&quot; content=&quot;summary_large_image&quot;&gt;</code></pre>
<h2 id="how-do-i-fix-the-security-header-checks">How do I fix the security header checks?</h2>
<p>Five checks look at transport and headers: <code>https</code>, <code>hsts</code>, <code>x_content_type</code>, <code>x_frame_options</code> and <code>csp</code>. They test your server response, not your HTML. A check passes when the header is present. It does not judge how strong the value is. If you audit a local file with <code>seoscore ./page.html</code>, there are no response headers, so expect these five to fail.</p>
<pre><code>Strict-Transport-Security: max-age=31536000
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Content-Security-Policy: default-src 'self'</code></pre>
<p>Set these in your web server or CDN. Start the CSP in report-only mode so you do not break scripts. The <a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP" rel="noopener">MDN CSP guide</a> explains the options.</p>
<h2 id="how-do-i-fix-the-structured-data-and-markup-checks">How do I fix the structured data and markup checks?</h2>
<p>Three checks cover markup: <code>jsonld</code>, <code>aria_landmarks</code> and <code>duplicate_ids</code>. Add at least one JSON-LD block, use both <code>&lt;main&gt;</code> and <code>&lt;nav&gt;</code>, and keep every id unique. Duplicate ids break anchors and scripts, and they confuse screen readers. The <code>jsonld</code> check drops any block that is not valid JSON, so one stray trailing comma makes it fail.</p>
<pre><code>&lt;script type=&quot;application/ld+json&quot;&gt;
{&quot;@context&quot;:&quot;https://schema.org&quot;,&quot;@type&quot;:&quot;Article&quot;,&quot;headline&quot;:&quot;Every SEO check&quot;}
&lt;/script&gt;</code></pre>
<p>Read Google's <a href="https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data" rel="noopener">structured data intro</a> to pick a type that fits your page.</p>
<h2 id="key-takeaways">Key takeaways</h2>
<ul><li>SEO is 50% of the findable score, so start here.</li><li>There are 36 SEO checks. Each has a fixed id and a simple pass test.</li><li>Fix the head first: title, description, charset, viewport, lang, doctype.</li><li><code>word_count</code>, <code>text_html_ratio</code> and <code>dom_size</code> are heuristics. Use judgment.</li><li>Security header checks pass when the header exists. They do not rate its strength.</li><li>Run <code>seoscore -fail-under 80 URL</code> in CI to catch regressions.</li></ul>`,
    "faq": [
      {
        "q": "How much of the findable score is SEO?",
        "a": "SEO is 50% of the score. AEO and GEO are 25% each. The SEO family has the most checks and covers metadata, headings, links, images, social tags, security headers and markup."
      },
      {
        "q": "Do I need to pass every SEO check?",
        "a": "No. Some checks are rules of thumb, like word_count and dom_size. Treat them as prompts to look at the page, not as laws. Fix the failing basics first, then decide on the rest."
      },
      {
        "q": "How do I see only SEO failures for my page?",
        "a": "Run seoscore on the URL and read the SEO family card. The issue list puts failing checks first, and each row shows the check id from this guide."
      },
      {
        "q": "Can I run these checks in CI?",
        "a": "Yes. Use the -fail-under flag, for example seoscore -fail-under 80 https://example.com. The command exits with code 1 when the score is below your number."
      }
    ]
  },
  {
    "slug": "seo-vs-aeo-vs-geo",
    "title": "SEO vs AEO vs GEO: The Difference in Plain Words",
    "description": "SEO, AEO and GEO each target a different reader: a search crawler, an answer box, or a generative AI model. See what wins in each and how to check them.",
    "datePublished": "2026-10-14",
    "dateModified": "2026-10-14",
    "html": `<h1>SEO vs AEO vs GEO: The Difference in Plain Words</h1>
<p>You have probably seen three new acronyms on every marketing blog. They sound like hype, but each one describes a real reader with real needs. This post explains them in plain words, then shows which findable checks map to each.</p>
<h2 id="what-is-seo">What is SEO?</h2>
<p>SEO is search engine optimization. You make a page easy for a search crawler to fetch, understand and rank. The reader is a bot first, a person second. Good SEO means a clear title, a meta description, clean headings, working links and a page served over https.</p>
<p>Google's own <a href="https://developers.google.com/search/docs/fundamentals/seo-starter-guide" rel="noopener">SEO starter guide</a> covers these basics. Most of them are boring on purpose. You write a good title, you describe the page, you link to related pages, and you keep the HTML tidy.</p>
<h2 id="what-is-aeo">What is AEO?</h2>
<p>AEO is answer engine optimization. You shape a page so an answer box, a voice assistant or a chat tool can lift a short, correct answer from it. The reader is a program that wants one clean passage, not a whole page.</p>
<p>That means question headings, a direct answer right under each one, and lists, tables and steps that are easy to parse. You also add structured data such as <a href="https://schema.org/FAQPage" rel="noopener">FAQPage</a> so a machine knows what each block is.</p>
<h2 id="what-is-geo">What is GEO?</h2>
<p>GEO is generative engine optimization. You make a page worth citing when a model writes an answer from many sources. The reader is a language model, or the retrieval step in front of it, that pulls passages from many pages and links some of them.</p>
<p>The <a href="https://arxiv.org/abs/2311.09735" rel="noopener">GEO paper</a> by Aggarwal et al. tested ways to raise visibility in generative answers. In their tests, adding citations, quotations and statistics helped. The lesson for you is simple: give a model facts it can quote, and say who wrote them.</p>
<h2 id="how-do-seo-aeo-and-geo-compare">How do SEO, AEO and GEO compare?</h2>
<p>They share a base but aim at different readers. SEO targets crawlers and rankings. AEO targets answer boxes. GEO targets generative models that cite sources. The table below shows what each one targets, what tends to win, and a few real findable check ids for each.</p>
<table><thead><tr><th></th><th>SEO</th><th>AEO</th><th>GEO</th></tr></thead><tbody><tr><td>Reader</td><td>Search crawler</td><td>Answer box, assistant</td><td>Generative model</td></tr><tr><td>Goal</td><td>Rank and get indexed</td><td>Be the quoted answer</td><td>Be the cited source</td></tr><tr><td>What wins</td><td>Clear title, meta, headings, fast clean HTML</td><td>Direct answers, question headings, structured data</td><td>Citations, statistics, authorship, entity markup</td></tr><tr><td>Example checks</td><td><code>title_length</code>, <code>desc_length</code>, <code>h1_exists</code>, <code>canonical</code></td><td><code>aeo_faq</code>, <code>aeo_direct_answers</code>, <code>aeo_question_headings</code></td><td><code>geo_citations</code>, <code>geo_statistics_density</code>, <code>geo_author_schema</code></td></tr><tr><td>Weight in score</td><td>50%</td><td>25%</td><td>25%</td></tr></tbody></table>
<p>Notice the overlap. A <code>jsonld</code> block helps SEO, and <code>aeo_structured_data</code> reads the same JSON-LD for AEO. Good structure pays off three times.</p>
<h2 id="how-do-you-check-each-one-with-findable">How do you check each one with findable?</h2>
<p>findable runs every check on a page and groups the results into the three families. You install it, point it at a URL and read the failures. Each check has a stable id, so you can search for it, track it and fix it.</p>
<pre><code>go install github.com/nadimtuhin/seoscore-cli/cmd/seoscore@latest
seoscore https://example.com -k &quot;your keyphrase&quot;</code></pre>
<p>Here are a few ids you will meet first. For SEO, <code>title_length</code> wants 30 to 60 characters and <code>desc_length</code> wants 120 to 160. For AEO, <code>aeo_direct_answers</code> looks for a paragraph of 30 to 80 words. For GEO, <code>geo_citations</code> wants at least one external link and <code>geo_passage_citability</code> wants a paragraph of 40 to 120 words.</p>
<h2 id="why-does-the-order-matter">Why does the order matter?</h2>
<p>Work in the order SEO, then AEO, then GEO. A model cannot cite a page it never fetched, and an answer box cannot quote a page that is blocked or hidden. That is why SEO carries half of the score.</p>
<p>After the base is clean, answer-ready structure is cheap to add. Then GEO signals, such as a named author in <code>geo_author_credentials</code> and <code>sameAs</code> links in <code>geo_knowledge_graph</code>, tell a model who stands behind the text. Google also explains how its AI features use your pages in <a href="https://developers.google.com/search/docs/appearance/ai-features" rel="noopener">AI features and your website</a>.</p>
<h2 id="key-takeaways">Key takeaways</h2>
<ul><li>SEO targets crawlers, AEO targets answer boxes, GEO targets generative models.</li><li>They stack. Fix SEO first, then add answer structure, then citation signals.</li><li>findable scores SEO 50%, AEO 25% and GEO 25%, and every check has a stable id.</li><li>Run <code>seoscore &lt;url&gt;</code> to see your failures, ranked, in one pass.</li><li>A short summary section helps you too: <code>aeo_summary</code> looks for the words &quot;key takeaway&quot;, &quot;summary&quot; or &quot;conclusion&quot;.</li></ul>`,
    "faq": [
      {
        "q": "Is GEO a replacement for SEO?",
        "a": "No. GEO builds on SEO. A page that crawlers cannot fetch or index will not be cited by anyone. Fix the SEO basics first, then add answer and citation signals on top."
      },
      {
        "q": "How does findable weigh SEO, AEO and GEO?",
        "a": "The score is SEO 50%, AEO 25% and GEO 25%. SEO carries the most weight because every other layer depends on a page being crawlable and readable."
      },
      {
        "q": "Can I check all three with one command?",
        "a": "Yes. Run seoscore with a URL and it audits all three families in one pass. Add -site to also check robots.txt and llms.txt."
      }
    ]
  },
  {
    "slug": "why-ai-answer-engines-skip-your-site",
    "title": "Why AI Answer Engines Skip Your Site and How to Fix It",
    "description": "ChatGPT, Perplexity and Claude may ignore your pages for five fixable reasons, from robots.txt blocks to missing answers. Learn each cause and its check id.",
    "datePublished": "2026-10-14",
    "dateModified": "2026-10-14",
    "html": `<h1>Why AI Answer Engines Skip Your Site and How to Fix It</h1>
<p>You wrote the page. It ranks fine. Then you ask ChatGPT, Perplexity or Claude about your topic and your site never shows up. This post lists the five most common causes and the findable check that catches each one.</p>
<h2 id="why-does-robots-txt-keep-ai-bots-out">Why does robots.txt keep AI bots out?</h2>
<p>It is easy to block AI crawlers without noticing. A broad <code>Disallow</code> rule, or a rule copied from a template, tells bots like GPTBot, ClaudeBot and PerplexityBot to stay away. If a bot cannot fetch your page, it cannot use it.</p>
<p>Check your file at <code>/robots.txt</code>. The <a href="https://developers.google.com/search/docs/crawling-indexing/robots/intro" rel="noopener">Google robots.txt intro</a> explains the format. Each bot has its own docs: <a href="https://platform.openai.com/docs/bots" rel="noopener">OpenAI crawlers</a>, <a href="https://support.anthropic.com/en/articles/8896518-does-anthropic-crawl-data-from-the-web-and-how-can-site-owners-block-the-crawler" rel="noopener">ClaudeBot</a>, <a href="https://docs.perplexity.ai/guides/bots" rel="noopener">PerplexityBot</a> and <a href="https://developers.google.com/search/docs/crawling-indexing/google-common-crawlers" rel="noopener">Google-Extended</a>.</p>
<p>This snippet allows all four bots:</p>
<pre><code>User-agent: GPTBot
Allow: /

User-agent: ClaudeBot
Allow: /

User-agent: PerplexityBot
Allow: /

User-agent: Google-Extended
Allow: /

User-agent: *
Allow: /</code></pre>
<p>If you do want to block one of them, that is your call. Just make it a choice. With <code>-site</code>, findable reports <code>robots-txt</code>, <code>robots-page</code>, <code>robots-gptbot</code>, <code>robots-claudebot</code>, <code>robots-perplexitybot</code> and <code>robots-google-extended</code>.</p>
<h2 id="why-does-a-missing-llms-txt-matter">Why does a missing llms.txt matter?</h2>
<p>An <code>llms.txt</code> file is a plain Markdown file at your site root that points language models to your best pages. It is a proposal, so no engine promises to read it. It costs little to add and gives tools a clean map of your site.</p>
<p>Read the <a href="https://llmstxt.org/" rel="noopener">llms.txt proposal</a> for the format. findable checks for it with the <code>llms-txt</code> site check. It is unscored, so a miss will not drop your grade. The check only tests that the file exists; it does not parse the Markdown or follow the links in it.</p>
<h2 id="why-can-t-bots-read-javascript-only-content">Why can't bots read JavaScript-only content?</h2>
<p>If your text only appears after a script runs, some crawlers will see an empty shell. Google can render JavaScript, but you should not assume every AI bot does. Plain HTML in the first server response is the safe path for every reader.</p>
<p>Read <a href="https://developers.google.com/search/docs/crawling-indexing/javascript/javascript-seo-basics" rel="noopener">Google's JavaScript SEO basics</a> for how rendering works. In findable, a thin shell shows up as a failing <code>word_count</code> (300+ words), a low <code>text_html_ratio</code> (15% or more) and a missing <code>h1_exists</code>. If those fail on a page that looks full in your browser, you are probably rendering on the client. Move the key text into server-rendered HTML.</p>
<h2 id="why-does-missing-structured-data-hurt">Why does missing structured data hurt?</h2>
<p>Structured data tells a machine what a block of your page is: an article, an FAQ, a person, an organization. Without it, a model has to guess what you meant. With it, your meaning is explicit, and the machine does not need to infer it.</p>
<p>The <a href="https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data" rel="noopener">structured data intro</a> shows how to add JSON-LD, and <a href="https://schema.org/" rel="noopener">schema.org</a> lists the types. findable checks <code>jsonld</code>, <code>aeo_structured_data</code>, <code>aeo_faq</code>, <code>geo_org_schema</code>, <code>geo_author_schema</code>, <code>geo_knowledge_graph</code> and <code>geo_entity_markup</code>. Start with one JSON-LD block, then add FAQPage and Organization.</p>
<h2 id="why-don-t-engines-find-a-direct-answer-on-your-page">Why don't engines find a direct answer on your page?</h2>
<p>A short passage that stands alone is the easiest thing to quote. If your page buries the point under a long intro, there is nothing clean to lift. Put the answer first, right under a question heading, and add detail after it.</p>
<p>Write a paragraph of 30 to 80 words for <code>aeo_direct_answers</code>, and a question heading for <code>aeo_question_headings</code>. Add <code>aeo_lists</code>, <code>aeo_tables</code> and <code>aeo_numbered_steps</code> where they fit. For citability, <code>geo_passage_citability</code> wants a paragraph of 40 to 120 words, and <code>geo_citations</code> wants at least one external link.</p>
<h2 id="how-do-you-find-all-five-problems-at-once">How do you find all five problems at once?</h2>
<p>Run one command against your home page with <code>-site</code>. You get the page checks plus the robots.txt and llms.txt results in one report. Each bot gets its own row, so you can see that GPTBot is allowed while ClaudeBot is blocked.</p>
<pre><code>seoscore https://example.com -site</code></pre>
<p>Fix the robots.txt rules first, since a blocked bot ignores everything else. Then server-render your text, add JSON-LD, and rewrite your top sections as question plus answer.</p>
<h2 id="key-takeaways">Key takeaways</h2>
<ul><li>Bots skip sites for five common reasons, and all five are fixable.</li><li>Check robots.txt first. Allow GPTBot, ClaudeBot, PerplexityBot and Google-Extended if you want to be found.</li><li>Add <code>/llms.txt</code>, put your text in server-rendered HTML, and add JSON-LD.</li><li>Lead each section with a direct answer.</li><li>Run <code>seoscore &lt;url&gt; -site</code> to see every one of these as a check id.</li></ul>`,
    "faq": [
      {
        "q": "Does blocking GPTBot hide my site from ChatGPT?",
        "a": "It stops the GPTBot crawler from fetching your pages. OpenAI documents separate bots for training and for search, so read their bot page and decide each one on purpose."
      },
      {
        "q": "Is llms.txt required?",
        "a": "No. It is a proposal, not a standard that every engine reads. It is cheap to add, and findable reports it as the llms-txt site check."
      },
      {
        "q": "How do I see which bots my robots.txt blocks?",
        "a": "Run seoscore with -site. It reports robots-gptbot, robots-claudebot, robots-perplexitybot and robots-google-extended as separate results."
      }
    ]
  }
];

// src/site.ts
var AI_CRAWLERS = ["GPTBot", "ClaudeBot", "PerplexityBot", "Google-Extended"];
var PAGE_CRAWLER = "*";
var MAX_SITE_FILE_BYTES = 512 * 1024;
var MAX_ROBOTS_WILDCARDS = 10;
function parseRobots(robots) {
  const groups = [];
  let inAgents = false;
  for (const raw of robots.split("\n")) {
    const line = raw.split("#")[0] ?? "";
    const colon = line.indexOf(":");
    if (colon < 0) continue;
    const key = line.slice(0, colon).trim().toLowerCase();
    const value = line.slice(colon + 1).trim();
    const current = groups[groups.length - 1];
    if (key === "user-agent") {
      const token = (value.split("/")[0] ?? "").trim().toLowerCase();
      if (!inAgents || !current) groups.push({ agents: [token], rules: [] });
      else current.agents.push(token);
      inAgents = true;
    } else if (key === "allow" || key === "disallow") {
      inAgents = false;
      if (current && value) current.rules.push({ allow: key === "allow", pattern: value });
    }
  }
  return groups;
}
__name(parseRobots, "parseRobots");
function rulesFor(groups, agent) {
  const name = agent.toLowerCase();
  const named = groups.filter((g) => g.agents.includes(name)).flatMap((g) => g.rules);
  if (groups.some((g) => g.agents.includes(name))) return named;
  return groups.filter((g) => g.agents.includes(PAGE_CRAWLER)).flatMap((g) => g.rules);
}
__name(rulesFor, "rulesFor");
function patternMatches(pattern, path) {
  if (pattern.split("*").length - 1 > MAX_ROBOTS_WILDCARDS) return false;
  const anchored = pattern.endsWith("$");
  const body = anchored ? pattern.slice(0, -1) : pattern;
  const expr = body.split("*").map((part) => part.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join(".*");
  return new RegExp(`^${expr}${anchored ? "$" : ""}`).test(path);
}
__name(patternMatches, "patternMatches");
function robotsAllows(robots, agent, path) {
  let best = { allow: true, pattern: "" };
  let bestLen = -1;
  for (const rule of rulesFor(parseRobots(robots), agent)) {
    if (!patternMatches(rule.pattern, path)) continue;
    const len = rule.pattern.length;
    if (len > bestLen || len === bestLen && rule.allow) {
      best = rule;
      bestLen = len;
    }
  }
  return best.allow;
}
__name(robotsAllows, "robotsAllows");
function siteCategory(pagePath, robots, llms) {
  const robotsText = robots ?? "";
  const checks = [
    { id: "robots-txt", label: "robots.txt is present", pass: robots !== null },
    {
      id: "robots-page",
      label: "robots.txt lets crawlers fetch this page",
      pass: robotsAllows(robotsText, PAGE_CRAWLER, pagePath)
    },
    ...AI_CRAWLERS.map((bot) => ({
      id: `robots-${bot.toLowerCase()}`,
      label: `robots.txt lets ${bot} fetch this page`,
      pass: robotsAllows(robotsText, bot, pagePath)
    })),
    { id: "llms-txt", label: "/llms.txt is present", pass: llms !== null }
  ];
  return { score: null, passed: checks.filter((c) => c.pass).length, total: checks.length, checks };
}
__name(siteCategory, "siteCategory");
async function optionalText(target, self) {
  try {
    const page2 = await guardedFetch(target, {}, self);
    if (page2.status !== 200) return null;
    return page2.body.slice(0, MAX_SITE_FILE_BYTES);
  } catch {
    return null;
  }
}
__name(optionalText, "optionalText");
async function siteChecks(pageUrl, self) {
  const u = new URL(pageUrl);
  const [robots, llms] = await Promise.all([
    optionalText(`${u.origin}/robots.txt`, self),
    optionalText(`${u.origin}/llms.txt`, self)
  ]);
  return siteCategory(u.pathname + u.search, robots, llms);
}
__name(siteChecks, "siteChecks");

// src/learn.ts
var AUTHOR = "Nadim Tuhin";
var AUTHOR_URL = "https://github.com/nadimtuhin";
var REPO_URL = "https://github.com/nadimtuhin/findable";
var SITE_NAME = "findable";
var INDEX_TITLE = "Learn SEO, AEO and GEO with findable";
var INDEX_DESCRIPTION = "Plain guides to SEO, AEO and GEO and to every check findable runs, so you can see why a page is hard to find and fix it.";
var LEARN_STYLE = "body{font:16px/1.6 system-ui,sans-serif;max-width:46rem;margin:0 auto;padding:1rem;color:#1a1a1a;background:#fff}a{color:#0b5cad}nav,footer{font-size:.9rem}nav a{margin-right:1rem}h1{line-height:1.2}table{border-collapse:collapse;display:block;overflow-x:auto}th,td{border:1px solid #ccc;padding:.3rem .5rem;text-align:left}pre{overflow-x:auto;background:#f4f4f4;padding:.75rem}code{background:#f4f4f4;padding:0 .2rem}pre code{padding:0}.byline{color:#555}details{margin:1rem 0}dt{font-weight:600;margin-top:1rem}dd{margin:.25rem 0 0}";
var escapeHtml = /* @__PURE__ */ __name((text) => text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;"), "escapeHtml");
var jsonLd = /* @__PURE__ */ __name((data) => `<script type="application/ld+json">${JSON.stringify(data).replace(/</g, "\\u003c")}<\/script>`, "jsonLd");
function page(input) {
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
<footer><p><a href="/">Audit your own page with ${SITE_NAME}</a> &middot; <a href="${REPO_URL}" rel="noopener">Source on GitHub</a></p></footer>
</body>
</html>
`;
}
__name(page, "page");
function renderLearnNotFound(origin) {
  return page({
    title: "Page not found | findable",
    description: "That learn page does not exist. Go back to the findable guides or run an audit on the home page.",
    path: "/learn",
    origin,
    ogType: "website",
    structuredData: [],
    body: `<h1>Page not found</h1>
<p>There is no guide at this address. See <a href="/learn">all findable guides</a> or <a href="/">audit a page</a>.</p>`
  });
}
__name(renderLearnNotFound, "renderLearnNotFound");
function postList(posts) {
  const items = posts.map(
    (post) => `<li><a href="/learn/${post.slug}">${escapeHtml(post.title)}</a><br>${escapeHtml(post.description)}</li>`
  );
  return `<ul>${items.join("")}</ul>`;
}
__name(postList, "postList");
var INDEX_FAQ = [
  {
    q: "What is findable?",
    a: "findable checks whether a page can be found by Google and by AI answer engines. It scores SEO, AEO and GEO and lists every failing check by id so you know what to fix."
  },
  {
    q: "Where should I start?",
    a: "Read the comparison of SEO, AEO and GEO first, then run an audit on your own page and open the check guide for each failing id."
  },
  {
    q: "Who writes these guides?",
    a: `${AUTHOR} writes them. ${AUTHOR} builds findable, so every guide describes checks that exist in the code.`
  }
];
function latestModified() {
  return learnPosts.map((p) => p.dateModified).sort()[learnPosts.length - 1] ?? "";
}
__name(latestModified, "latestModified");
function indexBody() {
  const updated = latestModified();
  const start = ["seo-vs-aeo-vs-geo", "findable-in-15-minutes", "why-ai-answer-engines-skip-your-site"];
  const steps = start.map((slug) => learnPosts.find((p) => p.slug === slug)).filter((p) => p !== void 0).map((p) => `<li><a href="/learn/${p.slug}">${escapeHtml(p.title)}</a></li>`);
  return [
    `<h1>${escapeHtml(INDEX_TITLE)}</h1>`,
    `<p class="byline">By <a href="${AUTHOR_URL}" rel="noopener">${AUTHOR}</a>. Updated <time datetime="${updated}">${updated}</time>.</p>`,
    `<p>${escapeHtml(INDEX_DESCRIPTION)} findable is a tool that audits one page for SEO, AEO and GEO. You can read the guides in any order, and each one links to the others.</p>`,
    `<h2 id="start-here">Where should you start?</h2>
<ol>${steps.join("")}</ol>`,
    `<h2 id="all-guides">All guides</h2>
${postList(learnPosts)}`,
    `<h2 id="quick-start">How do you run an audit from the command line?</h2>
<p>Pass a URL to the seoscore command. The <code>-fail-under</code> flag fails the run when the score drops below a number, which is useful in CI.</p>
<pre><code>seoscore -fail-under 80 https://example.com</code></pre>`,
    SCORING_NOTE,
    `<h2 id="summary">Summary</h2>
<ul><li>SEO is how crawlers find and index a page.</li><li>AEO is how easily an answer can be lifted from it.</li><li>GEO is how likely a generative model is to cite it.</li><li>Every check has a stable id, so you can look up a failure and fix it.</li></ul>`,
    faqSection(INDEX_FAQ).replace("<dl>", "<details open><summary>Questions and answers</summary><dl>").replace("</dl>", "</dl></details>")
  ].join("\n");
}
__name(indexBody, "indexBody");
function renderLearnIndex(origin) {
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
        url: `${origin}/learn`
      }),
      identityGraph(origin),
      faqLd(INDEX_FAQ)
    ],
    body: indexBody()
  });
}
__name(renderLearnIndex, "renderLearnIndex");
function tableOfContents(html) {
  const entries = [...html.matchAll(/<h2 id="([^"]+)">([\s\S]*?)<\/h2>/g)];
  if (entries.length === 0) return "";
  const items = entries.map((m) => `<li><a href="#${m[1]}">${m[2]}</a></li>`);
  return `<details><summary>Contents</summary><ol>${items.join("")}</ol></details>`;
}
__name(tableOfContents, "tableOfContents");
function faqSection(faq) {
  if (faq.length === 0) return "";
  const entries = faq.map((f) => `<dt>${escapeHtml(f.q)}</dt><dd>${escapeHtml(f.a)}</dd>`);
  return `<section>
<h2 id="faq">FAQ</h2>
<dl>${entries.join("")}</dl>
</section>`;
}
__name(faqSection, "faqSection");
function byline(post) {
  const published = `<time datetime="${post.datePublished}">${post.datePublished}</time>`;
  const modified = `<time datetime="${post.dateModified}">${post.dateModified}</time>`;
  return `<p class="byline">By <a href="${AUTHOR_URL}" rel="noopener">${AUTHOR}</a>. Published ${published}. Updated ${modified}.</p>`;
}
__name(byline, "byline");
var SCORING_NOTE = `<section>
<h2 id="how-findable-scores">How does findable score a page?</h2>
<table><thead><tr><th>Family</th><th>Weight</th><th>Question it asks</th></tr></thead><tbody>
<tr><td>SEO</td><td>50%</td><td>Can crawlers read and index this page?</td></tr>
<tr><td>AEO</td><td>25%</td><td>Is it easy to lift an answer from this page?</td></tr>
<tr><td>GEO</td><td>25%</td><td>Does the page give a generative model a reason to cite it?</td></tr>
</tbody></table>
<p>The hosted audit waits up to ${FETCH_TIMEOUT_MS / 1e3} s for a page and fetches at most ${MAX_PAGE_BYTES / (1024 * 1024)} MB. It reads robots.txt and llms.txt up to ${MAX_SITE_FILE_BYTES / 1024} KB each.</p>
</section>`;
function related(post) {
  const others = learnPosts.filter((p) => p.slug !== post.slug);
  return `<section>
<h2>More guides</h2>
${postList(others)}
</section>`;
}
__name(related, "related");
var personId = /* @__PURE__ */ __name((origin) => `${origin}/#author`, "personId");
var orgId = /* @__PURE__ */ __name((origin) => `${origin}/#organization`, "orgId");
var identityEntities = /* @__PURE__ */ __name((origin) => [
  { "@type": "Person", "@id": personId(origin), name: AUTHOR, url: AUTHOR_URL, sameAs: [AUTHOR_URL] },
  { "@type": "Organization", "@id": orgId(origin), name: SITE_NAME, url: origin, sameAs: [REPO_URL] }
], "identityEntities");
var identityGraph = /* @__PURE__ */ __name((origin) => jsonLd({ "@context": "https://schema.org", "@graph": identityEntities(origin) }), "identityGraph");
var faqLd = /* @__PURE__ */ __name((faq) => jsonLd({
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: faq.map((f) => ({
    "@type": "Question",
    name: f.q,
    acceptedAnswer: { "@type": "Answer", text: f.a }
  }))
}), "faqLd");
function postStructuredData(post, origin) {
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
        publisher: { "@id": orgId(origin) }
      },
      ...identityEntities(origin)
    ]
  });
  return post.faq.length === 0 ? [graph] : [graph, faqLd(post.faq)];
}
__name(postStructuredData, "postStructuredData");
function renderLearnPost(slug, origin) {
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
    related(post)
  ];
  return page({
    title: post.title,
    description: post.description,
    path: `/learn/${post.slug}`,
    origin,
    ogType: "article",
    structuredData: postStructuredData(post, origin),
    body: `<article>
${body.filter(Boolean).join("\n")}
</article>`
  });
}
__name(renderLearnPost, "renderLearnPost");

// src/ratelimit.ts
function fixedWindowLimiter(limit, windowMs, now = Date.now) {
  let windowStart = now();
  const counts = /* @__PURE__ */ new Map();
  return (key) => {
    const t = now();
    if (t - windowStart >= windowMs) {
      windowStart = t;
      counts.clear();
    }
    const n = (counts.get(key) ?? 0) + 1;
    counts.set(key, n);
    return n <= limit;
  };
}
__name(fixedWindowLimiter, "fixedWindowLimiter");

// src/site-files.ts
var HEAD_SLOT_MARKER = "<!--slot:head-->";
var AI_BOTS = ["GPTBot", "ClaudeBot", "PerplexityBot", "Google-Extended"];
var REPO = "https://github.com/nadimtuhin/findable";
var AUTHOR2 = { name: "Nadim Tuhin", url: "https://github.com/nadimtuhin" };
var TAGLINE = "findable: is your site findable by Google and by AI answer engines?";
var SUMMARY = "Open-source auditor that scores a page for traditional SEO, Answer Engine Optimization (AEO) and Generative Engine Optimization (GEO). The web API runs 54 page checks (26 SEO, 14 AEO, 14 GEO); the Go CLI runs 72 (36 SEO, 17 AEO, 19 GEO). Both add 7 separate site checks (robots.txt, AI crawler access, llms.txt).";
var HOW_TO_USE = [
  "- CLI: `go install github.com/nadimtuhin/findable/cmd/findable@latest`, then `findable https://example.com`.",
  "- MCP: `findable -mcp`, or `npx -y findable-mcp`. Tools: audit_page, audit_html_content, crawl_site, generate_csv_report.",
  "- API: `POST /api/scan` with JSON body `{url, keyphrase?, site?}`; `site: true` adds the site checks.",
  `- Source and license (MIT): ${REPO}`
];
function robotsTxt(origin) {
  return [
    ...["*", ...AI_BOTS].map((agent) => `User-agent: ${agent}`),
    "Allow: /",
    "Disallow: /api/",
    "",
    `Sitemap: ${origin}/sitemap.xml`,
    ""
  ].join("\n");
}
__name(robotsTxt, "robotsTxt");
function escapeXml(text) {
  return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}
__name(escapeXml, "escapeXml");
function sitemapXml(origin, paths) {
  const urls = paths.map((p) => `  <url><loc>${escapeXml(origin + p)}</loc></url>`).join("\n");
  return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls}
</urlset>
`;
}
__name(sitemapXml, "sitemapXml");
function llmsTxt(origin, pages) {
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
    ""
  ].join("\n");
}
__name(llmsTxt, "llmsTxt");
function llmsFullTxt(origin, pages) {
  const sections = pages.map((p) => `### ${p.title}

${origin}${p.path}

${p.description}
`);
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
    ...sections
  ].join("\n");
}
__name(llmsFullTxt, "llmsFullTxt");
function jsonLdScript(data) {
  const json = JSON.stringify(data).replace(/</g, "\\u003c");
  return `<script type="application/ld+json">${json}<\/script>`;
}
__name(jsonLdScript, "jsonLdScript");
function headSlot(origin, faq = []) {
  const context = "https://schema.org";
  const blocks = [
    {
      "@context": context,
      "@type": "SoftwareApplication",
      name: "findable",
      applicationCategory: "DeveloperApplication",
      operatingSystem: "Any",
      url: origin,
      offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
      sameAs: [REPO],
      author: { "@type": "Person", name: AUTHOR2.name, url: AUTHOR2.url }
    },
    { "@context": context, "@type": "Person", name: AUTHOR2.name, url: AUTHOR2.url, sameAs: [AUTHOR2.url] },
    {
      "@context": context,
      "@type": "Organization",
      name: "findable",
      url: origin,
      sameAs: [REPO],
      founder: { "@type": "Person", name: AUTHOR2.name, url: AUTHOR2.url }
    }
  ];
  if (faq.length > 0) {
    blocks.push({
      "@context": context,
      "@type": "FAQPage",
      mainEntity: faq.map((f) => ({
        "@type": "Question",
        name: f.question,
        acceptedAnswer: { "@type": "Answer", text: f.answer }
      }))
    });
  }
  return blocks.map(jsonLdScript).join("\n");
}
__name(headSlot, "headSlot");
function injectHead(html, slot) {
  return html.replace(HEAD_SLOT_MARKER, () => slot);
}
__name(injectHead, "injectHead");

// src/index.ts
var CORS_HEADERS = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
  "Access-Control-Allow-Headers": "Content-Type"
};
var MAX_ANALYZED_HTML_CHARS = 1024 * 1024;
var fallbackAllowScan = fixedWindowLimiter(10, 6e4);
async function allowScan(env, key) {
  if (!env.SCAN_LIMITER) return fallbackAllowScan(key);
  return (await env.SCAN_LIMITER.limit({ key })).success;
}
__name(allowScan, "allowScan");
var PUBLIC_ORIGIN = "https://findable.piply.xyz";
var PAGE_HTML = injectHead(UI_HTML, headSlot(PUBLIC_ORIGIN));
var uiPolicy;
async function sha256Source(text) {
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(text)));
  return `'sha256-${btoa(String.fromCharCode(...digest))}'`;
}
__name(sha256Source, "sha256Source");
async function hashesOf(tag) {
  const blocks = [...PAGE_HTML.matchAll(new RegExp(`<${tag}(?:\\s[^>]*)?>([\\s\\S]*?)</${tag}>`, "g"))];
  return (await Promise.all(blocks.map((m) => sha256Source(m[1] ?? "")))).join(" ");
}
__name(hashesOf, "hashesOf");
async function buildUiPolicy() {
  return [
    "default-src 'none'",
    `script-src ${await hashesOf("script")} https://static.cloudflareinsights.com`,
    "style-src 'unsafe-inline'",
    `style-src-elem ${await hashesOf("style")}`,
    "style-src-attr 'unsafe-inline'",
    "connect-src 'self' https://cloudflareinsights.com",
    "img-src 'self' data:",
    "base-uri 'none'",
    "form-action 'self'",
    "frame-ancestors 'none'"
  ].join("; ");
}
__name(buildUiPolicy, "buildUiPolicy");
function securityHeaders(csp) {
  return {
    "Content-Security-Policy": csp,
    "X-Content-Type-Options": "nosniff",
    "Strict-Transport-Security": "max-age=31536000; includeSubDomains",
    "X-Frame-Options": "DENY",
    "Referrer-Policy": "strict-origin-when-cross-origin"
  };
}
__name(securityHeaders, "securityHeaders");
var HOME_PAGE = {
  path: "/",
  title: "findable",
  description: "Scan any URL for SEO, AEO and GEO signals and get a score with the fixes that matter."
};
var LEARN_PAGES = [
  {
    path: "/learn",
    title: "Learn findable",
    description: "Plain guides to SEO, AEO and GEO, and how to check them with findable."
  },
  ...learnPosts.map((p) => ({ path: `/learn/${p.slug}`, title: p.title, description: p.description }))
];
var SITE_PAGES = [HOME_PAGE, ...LEARN_PAGES];
var SITE_FILES = {
  "/robots.txt": /* @__PURE__ */ __name((o) => ({ type: "text/plain", body: robotsTxt(o) }), "/robots.txt"),
  "/sitemap.xml": /* @__PURE__ */ __name((o) => ({
    type: "application/xml",
    body: sitemapXml(
      o,
      SITE_PAGES.map((p) => p.path)
    )
  }), "/sitemap.xml"),
  "/llms.txt": /* @__PURE__ */ __name((o) => ({ type: "text/plain", body: llmsTxt(o, SITE_PAGES) }), "/llms.txt"),
  "/llms-full.txt": /* @__PURE__ */ __name((o) => ({ type: "text/plain", body: llmsFullTxt(o, SITE_PAGES) }), "/llms-full.txt")
};
var learnPolicy;
async function buildLearnPolicy() {
  return [
    "default-src 'none'",
    `style-src ${await sha256Source(LEARN_STYLE)}`,
    "base-uri 'none'",
    "form-action 'none'",
    "frame-ancestors 'none'"
  ].join("; ");
}
__name(buildLearnPolicy, "buildLearnPolicy");
async function learnResponse(pathname, origin) {
  const slug = pathname.slice("/learn".length).replace(/^\/|\/$/g, "");
  const html = slug ? renderLearnPost(slug, origin) : renderLearnIndex(origin);
  learnPolicy ??= buildLearnPolicy();
  return new Response(html ?? renderLearnNotFound(origin), {
    status: html ? 200 : 404,
    headers: { "Content-Type": "text/html; charset=utf-8", ...securityHeaders(await learnPolicy) }
  });
}
__name(learnResponse, "learnResponse");
var OWN_HOSTS = ["findable.piply.xyz", "seoscore.piply.xyz"];
var worker = {
  async fetch(request, env = {}) {
    const url = new URL(request.url);
    if (request.method === "OPTIONS") {
      return new Response(null, { headers: CORS_HEADERS });
    }
    if (url.pathname === "/") {
      uiPolicy ??= buildUiPolicy();
      return new Response(PAGE_HTML, {
        headers: {
          ...CORS_HEADERS,
          "Content-Type": "text/html; charset=utf-8",
          ...securityHeaders(await uiPolicy)
        }
      });
    }
    const siteFile = request.method === "GET" ? SITE_FILES[url.pathname] : void 0;
    if (siteFile) {
      const { type, body } = siteFile(url.origin);
      return new Response(body, { headers: { ...CORS_HEADERS, "Content-Type": `${type}; charset=utf-8` } });
    }
    if (request.method === "GET" && (url.pathname === "/learn" || url.pathname.startsWith("/learn/"))) {
      return learnResponse(url.pathname, url.origin);
    }
    if (url.pathname === "/health") {
      return new Response(JSON.stringify({ status: "ok", name: "findable", version: "1.0.0" }), {
        headers: { ...CORS_HEADERS, "Content-Type": "application/json" }
      });
    }
    if (url.pathname === "/api/scan" && request.method === "POST") {
      try {
        const key = rateLimitKey(request.headers.get("CF-Connecting-IP") ?? "unknown");
        if (!await allowScan(env, key)) {
          return new Response(JSON.stringify({ error: "too many scans, try again in a minute" }), {
            status: 429,
            headers: { ...CORS_HEADERS, "Content-Type": "application/json" }
          });
        }
        const body = await readScanBody(request);
        if (typeof body === "string") return badRequest(body);
        const targetUrl = withScheme(body.url);
        const self = { hosts: [url.hostname, ...OWN_HOSTS], handle: /* @__PURE__ */ __name((r) => worker.fetch(r, env), "handle") };
        const scanRes = await auditUrl(targetUrl, self, body.keyphrase?.trim());
        if (body.site === true) scanRes.site = await siteChecks(targetUrl, self);
        return new Response(JSON.stringify(scanRes), {
          headers: { ...CORS_HEADERS, "Content-Type": "application/json" }
        });
      } catch (err) {
        const message = err instanceof Error && err.message ? err.message : "audit failed";
        return new Response(JSON.stringify({ error: message }), {
          status: err instanceof BlockedTargetError ? 400 : 500,
          headers: { ...CORS_HEADERS, "Content-Type": "application/json" }
        });
      }
    }
    return new Response("Not Found", { status: 404, headers: CORS_HEADERS });
  }
};
var index_default = worker;
var BARE_HOST = /^[a-z0-9-]+(\.[a-z0-9-]+)+(:\d+)?([/?#].*)?$/i;
function withScheme(target) {
  const t = target.trim();
  return BARE_HOST.test(t) ? `https://${t}` : t;
}
__name(withScheme, "withScheme");
function badRequest(error) {
  return new Response(JSON.stringify({ error }), {
    status: 400,
    headers: { ...CORS_HEADERS, "Content-Type": "application/json" }
  });
}
__name(badRequest, "badRequest");
async function readScanBody(request) {
  const body = await request.json().catch(() => null);
  if (typeof body !== "object" || body === null || Array.isArray(body)) return "body must be a JSON object";
  const { url, keyphrase, site } = body;
  if (typeof url !== "string" || !url.trim()) return "missing url";
  if (keyphrase !== void 0 && typeof keyphrase !== "string") return "keyphrase must be a string";
  return { url, keyphrase, site: site === true };
}
__name(readScanBody, "readScanBody");
function rateLimitKey(ip) {
  if (!ip.includes(":")) return ip;
  const [head = "", tail] = ip.toLowerCase().split("::");
  const headGroups = head ? head.split(":") : [];
  const tailGroups = tail ? tail.split(":") : [];
  const zeros = tail === void 0 ? [] : Array(8 - headGroups.length - tailGroups.length).fill("0");
  const prefix = [...headGroups, ...zeros, ...tailGroups].slice(0, 4);
  return `${prefix.map((g) => Number.parseInt(g, 16).toString(16)).join(":")}::/64`;
}
__name(rateLimitKey, "rateLimitKey");
async function auditUrl(targetUrl, self, keyphrase) {
  const page2 = await guardedFetch(
    targetUrl,
    {
      "User-Agent": "Mozilla/5.0 (compatible; findable-worker/1.0; +https://findable.piply.xyz)",
      Accept: "text/html,application/xhtml+xml"
    },
    self
  );
  const scan = analyzeHtml(page2.url, page2.body.slice(0, MAX_ANALYZED_HTML_CHARS), page2.headers, keyphrase);
  return { ...scan, url: targetUrl };
}
__name(auditUrl, "auditUrl");
function firstGroup(match) {
  return match?.[1]?.trim() ?? "";
}
__name(firstGroup, "firstGroup");
function analyzeHtml(targetUrl, rawHtml, headers, keyphrase) {
  const lowerHtml = rawHtml.toLowerCase();
  const parsedUrl = new URL(targetUrl);
  const titleMatch = rawHtml.match(/<title[^>]*>([^<]*)<\/title>/i);
  const title = firstGroup(titleMatch);
  const descMatch = rawHtml.match(/<meta[^>]+name=["']description["'][^>]+content=["']([^"']*)["']/i) || rawHtml.match(/<meta[^>]+content=["']([^"']*)["'][^>]+name=["']description["']/i);
  const metaDesc = firstGroup(descMatch);
  const canonicalMatch = rawHtml.match(/<link[^>]+rel=["']canonical["'][^>]+href=["']([^"']*)["']/i) || rawHtml.match(/<link[^>]+href=["']([^"']*)["'][^>]+rel=["']canonical["']/i);
  const canonical = firstGroup(canonicalMatch);
  const viewportMatch = rawHtml.match(/<meta[^>]+name=["']viewport["']/i);
  const charsetMatch = rawHtml.match(/<meta[^>]+charset=["']?utf-8/i) || lowerHtml.includes("charset=utf-8");
  const langMatch = rawHtml.match(/<html[^>]+lang=["']([^"']+)["']/i);
  const doctypeMatch = rawHtml.startsWith("<!DOCTYPE html") || rawHtml.startsWith("<!doctype html");
  const robotsMatch = rawHtml.match(/<meta[^>]+name=["']robots["'][^>]+content=["']([^"']*)["']/i);
  const robotsMeta = (robotsMatch?.[1] ?? "").toLowerCase();
  const h1Matches = [...rawHtml.matchAll(/<h1[^>]*>([\s\S]*?)<\/h1>/gi)];
  const h2Matches = [...rawHtml.matchAll(/<h2[^>]*>([\s\S]*?)<\/h2>/gi)];
  const visibleText = rawHtml.replace(/<script[\s\S]*?<\/script>/gi, " ").replace(/<style[\s\S]*?<\/style>/gi, " ").replace(/<svg[\s\S]*?<\/svg>/gi, " ").replace(/<[^>]+>/g, " ").replace(/\s+/g, " ").trim();
  const words = visibleText.split(/\s+/).filter(Boolean);
  const wordCount = words.length;
  const hasDirectAnswer = [...rawHtml.matchAll(/<p\b[^>]*>([\s\S]*?)<\/p>/gi)].some((m) => {
    const n = (m[1] ?? "").replace(/<[^>]+>/g, " ").split(/\s+/).filter(Boolean).length;
    return n >= 30 && n <= 80;
  });
  const ogTitle = rawHtml.match(/<meta[^>]+property=["']og:title["']/i);
  const ogDesc = rawHtml.match(/<meta[^>]+property=["']og:description["']/i);
  const ogImage = rawHtml.match(/<meta[^>]+property=["']og:image["']/i);
  const ogUrl = rawHtml.match(/<meta[^>]+property=["']og:url["']/i);
  const twitterCard = rawHtml.match(/<meta[^>]+name=["']twitter:card["']/i);
  const linkMatches = [...rawHtml.matchAll(/<a[^>]+href=["']([^"']+)["']/gi)];
  const internalLinks = [];
  const externalLinks = [];
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
    } catch {
    }
  }
  const jsonLdMatches = [
    ...rawHtml.matchAll(/<script[^>]+type=["']application\/ld\+json["'][^>]*>([\s\S]*?)<\/script>/gi)
  ];
  const schemaTypes = [];
  const sameAsLinks = [];
  let hasFAQ = false;
  let hasPerson = false;
  let hasOrg = false;
  for (const jm of jsonLdMatches) {
    try {
      const parsed = JSON.parse(firstGroup(jm));
      const scanObj = /* @__PURE__ */ __name((value) => {
        if (!value || typeof value !== "object") return;
        const o = value;
        if (o["@type"]) {
          const t = String(o["@type"]);
          schemaTypes.push(t);
          if (t.toLowerCase() === "faqpage") hasFAQ = true;
          if (t.toLowerCase() === "person") hasPerson = true;
          if (["organization", "bakery", "localbusiness"].includes(t.toLowerCase())) hasOrg = true;
        }
        if (o.sameAs) {
          if (Array.isArray(o.sameAs)) sameAsLinks.push(...o.sameAs.map(String));
          else sameAsLinks.push(String(o.sameAs));
        }
        const graph = o["@graph"];
        if (Array.isArray(graph)) graph.forEach(scanObj);
      }, "scanObj");
      scanObj(parsed);
    } catch {
    }
  }
  const hasDl = lowerHtml.includes("<dl") && lowerHtml.includes("<dt") && lowerHtml.includes("<dd");
  const hasOl = lowerHtml.includes("<ol");
  const hasUl = lowerHtml.includes("<ul");
  const hasTable = lowerHtml.includes("<table");
  const hasCode = lowerHtml.includes("<code") || lowerHtml.includes("<pre");
  const hasDetails = lowerHtml.includes("<details") && lowerHtml.includes("<summary");
  const hasBlockquote = lowerHtml.includes("<blockquote");
  const hasTime = lowerHtml.includes("<time");
  const statsMatches = visibleText.match(
    /(?:\$\s*\d+(?:\.\d+)?|\b\d+(?:\.\d+)?\s*(?:%|ms|s|MB|GB|KB|px|M|B|k|users|queries|hours|pages)\b)/g
  ) || [];
  const lowerText = visibleText.toLowerCase();
  const youCount = (lowerText.match(/\b(you|your|we)\b/g) || []).length;
  const seoChecks = [
    { id: "title_exists", label: title ? `Title tag found: "${title}"` : "No title tag found", pass: Boolean(title) },
    {
      id: "title_length",
      label: `Title length: ${title.length} characters (aim for 30-60)`,
      pass: title.length >= 30 && title.length <= 60
    },
    {
      id: "desc_exists",
      label: metaDesc ? `Meta description found (${metaDesc.length} chars)` : "Missing meta description",
      pass: Boolean(metaDesc)
    },
    {
      id: "desc_length",
      label: `Description length: ${metaDesc.length} chars (aim for 120-160)`,
      pass: metaDesc.length >= 120 && metaDesc.length <= 160
    },
    { id: "viewport", label: "Viewport meta tag found (mobile-friendly)", pass: Boolean(viewportMatch) },
    { id: "charset", label: "UTF-8 charset declared", pass: Boolean(charsetMatch) },
    { id: "lang_attr", label: "Language attribute set on HTML tag", pass: Boolean(langMatch) },
    { id: "html_doctype", label: "<!DOCTYPE html> declared", pass: Boolean(doctypeMatch) },
    { id: "robots_meta", label: "No robots meta tag blocking crawling", pass: !robotsMeta.includes("noindex") },
    {
      id: "h1_exists",
      label: h1Matches.length === 1 ? "Single H1 tag found" : `${h1Matches.length} H1 tags found (exactly 1 recommended)`,
      pass: h1Matches.length === 1
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
      pass: Boolean(headers.get("strict-transport-security"))
    },
    {
      id: "x_content_type",
      label: "X-Content-Type-Options: nosniff header present",
      pass: (headers.get("x-content-type-options") || "").includes("nosniff")
    },
    { id: "x_frame_options", label: "X-Frame-Options header present", pass: Boolean(headers.get("x-frame-options")) },
    {
      id: "csp",
      label: "Content-Security-Policy header declared",
      pass: Boolean(headers.get("content-security-policy"))
    },
    { id: "jsonld", label: `${jsonLdMatches.length} JSON-LD schema block(s) found`, pass: jsonLdMatches.length > 0 }
  ];
  if (keyphrase) {
    const kp = keyphrase.toLowerCase();
    seoChecks.push({
      id: "keyphrase_in_title",
      label: `Focus keyphrase "${keyphrase}" in title tag`,
      pass: title.toLowerCase().includes(kp)
    });
    seoChecks.push({
      id: "keyphrase_in_desc",
      label: `Focus keyphrase "${keyphrase}" in meta description`,
      pass: metaDesc.toLowerCase().includes(kp)
    });
    seoChecks.push({
      id: "keyphrase_in_h1",
      label: `Focus keyphrase "${keyphrase}" in H1 heading`,
      pass: h1Matches.some((m) => (m[1] ?? "").toLowerCase().includes(kp))
    });
    seoChecks.push({
      id: "keyphrase_in_content",
      label: `Focus keyphrase "${keyphrase}" appears in body copy`,
      pass: lowerText.includes(kp)
    });
  }
  const aeoChecks = [
    { id: "aeo_structured_data", label: "Structured data found for AI extraction", pass: jsonLdMatches.length > 0 },
    { id: "aeo_faq", label: "FAQ schema or Q&A section detected", pass: hasFAQ },
    { id: "aeo_direct_answers", label: "Concise answer paragraphs (30-80 words) found", pass: hasDirectAnswer },
    {
      id: "aeo_question_headings",
      label: "Question-style headings found (What/How/Why/?)",
      pass: h2Matches.some((m) => /\b(what|how|why|\?)\b/i.test(m[1] ?? ""))
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
      pass: youCount >= 3
    },
    {
      id: "aeo_summary",
      label: "Summary or Key Takeaways section present",
      pass: /key takeaway|summary|conclusion/i.test(lowerText)
    },
    {
      id: "aeo_clean_url",
      label: "Clean lowercase URL structure",
      pass: !parsedUrl.pathname.includes("_") && parsedUrl.pathname === parsedUrl.pathname.toLowerCase()
    },
    { id: "aeo_ai_robots", label: "No AI-blocking directives detected", pass: !robotsMeta.includes("noindex") }
  ];
  const geoChecks = [
    {
      id: "geo_statistics_density",
      label: `${statsMatches.length} quantitative data points / metrics found`,
      pass: statsMatches.length >= 3
    },
    { id: "geo_data_viz", label: "Structured data tables or charts found", pass: hasTable },
    {
      id: "geo_comparison",
      label: "Comparative sections or tables detected",
      pass: hasTable || /\bvs\b/i.test(lowerText)
    },
    { id: "geo_expert_quotes", label: "Expert quotes or blockquotes found", pass: hasBlockquote },
    {
      id: "geo_citations",
      label: `${externalLinks.length} external source citations found`,
      pass: externalLinks.length > 0
    },
    {
      id: "geo_definitions",
      label: "Clear concept definitions found ('X is a...')",
      pass: /\bis (a|the|defined as)\b/i.test(lowerText)
    },
    { id: "geo_author_schema", label: "Author schema markup found", pass: hasPerson },
    {
      id: "geo_author_credentials",
      label: "Visible author byline with credentials or dates",
      pass: /\bby\s+[a-z]/i.test(lowerText) || hasPerson
    },
    { id: "geo_org_schema", label: "Organization or Business schema markup declared", pass: hasOrg },
    {
      id: "geo_knowledge_graph",
      label: `${sameAsLinks.length} authoritative sameAs links in schema`,
      pass: sameAsLinks.length > 0
    },
    {
      id: "geo_trust",
      label: "Trust signals (About and Contact links) present",
      pass: internalLinks.some((l) => /about/i.test(l)) && internalLinks.some((l) => /contact/i.test(l))
    },
    {
      id: "geo_freshness",
      label: "Content freshness signals (dates/time elements) present",
      pass: hasTime || /\b202[4-9]\b/.test(lowerText)
    },
    { id: "geo_time_element", label: "<time> machine-readable tags present", pass: hasTime },
    { id: "geo_content_length", label: `${wordCount} words content depth (300+ recommended)`, pass: wordCount >= 300 }
  ];
  const calcCat = /* @__PURE__ */ __name((checks) => {
    const passed = checks.filter((c) => c.pass).length;
    const total = checks.length;
    const score = total > 0 ? Math.round(passed / total * 100) : 0;
    return { score, passed, total, checks };
  }, "calcCat");
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
    scannedAt: (/* @__PURE__ */ new Date()).toISOString()
  };
}
__name(analyzeHtml, "analyzeHtml");
export {
  index_default as default
};
//# sourceMappingURL=index.js.map
