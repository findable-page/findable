# Roadmap: make findable a must-have CLI

Goal: a developer or agent auditing a site's search and AI (GEO) findability should reach for `findable` because it does what a chat assistant cannot. A chat LLM can explain best practice and read one pasted page. It cannot crawl 500 URLs, see what GPTBot actually gets back from a WAF, compare raw HTML with the rendered DOM, measure bytes and latency, diff this week's audit with last week's, or fail a build. Every item below is judged on that line: **does it need live access, scale, determinism or memory?** If not, it is a doc page, not a feature.

Status: in progress. Phases 0.1–0.3 are partial: audit/bots/mcp/serve dispatch, metadata, fix-list output and strict config loading exist, but remaining subcommands, full evidence and config-backed gates/crawl/bots/ignore semantics do not. Phase 1.1 fix-list logic and Phase 1.2 bot probe are partial. Phase 1.3–1.8 and Phase 1 exit criteria remain open.

## 1. Where findable is today (audit of this repo)

Strengths worth keeping:

- One Go binary, one dependency (`golang.org/x/net/html`), fully offline for local files. `go install` and done.
- 160 page checks (SEO 98, AEO 35, GEO 27) in `internal/audit/checks_*.go`, driven by one signal pass (`signals.go: extractSignals`). Data-driven pass/fail cases in `testdata/check_cases*.json` run by `internal/audit/check_cases_test.go`.
- Robots.txt parser with longest-match and wildcard support (`internal/audit/site.go`), split into AI search crawlers and AI training crawlers.
- SSRF guard (`internal/fetch/guard.go`): public hosts, ports 80/443, 5 redirects. This is a real moat for the MCP and server modes.
- Output: CSV, JSON, SARIF (`internal/report/`), `-fail-under`, a composite GitHub Action (`action.yml`), MCP over stdio (`internal/mcpserver/`, 4 tools), local HTTP API (`internal/server/`).

Gaps found (file references are where the fix lands):

| # | Gap | Where | Why it matters |
|---|---|---|---|
| G1 | A `Check` is `{ID, Label, Pass, Applicable}`. No severity, no evidence (which element, which value), no fix text. | `internal/audit/scan.go` | Output is a checklist, not a fix list. Users cannot act without re-reading the page. This is the biggest gap. |
| G2 | Every check weighs the same: score = passed/total per family. `aeo_conversational` counts as much as `canonical_self`. | `engine.go: buildCategory` | Scores move on trivia; users chase the wrong fixes. |
| G3 | No cross-page analysis. Crawl returns `[]Page` and each page is judged alone. No duplicate titles/descriptions, no broken internal links, no redirect chains, no canonical target checks, no orphan pages, no sitemap vs crawl diff. | `internal/fetch/fetch.go: Crawl`, `sitemap.go` | This is exactly the work a chat LLM cannot do and Screaming Frog charges for. |
| G4 | Crawl is sequential, does not respect robots.txt for itself, does not discover the sitemap from robots.txt, drops non-200 pages silently to stderr. | `fetch.go`, `sitemap.go` | Slow on real sites; errors (404, 5xx) are findings, not noise. |
| G5 | AI crawler checks read robots.txt only. They miss `X-Robots-Tag` headers, the user-triggered fetchers (`ChatGPT-User`, `Perplexity-User`, `Claude-User`), `Bingbot` (ChatGPT search leans on Bing), and the most common real-world block: a WAF/CDN returning 403 or a challenge page to bot user agents. `aeo_ai_robots` only looks at `<meta name=robots>` noindex. | `site.go`, `checks_aeo.go:80`, `checks_geo.go:66` | "robots.txt allows GPTBot" is false comfort when Cloudflare blocks it. Only a live probe can tell. |
| G6 | No JS rendering and no raw-vs-rendered comparison. | `fetch.go: fetchAndAudit` | OpenAI, Anthropic and Perplexity crawlers are reported to read raw HTML without running JS. Content that only exists after hydration is invisible to them. |
| G7 | `llms.txt` is presence-only. No format validation, no link check, no generator. No robots.txt or JSON-LD generator either. | `site.go: SiteCategory` | Competitors (GeoReady `geo fix`) write these files. |
| G8 | No baseline, no compare, no history. `-fail-under` is overall-score only. | `cmd/findable/main.go: exitIfBelow` | CI needs "no new failures" and per-family gates, not one number. |
| G9 | No Markdown or HTML report. CSV is the default. | `internal/report/format.go` | Markdown is what goes into PR comments and agent context; HTML is what gets sent to a client. |
| G10 | CWV is an empty category. No timing, no page weight, no compression/cache checks. `html_weight` and `inline_css_size` are the only size checks. | `engine.go` (`CWV: Category{Checks: []Check{}}`) | Measured numbers are something an LLM cannot produce. |
| G11 | MCP exposes 4 tools; no fix generation, no compare, no bot probe. | `internal/mcpserver/mcp_tools.go` | Agents are the fastest-growing user; MCP is the distribution channel. |
| G12 | Error and gate failures share exit code 1 (`fail()` and `exitIfBelow`). | `cmd/findable/main.go` | CI cannot tell "site is bad" from "network broke". |
| G13 | No evidence tier on checks. llms.txt, "conversational tone", "summary section" are heuristics; Google says it does not use llms.txt for AI features. | all `checks_*.go` | Honesty is a differentiator: competitors score llms.txt as if it ranks. |
| G14 | README says the offline engine ignores `-k`; CHANGELOG says the Go engine runs the four keyphrase checks (`keyphrase_in_*` exist in `checks_seo.go`). | `README.md` | Small, fix in passing. |
| G15 | Single flat flag set; `main.go` is already near the `funlen` limit (80 lines) in `.golangci.yml`. | `cmd/findable/main.go` | New features need subcommands. |

## 2. What the research says (pwm/Perplexity, 2026)

Ranking/citation factors with the strongest evidence:

1. **Crawler access, per bot.** OpenAI: opting out of `OAI-SearchBot` removes a site from ChatGPT Search; `GPTBot` is training only; `ChatGPT-User` is user-triggered and not used for search eligibility. Perplexity: `PerplexityBot` indexes, `Perplexity-User` fetches on request and generally ignores robots.txt. Anthropic: `ClaudeBot` (training), `Claude-SearchBot` (search), `Claude-User` (on request). Sources: developers.openai.com/api/docs/bots, docs.perplexity.ai/docs/resources/perplexity-crawlers, support.claude.com article 8896518. WAF/bot-management blocks are called out as a common failure.
2. **Content in the initial HTML.** Independent 2026 tests report OAI-SearchBot, ChatGPT-User, Anthropic crawlers and PerplexityBot fetch raw HTML and do not execute JS. Googlebot renders.
3. **Google AI Overviews / AI Mode ride on normal Google indexing.** Google says no special AI markup, no llms.txt, no required chunking (developers.google.com/search/docs/appearance/ai-features).
4. **ChatGPT search uses Bing among other providers.** Bing indexability and IndexNow help discovery for that path; IndexNow is not a ranking signal.
5. **Structured data that matches visible content, consistent entity identity (`sameAs`, Organization, author), dates, original data and citations.** Supported by Google quality guidance; no published weight.
6. **llms.txt:** no major engine documents reading site-owner llms.txt for ranking or citation. It can help user-triggered retrieval of large docs sites. Treat as optional hygiene.

Competitors:

| Tool | Strength | Gap findable can own |
|---|---|---|
| Screaming Frog CLI, Sitebulb | Deep crawl, redirects, duplicates, canonicals | Paid, desktop-first, no AI crawler/WAF probe, no GEO checks, weak CI gating |
| Lighthouse CI, Unlighthouse | Rendered perf, budgets, CI exit codes | Shallow SEO, zero AI/GEO, page-oriented |
| seonaut | Self-hosted crawler dashboard | No CI assertions, no AI checks |
| Ahrefs / Semrush Site Audit | 170+ issues, scheduled crawls | SaaS, not a local gate, little AI readiness |
| Otterly, Profound, Peec, AthenaHQ, Goodie | Measure whether LLMs cite a brand for prompts | Do not find technical defects; nondeterministic; SaaS |
| GeoReady / geo-optimizer-skill (Python, MIT) | 27 AI bots, 47 checks, `geo fix` writes robots/llms.txt/JSON-LD, MCP (12 tools), GitHub Action, 8 output formats | Closest rival. Scores llms.txt as a factor; Python install; no SEO crawl depth (duplicates, redirects, broken links); no raw-vs-rendered diff |
| seo-geo-audit, aeo-audit (Node) | Combined SEO + AI crawler checks | Small, thin CI story |

Positioning: **the only single-binary tool that combines a technical SEO crawl, a live per-bot access probe, a raw-vs-rendered diff and GEO content checks, with evidence-tiered scoring, deterministic CI gates and an MCP server.** Answer-visibility monitoring (the Otterly lane) stays out of the CI gate; see Phase 3.

## 3. Design rules for everything below

- **Deterministic first.** Anything that gates CI must give the same result for the same site state. LLM-judged checks never gate.
- **Evidence tiers on every check:** `documented` (vendor docs say so), `reported` (independent tests), `heuristic` (our judgment). Weights follow tiers. Reports show the tier. This is the anti-snake-oil stance and a selling point.
- **No new Go dependency without a written reason.** Rendering shells out to an installed Chrome (`--headless --dump-dom`) instead of adding chromedp. Timing uses `net/http/httptrace`. Config is JSON via `encoding/json`.
- **Backward compatible.** `findable <url>` and every current flag keep working. New features arrive as subcommands.
- **Guard stays on.** Every new fetch (bot probe, link check, subresources, Chrome) goes through the SSRF guard unless `-allow-private`. Chrome rendering needs special care (see 1.4).
- **Exit codes:** `0` pass, `1` gate failed, `2` usage, `3` runtime/network error. (Fixes G12; document in README.)

## 4. Phase 0: foundations (small, unblocks everything)

### 0.1 Subcommands

```
findable [legacy flags] <url|file>          # unchanged, = findable audit
findable audit   [flags] <url|file>
findable crawl   [flags] <url>              # Phase 1.3
findable bots    [flags] <url>              # Phase 1.2
findable render  [flags] <url>              # Phase 1.4
findable fix     [flags] <url|report.json>  # Phase 1.6
findable compare <old.json> <new.json>      # Phase 1.5
findable mcp | findable serve <addr>
```

- Dispatch on `os.Args[1]` against a fixed map; anything else falls through to the legacy flag parser. One `flag.FlagSet` per subcommand; shared flags (`-o`, `-format`, `-allow-private`, `-q`, `-config`) registered by one helper.
- Files: `cmd/findable/main.go` (dispatch only), new `cmd/findable/cmd_audit.go`, `cmd_crawl.go`, etc. as each lands. No CLI framework.
- Tests: `cmd/findable/dispatch_test.go`: legacy `findable -q x.html` and `findable audit -q x.html` produce identical JSON; unknown subcommand exits 2.

### 0.2 Check metadata registry (fixes G1, G2, G13 without touching 160 call sites)

Keep `l.add(id, label, pass)`. Add one table keyed by check ID:

```go
// internal/audit/meta.go
type Meta struct {
    Severity Severity // critical | high | medium | low | info
    Tier     Tier     // documented | reported | heuristic
    Fix      string   // one imperative sentence, may use {evidence}
    Doc      string   // /learn slug
}
var registry = map[string]Meta{ "canonical_self": {...}, ... }
```

- `Check` gains `Severity`, `Tier`, `Fix`, `Evidence []Evidence` (JSON `omitempty`, so existing JSON consumers keep working). `Evidence{Selector, Value, URL string}`.
- `buildCategory` attaches meta by ID. Scores become weighted: `critical 8, high 4, medium 2, low 1, info 0`, multiplied by tier factor `documented 1.0, reported 0.75, heuristic 0.5`. Old flat score stays available as `scoreModel: "flat"` behind `-score-model flat` for one minor release, then goes.
- Evidence capture: add a selector builder to `signals.go` (`nodePath(n)` producing `main > article > img:nth-of-type(3)`), and fill `Evidence` for the checks where location matters: images (`img_alt`, `img_dimensions`, `img_lazy_loading`, `img_modern_format`), headings (`h1_exists`, `heading_hierarchy`), links (`empty_hrefs`, `link_noopener`, `mixed_content`), head tags (`title_*`, `desc_*`, `canonical*`, `og_*`, `hreflang_*`), JSON-LD (`jsonld_valid`, `faq_valid`, `article_*`: block index + JSON pointer, e.g. `script[type="application/ld+json"]:nth-of-type(2) #/author`). Cap 20 evidence items per check.
- Tests:
  - `meta_test.go`: run `Analyze` over every file in `testdata/site/`, collect all emitted IDs, assert each has a registry entry with non-empty `Fix` (prevents drift when someone adds a check).
  - Extend `check_cases*.json` schema with optional `"evidence": ["selector"]` and assert on it for 10 representative checks.
  - Weighted score test with hand-computed expected values (independent of the implementation).

### 0.3 Config file

`findable.json` (or `-config path`) in the repo root, read with `encoding/json`:

```json
{
  "baseline": ".findable/baseline.json",
  "gate": { "overall": 70, "seo": 80, "geo": 60, "noNewFailures": true, "maxSeverity": "high" },
  "ignore": [ { "check": "aeo_conversational" }, { "check": "img_modern_format", "url": "/legacy/*" } ],
  "crawl": { "limit": 200, "concurrency": 4, "include": ["/docs/*"], "exclude": ["/tag/*"] },
  "bots": ["OAI-SearchBot", "PerplexityBot", "Claude-SearchBot", "Googlebot", "Bingbot"]
}
```

- Flags override config. Unknown keys are an error (use `json.Decoder.DisallowUnknownFields`).
- File: `internal/config/config.go` + test with a fixture.

## 5. Phase 1: highest leverage, buildable now

Order is by leverage per day of work. Each item is one vertical slice: CLI, engine, report, test.

### 1.1 Prioritized fix list (the default output) — partial (terminal list; report-first output remains)

The terminal output and every report lead with **what to fix, in order**, not a score dump.

- Ranking key: `severity weight × tier factor × pages affected` (for crawls), ties broken by check ID for stable output.
- Terminal:

```
findable  https://example.com  overall 64 (SEO 71  AEO 58  GEO 55)

 1 CRITICAL  documented  OAI-SearchBot gets 403 from the server (robots.txt allows it)
             fix: allow the OAI-SearchBot user agent in your WAF/bot rules
             evidence: GET / as OAI-SearchBot -> 403 cf-mitigated: challenge
 2 HIGH      documented  12 pages share the title "Home | Example"           [12 pages]
             fix: give each page a unique <title>; first: /about /pricing /blog
 3 HIGH      reported    Main content missing from raw HTML (82% of words appear only after JS)
             fix: server-render <main>; AI crawlers read raw HTML
 ...
 31 more (medium 12, low 19). Full list: findable-report-example.com-2026-11-08.md
```

- Flags: `-top N` (default 10), `-min-severity low|medium|high|critical`.
- Files: `internal/report/fixlist.go` (pure: `[]audit.Page -> []Finding`), `cmd/findable/print.go` replaces `printFailures`.
- Test: `fixlist_test.go` with a hand-built `[]Page` and expected order; golden file `testdata/golden/fixlist.txt`.

### 1.2 `findable bots`: live AI crawler access truth (fixes G5) — partial (live CLI probe; audit integration and IP verification remain)

What a chat LLM cannot do: request the page as each bot and report what came back.

```
findable bots [-bots list] [-path /x] [-format ...] <url>
```

For each bot in the list (default: `OAI-SearchBot, ChatGPT-User, GPTBot, PerplexityBot, Perplexity-User, Claude-SearchBot, Claude-User, ClaudeBot, Googlebot, Google-Extended, Bingbot, Applebot-Extended, meta-externalagent, CCBot`):

| Column | Source |
|---|---|
| robots.txt verdict for this path | existing `RobotsAllows` |
| HTTP status as this UA | live GET with the bot's published UA string |
| challenge/block detected | 403/429/503, `cf-mitigated`, `server: cloudflare` + challenge markers, body under 2 KB vs browser body |
| content parity | word count and `<title>` vs the browser-UA fetch (flag if under 50%) |
| `X-Robots-Tag` for this UA | header parse incl. `X-Robots-Tag: GPTBot: noindex` form |
| meta robots for this UA | `<meta name="robots">` and `<meta name="GPTBot">` etc. |

- Verdict per bot: `allowed`, `blocked-by-robots`, `blocked-by-server`, `degraded-content`, `noindex`. Purpose per bot (`search`, `user-fetch`, `training`) printed next to it, with the documented consequence ("blocking OAI-SearchBot removes you from ChatGPT Search").
- Honest caveat in output: requests come from your IP, not the vendor's ranges; WAFs that verify bot IPs will block a spoofed UA that the real bot would pass, and vice versa. Report `unverifiable` when the response differs only by IP-verification markers.
- Also folded into `audit -site` and `crawl` as site checks: `bot_<name>_server`, `x_robots_tag_noindex`, plus the new user-fetch bots in `site.go`.
- Bot list with UA strings, purpose and doc URL lives in one table: `internal/audit/bots.go` (moves `aiSearchCrawlers`/`trainingCrawlers` there).
- Politeness: one request per bot, sequential, 250 ms apart, `-bots` to narrow.
- Tests: `internal/fetch/bots_test.go` with `httptest.Server` that returns 403 + `cf-mitigated` for `GPTBot`, a 1 KB challenge page for `PerplexityBot`, `X-Robots-Tag: ClaudeBot: noindex`, and 200 for the rest; assert the verdict table. Parser test for `X-Robots-Tag` variants in `internal/audit/xrobots_test.go`.

### 1.3 `findable crawl`: site graph and cross-page checks (fixes G3, G4)

```
findable crawl [-limit 200] [-concurrency 4] [-sitemap auto|url|off]
               [-include glob] [-exclude glob] [-respect-robots=true] [-check-external=false] <url>
```

Crawler changes (`internal/fetch/crawl.go`, split out of `fetch.go`):

- Bounded worker pool (`-concurrency`, default 4, max 16), per-host delay honouring `Crawl-delay`.
- Discover sitemaps from `robots.txt` `Sitemap:` lines when `-sitemap auto`; merge sitemap URLs into the frontier.
- Record for every URL: final status, redirect chain (hop list with status codes, via a `CheckRedirect` hook that appends to a slice instead of following blindly), TTFB, bytes, content type, canonical, `hreflang` set, meta robots, `X-Robots-Tag`, outlinks with anchor text, depth from start.
- Non-HTML and non-200 responses become graph nodes, not stderr lines.
- Respect robots.txt for `findable`'s own UA by default.

New site-level checks (`internal/audit/sitechecks.go`, pure functions over a `SiteGraph`):

| ID | Severity / tier | Rule |
|---|---|---|
| `dup_title` | high / documented | same normalized title on 2+ indexable pages |
| `dup_description` | medium / documented | same meta description on 2+ indexable pages |
| `dup_content` | medium / reported | SimHash of main text, Hamming distance <= 3 |
| `broken_internal_link` | high / documented | link target 4xx/5xx; evidence = source page + selector + anchor text |
| `redirect_chain` | medium / documented | internal link hits 2+ hops, or any loop |
| `redirect_to_non200` | high / documented | chain ends in non-200 |
| `canonical_target_bad` | high / documented | canonical points to non-200, redirected, noindexed or robots-blocked URL |
| `canonical_cluster` | info | groups pages by canonical target |
| `hreflang_reciprocal` | high / documented | A lists B, B does not list A; `x-default` present; targets 200 |
| `sitemap_non200` | medium / documented | sitemap lists redirected/404/noindex URLs |
| `sitemap_missing` | low / reported | indexable crawled pages absent from sitemap |
| `orphan_page` | medium / reported | in sitemap, no internal inlinks |
| `deep_page` | low / heuristic | depth > 4 clicks |
| `noindex_in_sitemap` | high / documented | conflict |
| `mixed_signals` | high / documented | canonical says A, `og:url` says B, sitemap lists C |

Output: per-page rows as today plus a `site` section with these findings, plus `findable-graph-<host>.json` (nodes + edges) for anyone who wants to plot it. Markdown report gets a "Site issues" section above page issues.

Tests:
- `internal/fetch/crawl_test.go`: `httptest.Server` fixture site (12 pages) under `testdata/crawlsite/` with deliberate dup titles, a 404 link, a 3-hop redirect, a canonical to a redirect, a non-reciprocal hreflang pair, an orphan in the sitemap, a `Disallow`ed path. Assert exact finding set.
- `internal/audit/sitechecks_test.go`: table tests on hand-built `SiteGraph` values (no network).
- Concurrency: `-race` already in `make test`; add a test that a crawl with `-concurrency 8` yields the same finding set as `-concurrency 1`.

### 1.4 `findable render`: raw vs rendered diff (fixes G6)

```
findable render [-chrome path] [-wait 3s] <url>
findable audit -render <url>          # audit both, report the diff
```

- Find Chrome/Chromium (`-chrome`, `$FINDABLE_CHROME`, then the usual paths and `PATH` names). If absent: exit 3 with an install hint; never silently skip.
- Run `chrome --headless=new --disable-gpu --dump-dom --virtual-time-budget=<wait ms> --user-agent=<findable UA> <url>` with `exec.CommandContext` and a timeout. No new Go dependency.
- Compare raw HTML signals vs rendered DOM signals (both through `extractSignals`):
  - share of visible words present only after render (`js_content_share`), title/meta/canonical/JSON-LD/h1 that differ or appear only after render, internal links only after render (crawlers that do not render cannot discover them).
- New checks: `raw_html_has_main_content` (high / reported: < 20% JS-only words), `raw_html_has_jsonld` (high / reported), `raw_html_has_canonical` (high / documented: Google warns against JS-changed canonicals), `raw_html_has_links` (medium / reported).
- Security: Chrome does its own networking and bypasses the Go SSRF guard. Without `-allow-private`, resolve the host first with the guard's `PublicAddr` and pass `--host-resolver-rules="MAP * ~NOTFOUND, EXCLUDE <host>"` so Chrome can only reach the audited host; subresources from other hosts fail, which is acceptable for a text diff. `-serve` and MCP never enable rendering unless the operator passes `-allow-render` at startup. Document in `SECURITY.md`.
- Files: `internal/render/chrome.go`, `internal/audit/renderdiff.go` (pure).
- Tests: `renderdiff_test.go` with paired fixtures `testdata/render/spa-raw.html` / `spa-rendered.html` (no Chrome needed). One integration test behind `//go:build chrome` for local runs.

### 1.5 Baseline, compare and CI gates (fixes G8, G12)

```
findable audit|crawl ... -baseline .findable/baseline.json          # compare and gate
findable audit|crawl ... -write-baseline .findable/baseline.json    # record
findable compare old.json new.json [-format md|json]
```

Gate flags (all also in config `gate`):

| Flag | Meaning |
|---|---|
| `-fail-under N` | existing: overall per page |
| `-fail-under-seo N`, `-fail-under-aeo N`, `-fail-under-geo N` | per family |
| `-fail-on-new` | any check failing now that passed (or did not exist) in the baseline |
| `-fail-on critical\|high\|medium` | any failing finding at or above this severity |
| `-max-regression N` | overall dropped more than N points on any page |

- Diff unit: `(url, checkID)`. Output lists `new failures`, `fixed`, `still failing`, score deltas per family and per page, pages added/removed.
- Baseline format = the JSON report plus `{"findableVersion", "registryHash"}`; if the registry changed (checks added), new checks are reported as `new check` and never count for `-fail-on-new`. That keeps upgrades from breaking CI.
- Exit codes per section 3.
- `action.yml`: new inputs `baseline`, `fail-on`, `fail-on-new`; post the Markdown diff as a job summary (`$GITHUB_STEP_SUMMARY`).
- Files: `internal/report/diff.go` (pure), `cmd/findable/gate.go` (replaces `exitIfBelow`).
- Tests: `diff_test.go` with two hand-written JSON reports in `testdata/diff/`; `gate_test.go` for each flag; `cmd/findable` test that a runtime error returns 3 and a gate failure returns 1 (run the binary via `os/exec` against `httptest`).

### 1.6 `findable fix`: generated fixes (fixes G7)

```
findable fix [-only robots,llms,jsonld,meta] [-out dir] [-diff] <url | report.json>
```

Writes files to `-out` (default `findable-fixes/`), never edits a site. `-diff` prints a unified diff against the live file.

| Artifact | How it is built |
|---|---|
| `robots.txt` | Starts from the live file, preserves all existing groups, adds explicit `Allow` groups for search and user-fetch bots that are blocked, leaves training-bot policy as the user set it (comment explains the choice), adds `Sitemap:` if missing. Policy flag `-training allow\|block\|keep` (default keep). |
| `llms.txt` | From crawl/sitemap: H1 site name, `>` summary from home meta description, sections grouped by top path segment, `- [title](url): description` lines, capped at 200 links. Header comment states it is optional and not a ranking signal. |
| `jsonld/<page-slug>.json` | Organization + WebSite for home (name, url, logo, `sameAs` from existing links to known social/Wikipedia/Wikidata hosts); Article/BlogPosting for pages with author and date signals; BreadcrumbList from URL path and titles; FAQPage only when question headings with answers exist on the page (Google restricts FAQ rich results; tier note in the file). Every value comes from the page; missing values are left as `"TODO: ..."` so nothing is invented. |
| `head-snippets/<page-slug>.html` | Missing title/description/canonical/OG/twitter tags, filled from page content, marked for review. |
| `FIXES.md` | The prioritized fix list (1.1) with links to the files above. |

Also add `llms.txt` validation to site checks: `llms_txt_format` (H1 first, links resolve, under 2 MB; low / heuristic), `llms_txt_links_ok` (low / heuristic).

- Files: `internal/fix/robots.go`, `llms.go`, `jsonld.go`, `head.go`; all pure (input: signals/graph; output: bytes).
- Tests: golden files in `testdata/fix/` per generator; a round-trip test that the generated robots.txt, parsed by `RobotsAllows`, allows every search bot and keeps every original `Disallow`; generated JSON-LD passes our own schema checks (`jsonld_valid`, `article_*` etc. are N/A or pass).

### 1.7 Markdown and HTML reports (fixes G9)

- `-format md`: fix list first, then site issues, then per-page tables, then passed checks collapsed in `<details>`. Fits a PR comment (truncate at 60 KB with a pointer to the full file).
- `-format html`: single self-contained file (inline CSS, no JS needed, prints well), same order, with evidence selectors and a score-delta block when a baseline is given. Reuse styles from `internal/server/ui.go` where possible.
- Multiple formats in one run: `-format md,json,sarif`.
- SARIF: add `level` from severity, `properties.tier`, and `help.text` from `Fix` (today `internal/report/sarif.go` emits rule ID, label and one fixed `level` for every result).
- Files: `internal/report/markdown.go`, `html.go` (`html/template`), `format.go`.
- Tests: golden files; HTML test asserts it parses with `x/net/html` and contains no `<script>`.

### 1.8 MCP upgrade (fixes G11)

New tools in `internal/mcpserver/mcp_tools.go`, each a thin wrapper over the Phase 1 functions:

| Tool | Returns |
|---|---|
| `findable_fix_list` | prioritized findings (1.1) for a URL, top N |
| `check_bot_access` | the 1.2 verdict table |
| `crawl_site` (existing, upgraded) | site findings (1.3), limit max 100 |
| `compare_reports` | 1.5 diff of two JSON reports passed inline |
| `generate_fixes` | 1.6 artifacts as text, never written to disk by MCP |
| `explain_check` | registry entry: severity, tier, fix, doc URL |
| `render_diff` | 1.4, only when started with `-allow-render` |

- Return compact JSON by default (agents pay per token): findings, not 160 PASS rows. `verbose: true` restores full detail.
- Add MCP `resources` for the latest report per host in the session.
- Keep `internal/mcpserver` stdlib-only JSON-RPC; add `tools/list` schema tests and one end-to-end stdio test per new tool with `httptest`.

### Phase 1 exit criteria

- `findable crawl https://<real docs site> -limit 200` finishes under 2 minutes with concurrency 4 and lists duplicate titles, broken links and redirect chains that Screaming Frog also finds on the same site (spot-check 2 sites, record in `docs/roadmap/validation.md`).
- `findable bots` correctly flags a Cloudflare "Block AI bots" site (find one public example; record it).
- A GitHub Action run on this repo's own site gates on `-fail-on-new` and posts a Markdown summary.
- README updated: subcommands, exit codes, evidence tiers, the llms.txt stance, G14 fixed.

## 6. Phase 2: depth (after Phase 1 ships)

### 2.1 Measured performance and page weight (fills the empty CWV category, G10)

- `net/http/httptrace` per fetch: DNS, connect, TLS, TTFB, download. Report p50/p95 across a crawl.
- `-weight`: fetch subresources referenced in raw HTML (CSS, JS, images, fonts) through the guarded client with HEAD then ranged GET, sum transfer bytes, flag `page_weight` (> 2.5 MB), `js_weight` (> 500 KB), `uncompressed_text` (no gzip/br on text), `no_cache_headers`, `render_blocking_css_js` in head, `image_oversize` (bytes vs declared dimensions).
- Lab LCP/CLS/INP need the Chrome DevTools Protocol; `--dump-dom` cannot read them. Decision point: add `chromedp` (one dep) or point to Lighthouse/CrUX for CWV. Recommendation: point to Lighthouse/CrUX and keep findable on bytes and server timing, which crawlers actually feel.
- Files: `internal/fetch/trace.go`, `internal/audit/checks_weight.go`. Tests with `httptest` serving known byte sizes and missing `Content-Encoding`.

### 2.2 Schema.org validation against Google rich result requirements

- Today `checks_seo_schema.go` covers Organization, WebSite, Breadcrumb, FAQ, Article, Product, LocalBusiness, Video. Add required/recommended property tables per type from Google's structured data docs (Product offers/review/aggregateRating, Recipe, Event, HowTo deprecation note, JobPosting, SoftwareApplication, Dataset, ProfilePage, DiscussionForumPosting) as a data file `internal/audit/schema_rules.json` so updates are data, not code.
- Cross-checks against visible content: `schema_matches_visible` (Product price/name, Article headline vs h1, FAQ answers present in visible text). Google requires markup to match visible content; a chat LLM given one page could do this, but not across 500 pages.
- `@id` graph resolution across blocks and pages (Organization `@id` reused site-wide).

### 2.3 Entity and E-E-A-T signals, site-wide

- Site-level: about, contact, privacy, terms, author pages exist and are linked from the footer of most pages; Organization identity consistent across pages (same name, logo, `sameAs` set); every article has an author whose page exists and carries `Person` with `sameAs`.
- `sameAs` link liveness (HEAD each, guarded) and known-authority hosts (Wikipedia, Wikidata, LinkedIn, Crunchbase, GitHub).
- Tier: reported/heuristic. Files: `internal/audit/sitechecks_entity.go`.

### 2.4 Citation readiness per passage

- Split main content into passages by heading. For each: word count, whether it opens with a direct answer sentence, contains a number/unit, a named source or link, a date; whether its heading is a question.
- Output top 5 "most quotable" and bottom 5 passages per page with selectors, so the fix is "rewrite the passage under `h2#pricing`", not "improve citability". Deterministic heuristics only; label tier `heuristic`.
- Keep Google's "no chunking needed" guidance visible in the report next to this section.

### 2.5 Bing and IndexNow

- Detect IndexNow key file (`/<key>.txt` referenced by meta or common locations is not discoverable; instead offer `findable indexnow submit -key K -urls changed.txt` that posts to `api.indexnow.org` after a deploy). Opt-in, explicit command, never during audit.
- `bingbot` already in 1.2.

### 2.6 Watch mode and history

- `findable audit|crawl ... -history .findable/history/` appends dated JSON reports; `findable trend .findable/history/` prints score lines and the checks that flipped. Plain files, no database. Feeds the hosted Pro "watch" plan in `docs/PRO.md`.

## 7. Phase 3: optional answer-visibility probes (never gates CI)

- `findable cite -prompts prompts.txt -engine perplexity|openai -base-url ...`: asks each prompt N times through an OpenAI-compatible endpoint the user configures (key from env), records which URLs from the audited host are cited and how often. Reports frequency with run count; never a single yes/no.
- Explicitly nondeterministic: excluded from scores and gates, report states model, date, run count.
- Only build if Phase 1 and 2 users ask for it; the Otterly/Profound lane is crowded and SaaS-shaped.

## 8. Deliberately not doing

- LLM-judged content scoring inside the audit. Not reproducible; an agent calling the MCP server can do this itself with the evidence we return.
- Keyword research, rank tracking, backlink data. Needs third-party data; not our moat.
- Scoring llms.txt as a ranking factor. It stays a low-severity heuristic check with an honest label.
- A plugin system. Revisit after a third real request.

## 9. Sequencing and size

| Step | Depends on | Rough size |
|---|---|---|
| 0.1 subcommands | none | 1 day |
| 0.2 registry, severity, evidence, weighted score | 0.1 | 3 days (writing 160 fix sentences is most of it) |
| 0.3 config | 0.1 | 0.5 day |
| 1.1 fix list | 0.2 | 1 day |
| 1.2 bots | 0.2 | 2 days |
| 1.3 crawl graph | 0.2 | 4 days |
| 1.5 baseline/gates | 0.2, 0.3 | 2 days |
| 1.7 md/html/sarif | 1.1 | 2 days |
| 1.6 fix generators | 1.3 | 3 days |
| 1.4 render diff | 0.2 | 2 days |
| 1.8 MCP | 1.1-1.6 | 1.5 days |

Ship order: 0.1, 0.2, 1.1, 1.2 (first release: "the fix list + bot probe"), then 1.3 + 1.5 + 1.7 (second release: "crawl and gate"), then 1.6, 1.4, 1.8.

Each step: failing test first at the public seam (CLI output, report bytes, or MCP response), minimal implementation, `make lint test`, CHANGELOG entry, separate reviewer pass before merge.
