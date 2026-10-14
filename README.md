<p align="center"><img src="docs/cover.png" alt="findable: is your site findable by Google and by ChatGPT and Perplexity?" width="100%"></p>

# Is your site findable by Google AND by ChatGPT/Perplexity?

**findable: SEO + AEO + GEO auditor for the AI search era (CLI, MCP, edge API).** Scores summarise findable's own checks; they do not measure Google rankings or AI citations.

[![CI](https://github.com/nadimtuhin/findable/actions/workflows/ci.yml/badge.svg)](https://github.com/nadimtuhin/findable/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nadimtuhin/findable)](https://goreportcard.com/report/github.com/nadimtuhin/findable)
[![Release](https://img.shields.io/github/v/release/nadimtuhin/findable?include_prereleases)](https://github.com/nadimtuhin/findable/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Live demo: https://findable.piply.xyz** · **Learn: https://findable.piply.xyz/learn** · [The build story](docs/STORY.md)

![findable web UI showing a score hero, family cards and a failing-first issue list](docs/screenshots/after-light-1280.png)

## Why

AI answer engines now sit between your page and your reader. They tend to quote short, structured, attributed passages. A page can rank well on Google and still not be cited in a ChatGPT or Perplexity answer.

findable checks the signals answer engines are documented or commonly reported to use: whether a machine can lift a clean answer out of the page (AEO), and whether the page carries trust signals a generative engine may use to cite it (GEO), alongside classic SEO. It is written in Go with no account and no quota. Scoring a local HTML file is fully offline; URL, crawl and site checks make network requests, and the hosted API exists if you prefer not to run it yourself.

## Scoring model

One overall score from 0 to 100, built from three families:

| Family | Weight | Page checks | What it measures |
|---|---|---|---|
| SEO | 50% | 36 | Title and meta tags, headings, canonical, Open Graph and Twitter cards, mobile, HTTPS and security headers, accessibility landmarks |
| AEO (answer engine optimization) | 25% | 17 | Structured data, FAQ markup, short direct answers, question headings, lists, tables, steps, readability |
| GEO (generative engine optimization) | 25% | 19 | Author and organization markup, dates, citations and quotes, passage-ready blocks, trust links, no snippet blocking |

That is 72 page checks in the Go engine. The edge API runs a smaller port with 54. Seven more site checks are opt-in (`-site`) and unscored: robots.txt present, the page allowed for `*`, GPTBot, ClaudeBot, PerplexityBot and Google-Extended, and `/llms.txt` present. Every failing check names what to fix.

## Quickstart

```bash
go install github.com/nadimtuhin/findable/cmd/findable@latest   # Go 1.26+
npx findable-mcp                                                # MCP server for agents
```

Homebrew: planned for first release; use go install.

Then:

```bash
findable https://example.com
findable ./page.html          # local file, no server needed
```

It prints the four scores and the failing checks, and saves a report file in `-o` (default `.`).

## CLI

```
findable [-o dir] [-format csv|json|sarif] [-fail-under N] [-q] [-k keyphrase]
         [-site] [-crawl] [-sitemap url] [-limit N] [-serve :port] [-mcp] [-allow-private] <url or file>
```

| Flag | Meaning |
|---|---|
| `-format csv\|json\|sarif` | Report format. `json` is the same object `/api/scan` returns, `sarif` is SARIF 2.1.0 for GitHub code scanning. Default `csv`. |
| `-fail-under N` | Exit 1 and name every page whose overall score is below N. |
| `-sitemap URL` | Audit the pages a sitemap lists. Sitemap indexes are followed. |
| `-site` | Also run the 7 site checks (robots.txt, AI crawlers, llms.txt). |
| `-crawl` | Follow internal links, up to `-limit` pages (default 20). |
| `-k` | Focus keyphrase. Only the edge API scores it today; the offline engine ignores it. |
| `-o`, `-q` | Report directory; `-q` hides the failing-check list. |
| `-serve :8080` | Local API and web UI. Keep it on localhost. |
| `-mcp` | MCP server over stdio. |
| `-allow-private` | Allow localhost and private-network targets. Without it every fetch, including crawl and sitemap links and MCP, refuses non-public hosts. |

```bash
findable -q -format sarif -fail-under 70 https://example.com
findable -sitemap https://example.com/sitemap.xml -limit 50
findable -site -crawl -limit 10 https://example.com
```

## MCP for agents

Add to Claude Desktop, Cursor or Windsurf:

```json
{ "mcpServers": { "findable": { "command": "findable", "args": ["-mcp"] } } }
```

No Go toolchain? Use `{ "command": "npx", "args": ["-y", "findable-mcp"] }` instead.

Tools: `audit_page` (URL or file), `audit_html_content` (raw HTML, no network), `crawl_site`, `generate_csv_report`. Prompt: `seo_audit_prompt`.

## API

Hosted at https://findable.piply.xyz, or run your own with `findable -serve :8080` or `cd worker && npx wrangler deploy`.

```bash
curl -s -X POST https://findable.piply.xyz/api/scan \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","keyphrase":"example","site":true}'
```

`/api/scan` only fetches public hosts: ports 80 and 443, 5 redirects, 10 seconds, 5 MB, rate limited. See [SECURITY.md](SECURITY.md).

## GitHub Action

```yaml
permissions:
  security-events: write
steps:
  - uses: nadimtuhin/findable@v0.2.0
    with:
      url: https://example.com
      fail-under: 70
      args: -site
```

The action builds the CLI, uploads failing checks to code scanning as SARIF, then fails the job below `fail-under`.

## Development

```bash
git clone https://github.com/nadimtuhin/findable && cd findable
(cd mcp && bun install) && (cd worker && bun install)
make lint test
```

See [CONTRIBUTING.md](CONTRIBUTING.md) and [CHANGELOG.md](CHANGELOG.md). `testdata/` holds sample pages and a scoring benchmark script.

## License

MIT
