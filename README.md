# seoscore-cli

[![CI](https://github.com/nadimtuhin/seoscore-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/nadimtuhin/seoscore-cli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nadimtuhin/seoscore-cli)](https://goreportcard.com/report/github.com/nadimtuhin/seoscore-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Quick start

```bash
go install github.com/nadimtuhin/seoscore-cli/cmd/seoscore@latest
seoscore https://example.com
```

Or try it in the browser: **https://seoscore.piply.xyz**

A fast, unified open-source website auditor for **Traditional SEO**, **Answer Engine Optimization (AEO)**, and **Generative Engine Optimization (GEO)**.

Audits live URLs or local HTML files completely offline in Go, exports identical CSV reports matching [seoscore.tools](https://seoscore.tools), runs a local API daemon with modern Web UI, serves Model Context Protocol (MCP) tools for AI agents, and includes a deployable Cloudflare Worker backend.

## Features

- **100% Offline / Local Engine:** Evaluates 80+ checks in Go without rate limits or external dependencies.
- **Tri-Family Scoring:** Traditional SEO (50%), AEO (25%), and GEO (25%) composite score.
- **Model Context Protocol (MCP):** Native `-mcp` mode for Claude Desktop, Cursor, Hermes Agent, and Windsurf.
- **Interactive Web UI:** Dark/Light theme, circular SVG score rings, interactive fix recommendations, and direct share links.
- **Crawl Site (Beta):** Crawls internal links and provides page-by-page score tables.
- **Dual Deployment:** Local API daemon (`-serve :8080`) or edge deployment on Cloudflare Workers (`worker/`).

## Install

```bash
go install github.com/nadimtuhin/seoscore-cli/cmd/seoscore@latest
```

## CLI Usage

### 1. Single Page Audit

Audit a live URL:
```bash
seoscore https://example.com
```

Audit a local HTML file directly without any web server:
```bash
seoscore ./path/to/page.html
```

With focus keyphrase:
```bash
seoscore -k "cloud performance" https://example.com
```

### 2. Multi-Page Site Crawl

```bash
seoscore -crawl -limit 20 https://example.com
```

### Output formats

`-format csv` (default), `-format json` (the same object `/api/scan` returns) or `-format sarif`
(SARIF 2.1.0 for GitHub code scanning). `/api/scan` takes the same choice as `"format"` in the body.

### CI gate

`-fail-under 70` exits 1 and names each page whose overall score is below 70 (any crawled page counts).

### 3. Run Local REST API & Web UI

```bash
seoscore -serve :8080
```
Open `http://localhost:8080` in your browser for the full web dashboard.

---

## Model Context Protocol (MCP) Tools

Use `seoscore` directly inside Claude Desktop, Cursor, Windsurf, or Hermes Agent to audit sites and optimize content for AI search engines (Google AI Overviews, Perplexity, ChatGPT).

### 1. Native Go CLI MCP Server (Zero Dependencies)

Add to your `claude_desktop_config.json` or Cursor `settings.json`:

```json
{
  "mcpServers": {
    "seoscore": {
      "command": "seoscore",
      "args": ["-mcp"]
    }
  }
}
```

### 2. Node / NPX MCP Runner

```json
{
  "mcpServers": {
    "seoscore": {
      "command": "npx",
      "args": ["-y", "seoscore-mcp"]
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

- **Live Endpoint:** `https://seoscore.piply.xyz`
- **Audit Endpoint:** `POST https://seoscore.piply.xyz/api/scan`

```bash
curl -s -X POST https://seoscore.piply.xyz/api/scan \
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
