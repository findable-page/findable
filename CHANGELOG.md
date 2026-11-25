# Changelog

All notable changes to this project are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Changed (breaking)
- CLI audits now default to severity/evidence-weighted scoring; use `-score-model flat` to retain legacy scores and CI thresholds. The GitHub Action defaults its new `score-model` input to `flat` for existing gates.
- Audit findings now carry severity, evidence tier, fix and evidence fields in reports. `bots` probes crawler access; config loading and runtime errors can exit with code 2 and 3 respectively.

### Added
- `findable fanout -query TOPIC` builds a deterministic sub-intent cluster and reports which intents the page covers in a terminal table or JSON.
- `findable hydrate` compares raw server HTML with a rendered DOM (`-rendered` file or `-render-cmd` program). It reports `seo_ssr_title_match`, `seo_ssr_canonical_match`, `seo_ssr_headings_present`, `aeo_ssr_schema_present` and `geo_ssr_content_parity` when signals exist only after client rendering. These checks are not part of the scored audit.
- Scans report an `access` gate. `noindex`, a robots `none` directive, `nosnippet`, a non-2xx response, a robots.txt block on this page, or a disallowed AI search crawler sets `access.status` to `BLOCKED_FROM_INDEXING` and `access.accessBlocked` to true. The CLI prints that banner before the scores, and the web report shows it above the score. Training-crawler blocks do not close the gate. Content scores are unchanged.
- `findable simulate` posts a query to a configured answer-engine URL and reports brand mention, sentiment, whether the target domain was cited, and the other cited domains. The engine URL comes from `-engine` or `FINDABLE_SIMULATE_ENGINE`; `FINDABLE_SIMULATE_TOKEN` is sent as a bearer token.
- Two E-E-A-T schema checks: `geo_schema_knows_about` (Organization/WebSite JSON-LD should declare a `knowsAbout` array with 3 or more core topics; industry GEO studies report a citation lift from explicit topical enumeration) and `geo_author_credential_depth` (author Person JSON-LD should carry a verifiable credential: `jobTitle`, `alumniOf`, `hasCredential`, or an academic/professional profile such as ORCID or LinkedIn in `sameAs`). Both are reported as not applicable when the page declares no Organization/WebSite or author Person schema.
- Two RAG chunk integrity checks: `aeo_chunk_token_budget` (h2-h4 sections over ~500 tokens, about 375 words, without a subheading get split mid-passage by AI retrieval chunkers; token counts use the words x 1.33 approximation) and `aeo_heading_semantic_specificity` (generic h2/h3 labels like Overview, Details, mehr Infos or Section 2 weaken metadata-assisted vector retrieval).
- `findable diff [-fail-on-regression] [-max-drop N] [-markdown] <baseline.json> <current.json>`: compare any JSON report (page, crawl or sitemap) against a stored baseline, showing per-page score movement, checks that stopped passing and improvements; `-fail-on-regression` exits 1 for CI score-regression gating.
- Site check `llms-txt-spec`: when `/llms.txt` is present, its Markdown structure is validated against the llmstxt.org v2 format (single H1 title first, blockquote summary immediately after it, H2 sections whose lists are `- [Title](url)` links with absolute http(s) URLs); failures report the offending line. Not applicable when the file is missing. Link reachability checks are out of scope.
- Three answer-extractability checks: `aeo_lead_direct_answer` (a 20-80 word direct-answer paragraph within the first 100 words after the H1), `geo_snippet_boilerplate_control` (long nav/header/footer text should sit under `data-nosnippet` div/span/section wrappers, the only elements Google honours), and `meta_robots_snippet_directives` (flags explicitly restrictive `max-snippet`, `max-image-preview` or `max-video-preview` robots meta values; omitted directives keep Google's defaults and pass).
- New `geo_entity_authoritative_sameas` check: schema `sameAs` links must point to authoritative entity registries (wikidata.org, wikipedia.org, crunchbase.com, subdomains included), not only the site's own pages or social profiles; failing links are listed as evidence. `geo_knowledge_graph` keeps checking plain `sameAs` presence.
- `geo_schema_id_graph_coherence` fails when an `@id` reference (author, publisher, isPartOf and the other entity links) does not resolve to a node that declares that `@id`. `geo_brand_entity_consistency` compares `og:site_name`, `WebSite.name`, `Organization.name` and a title site suffix split on `|` or `-`.
- Princeton GEO evidentiary density checks: `geo_statistics_density` now measures density (at least one quantitative data point per 300 words) instead of a flat count of 3, and new `geo_assertion_evidence_pairing` flags content sections (h2/h3 blocks over 100 words) that carry no figure, date, or citation/external link. Sample sizes (`n=123`) now count as quantitative evidence.
- Site checks now distinguish AI search retrieval crawlers (`OAI-SearchBot`, `PerplexityBot`, `Claude-SearchBot`, `Googlebot`) from AI model training crawlers (`GPTBot`, `ClaudeBot`, `Google-Extended`, `Applebot-Extended`, `meta-externalagent`), with labels explaining that blocking training bots does not affect AI search citations and that `Google-Extended` governs Gemini training, not AI Overviews.

### Fixed
- `geo_intent_clarity` no longer fails a title that is the H1. A one-word title such as "React", and a short subject beside a site suffix such as "URL - Wikipedia", match the H1. "AI" still does not match "Email", and a multi-word title still needs two shared words longer than 3 letters. Pages that were false failures gain this low-severity check, so their GEO and overall scores can rise.
- Percentages like `9%.` never matched the quantitative-statistics pattern because a word boundary after `%` can never hold; percentages are now a dedicated alternative, so they count toward `geo_statistics_density` and section evidence.
- `link_noopener` now matches whole `rel` tokens; values such as `notnoopener` no longer hide unsafe `_blank` links.
- Signal extraction no longer runs HTML element collectors on SVG/MathML elements: an inline SVG icon's `<title>` (an accessibility label for the graphic) no longer overwrites the page `<title>` used by `title_exists`, `title_length`, `keyphrase_in_title` and `geo_intent_clarity`.
- Robots meta and X-Robots-Tag `none` directive (documented by Google as `noindex, nofollow`) is now treated as noindex by `robots_meta`, `aeo_ai_robots`, `x_robots_tag_noindex`, the `bots` probe verdicts and `meta_robots_conflict`, instead of silently passing.
- `title_length`, `desc_length` and `img_alt_length` now count characters instead of bytes, so titles, meta descriptions and alt text containing non-ASCII text (CJK, Arabic, accented Latin, emoji) are no longer penalized as too long.
- `geo_intent_clarity` now compares the title with the primary `<h1>` instead of whichever heading appears first on the page, so pages whose navigation renders an `<h2>` before the main `<h1>` are no longer falsely flagged. When a page has no title or no `<h1>`, the check is reported as not applicable instead of failing.
- The `charset` check now honours the three HTML-spec declaration sources — `<meta charset>`, `<meta http-equiv="content-type">` and the `Content-Type` response header — instead of searching the raw HTML for the string `utf-8`, which false-passed on pages that only mention it in scripts or copy and false-failed pages declaring it via the HTTP header.

## [0.3.0] - 2026-10-15

### Added
- Local audit engine in Go: 72 page checks (36 SEO, 17 AEO, 19 GEO) and 7 opt-in site checks.
- Output formats `-format csv|json|sarif`, CI gate `-fail-under N`, `-sitemap` and `-site`.
- Composite GitHub Action that uploads SARIF to code scanning and gates on score.
- MCP server (`-mcp`) and `findable-mcp` npm runner.
- Cloudflare Worker edge API and web UI served from one source, with native rate limiting.
- `/learn` articles on SEO, AEO, GEO and every check group.
- Logo drafts, build story, cover image, CONTRIBUTING, SECURITY with an SSRF policy, issue and PR templates.
- Strict golangci-lint and Biome configs, CI jobs and GoReleaser config.
- Worker site files: `/robots.txt`, `/sitemap.xml`, `/llms.txt`, `/llms-full.txt`, JSON-LD in the home page head, and HSTS, X-Frame-Options and Referrer-Policy headers.
- Worker static assets in `worker/public/` (favicon, touch icon, og image), with a test that keeps them in step with the UI.
- A test that scans the home page in-process and requires SEO 90, AEO 85 and GEO 80.
- Bare hosts like `example.com` are accepted and scanned over https.
- Share button, Markdown export and a WebMCP `scan_url` tool in the web UI.
- The Go engine runs the four focus keyphrase checks.

### Changed
- Renamed the product to findable; module path is `github.com/nadimtuhin/findable`.
- Split the CLI into `cmd/findable` and `internal/` packages.
- Results UI shows a score hero, family cards and a failing-first issue list with AA contrast colors.
- The API returns the composite overall score.
- Redesigned the web UI for the AI era; the worker bundle budget is now 50 KiB gzip.
- JSON-LD parsing walks arrays, `@type` lists and nested nodes in both engines; only page-level or `author` Persons count as author schema.
- A heading counts as a question when it starts with what, how or why, or ends with `?`, in both engines.

### Removed
- The seoscore.tools remote mode; findable runs standalone.
- Committed `worker/dist`, agent scratch notes and design mockups (now in nadimtuhin/design-fusion).

### Security
- Scan fetches block private and reserved addresses (including 6to4, Teredo and NAT64), allow only ports 80 and 443, re-check up to 5 redirects, and cap pages at 5 MB.
- The worker UI is served with a hash-based CSP that refuses framing, plus nosniff.
- Scans are rate limited to 10 a minute per client, IPv6 clients by /64.
- `-serve` refuses local files and private targets.
