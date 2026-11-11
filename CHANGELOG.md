# Changelog

All notable changes to this project are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Changed (breaking)
- CLI audits now default to severity/evidence-weighted scoring; use `-score-model flat` to retain legacy scores and CI thresholds. The GitHub Action defaults its new `score-model` input to `flat` for existing gates.
- Audit findings now carry severity, evidence tier, fix and evidence fields in reports. `bots` probes crawler access; config loading and runtime errors can exit with code 2 and 3 respectively.

### Added
- Two schema E-E-A-T checks from the topical-depth research: `geo_schema_knows_about` (Organization/WebSite JSON-LD should declare a `knowsAbout` array with 3 or more topics; strings or Thing objects count) and `geo_author_credential_depth` (article author Person JSON-LD should carry a `jobTitle`, `alumniOf`, `hasCredential` or a professional profile such as LinkedIn in `sameAs`). Both are reported as not applicable when the page has no such schema (or is not an article).
- Three answer-extractability checks: `aeo_lead_direct_answer` (a 20-80 word direct-answer paragraph within the first 100 words after the H1), `geo_snippet_boilerplate_control` (long nav/header/footer text should sit under `data-nosnippet` div/span/section wrappers, the only elements Google honours), and `meta_robots_snippet_directives` (flags explicitly restrictive `max-snippet`, `max-image-preview` or `max-video-preview` robots meta values; omitted directives keep Google's defaults and pass).
### Fixed
- `geo_intent_clarity` now compares the title with the primary `<h1>` instead of whichever heading appears first on the page, so pages whose navigation renders an `<h2>` before the main `<h1>` are no longer falsely flagged. When a page has no title or no `<h1>`, the check is reported as not applicable instead of failing.

### Added
- Site checks now distinguish AI search retrieval crawlers (`OAI-SearchBot`, `PerplexityBot`, `Claude-SearchBot`, `Googlebot`) from AI model training crawlers (`GPTBot`, `ClaudeBot`, `Google-Extended`, `Applebot-Extended`, `meta-externalagent`), with labels explaining that blocking training bots does not affect AI search citations and that `Google-Extended` governs Gemini training, not AI Overviews.

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
