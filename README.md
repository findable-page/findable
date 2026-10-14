# findable

**findable: is your site findable by Google and by AI answer engines?**

[![CI](https://github.com/nadimtuhin/findable/actions/workflows/ci.yml/badge.svg)](https://github.com/nadimtuhin/findable/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nadimtuhin/findable)](https://goreportcard.com/report/github.com/nadimtuhin/findable)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Quick start

```bash
go install github.com/nadimtuhin/findable/cmd/findable@latest
findable https://example.com
```

Or try it in the browser: **https://findable.piply.xyz**

A fast, unified open-source website auditor for **Traditional SEO**, **Answer Engine Optimization (AEO)**, and **Generative Engine Optimization (GEO)**.

Audits live URLs or local HTML files completely offline in Go, exports identical CSV reports matching [seoscore.tools](https://seoscore.tools), runs a local API daemon with modern Web UI, serves Model Context Protocol (MCP) tools for AI agents, and includes a deployable Cloudflare Worker backend.

## Features

- **100% Offline / Local Engine:** Evaluates 72 checks in Go without rate limits or external dependencies.
- **Tri-Family Scoring:** Traditional SEO (50%), AEO (25%), and GEO (25%) composite score.
- **Model Context Protocol (MCP):** Native `-mcp` mode for Claude Desktop, Cursor, Hermes Agent, and Windsurf.
- **Interactive Web UI:** Dark/Light theme, circular SVG score rings, interactive fix recommendations, and direct share links.
- **Crawl Site (Beta):** Crawls internal links and provides page-by-page score tables.
- **Dual Deployment:** Local API daemon (`-serve :8080`) or edge deployment on Cloudflare Workers (`worker/`).

## Install

```bash
go install github.com/nadimtuhin/findable/cmd/findable@latest
```

## CLI Usage

### 1. Single Page Audit

Audit a live URL:
```bash
findable https://example.com
```

Audit a local HTML file directly without any web server:
```bash
findable ./path/to/page.html
```

With focus keyphrase:
```bash
findable -k "cloud performance" https://example.com
```

### 2. Multi-Page Site Crawl

```bash
findable -crawl -limit 20 https://example.com
```

Or audit the pages a sitemap lists (sitemap indexes are followed, `-limit` still applies):

```bash
findable -sitemap https://example.com/sitemap.xml -limit 50
```

### Site checks

`-site` also fetches `/robots.txt` and `/llms.txt` from the page's origin and reports a separate, unscored
"Site" section: robots.txt present, the page not blocked for `*`, GPTBot, ClaudeBot, PerplexityBot and
Google-Extended, and llms.txt present. The page score formula is unchanged. The API (`"site": true`),
the MCP `audit_page` tool (`site`) and the worker accept the same switch.

### Output formats

`-format csv` (default), `-format json` (the same object `/api/scan` returns) or `-format sarif`
(SARIF 2.1.0 for GitHub code scanning). `/api/scan` takes the same choice as `"format"` in the body.

### CI gate

`-fail-under 70` exits 1 and names each page whose overall score is below 70 (any crawled page counts).

### GitHub Action

```yaml
permissions:
  security-events: write
steps:
  - uses: nadimtuhin/findable@main
    with:
      url: https://example.com
      fail-under: 70
      args: -site
```

The action builds the CLI, uploads failing checks to code scanning as SARIF, then fails the job below `fail-under`.

### 3. Run Local REST API & Web UI

```bash
findable -serve :8080
```
Open `http://localhost:8080` in your browser for the full web dashboard.

---

## Model Context Protocol (MCP) Tools

Use `findable` directly inside Claude Desktop, Cursor, Windsurf, or Hermes Agent to audit sites and optimize content for AI search engines (Google AI Overviews, Perplexity, ChatGPT).

### 1. Native Go CLI MCP Server (Zero Dependencies)

Add to your `claude_desktop_config.json` or Cursor `settings.json`:

```json
{
  "mcpServers": {
    "findable": {
      "command": "findable",
      "args": ["-mcp"]
    }
  }
}
```

### 2. Node / NPX MCP Runner

```json
{
  "mcpServers": {
    "findable": {
      "command": "npx",
      "args": ["-y", "findable-mcp"]
    }
  }
}
```

### Available MCP Tools & Prompts

- `audit_page`: Audits a URL or local file path for SEO, AEO, and GEO signals.
- `audit_html_content`: Audits raw HTML markup in-memory without any network requests.
- `crawl_site`: Crawls up to N internal pages and returns per-page scores and aggregate statistics.
- `generate_csv_report`: Returns a full CSV report string formatted according to seoscore.tools.
- `seo_audit_prompt`: Structured prompt template generating prioritized technical, AEO, and GEO fixes.

---

## Free Cloudflare Edge API

A production-ready edge API is deployable via Wrangler or accessible directly:

- **Live Endpoint:** `https://findable.piply.xyz`
- **Audit Endpoint:** `POST https://findable.piply.xyz/api/scan`

```bash
curl -s -X POST https://findable.piply.xyz/api/scan \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","keyphrase":"example"}'
```

To deploy your own instance:
```bash
cd worker
npx wrangler deploy
```

---

## License

MIT
