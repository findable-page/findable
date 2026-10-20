---
title: "Get Findable in 15 Minutes: A Step by Step Checklist"
description: "Install findable, audit a page and your site, fix the top failures, then add MCP, the web UI, the API and a CI gate. Every step has a copy-paste command."
slug: findable-in-15-minutes
date: 2026-10-14
updated: 2026-10-14
author: "findable team"
faq:
  - q: "Does findable need an API key?"
    a: "No. The CLI runs locally and audits any URL or local HTML file. You only need Go to install it."
  - q: "What is the binary called?"
    a: "The binary is named findable. Use findable in every command."
  - q: "How do I fail a build on a low score?"
    a: "Pass -fail-under N. The CLI exits with code 1 when the score is below N, which stops your CI job."
  - q: "Can I run it on a local file?"
    a: "Yes. Run findable ./page.html to audit a file on disk with no server."
---

# Get Findable in 15 Minutes: A Step by Step Checklist

You can go from nothing to a working audit, an editor integration and a CI gate in about 15 minutes. Follow the steps in order. Each one is a command you can paste.

## How do you install findable?

You install it with one Go command. It puts a binary called `findable` on your path. You need a recent Go toolchain, and this takes about one minute.

```bash
go install github.com/findable-page/findable/cmd/findable@latest
```

## How do you run your first audit?

Point `findable` at a URL and pass your target phrase with `-k`. It fetches the page, runs every check and prints a score. You can also audit an HTML file on disk before you deploy it. Expect this to take about one minute.

```bash
findable -k "your keyphrase" https://example.com
findable ./page.html
```

## How do you check robots.txt and llms.txt?

Add the `-site` flag. It also fetches `/robots.txt` and `/llms.txt` and reports whether GPTBot, ClaudeBot, PerplexityBot and Google-Extended are allowed. These site results are unscored, so they never change your grade. The flag needs a real URL and refuses a local file.

```bash
findable -k "your keyphrase" -site https://example.com
```

## How do you read the output?

The score is SEO 50%, AEO 25% and GEO 25%. Each line is a check with a stable id, such as `title_length` or `aeo_faq`. Failures matter most. Read the failing ids first, then the site section if you used `-site`.

For a machine readable report, choose a format. You can write `csv`, `json` or `sarif`, send files to a folder with `-o`, and quiet the console with `-q`. These are handy when you want to diff two runs, or feed the findings into another tool.

```bash
findable -format json -o ./reports -q https://example.com
```

## How do you fix the top failures?

Take the first three failing ids and fix them. Quick wins are usually `title_length` (30 to 60 characters), `desc_length` (120 to 160), `h1_exists` (exactly one h1) and `canonical`. Then re-run the same command and watch the score rise. Repeat until the top fails are gone.

```bash
findable -k "your keyphrase" -site https://example.com
```

To audit more than one page, crawl. Use `-crawl` with `-limit`, or feed a sitemap with `-sitemap`.

```bash
findable -crawl -limit 25 https://example.com
findable -sitemap https://example.com/sitemap.xml
```

## How do you set up MCP?

Run findable as an MCP server so your editor agent can audit pages for you. Add this to your MCP config. To run it without Go, build the server from source: `cd mcp && bun install && bun run build`, then point `"command": "node"` with `"args": ["/path/to/findable/mcp/dist/index.js"]` at it. The tools are `audit_page`, `audit_html_content`, `crawl_site` and `generate_csv_report`.

```json
{"mcpServers":{"findable":{"command":"findable","args":["-mcp"]}}}
```

The server also offers a prompt called `seo_audit_prompt`. It speaks over stdio, so there is no port to open and nothing is exposed to your network. Ask your agent to audit a page, and it calls these tools for you.

## How do you use the web UI?

Open [findable.page](https://findable.page) and paste a URL to get a scored report in your browser, with no install at all. If you want it local, serve it yourself. The same binary hosts the web UI and the API on a port you choose.

```bash
findable -serve 127.0.0.1:8080
```

## How do you call the API with curl?

Send a POST to `/api/scan` with a JSON body. Set `site` to true for the robots.txt and llms.txt checks, and set `format` to `json` so you can parse the response in a script or a dashboard. The response uses the same check ids as the CLI.

```bash
curl -X POST https://findable.page/api/scan \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","keyphrase":"your keyphrase","site":true,"format":"json"}'
```

## How do you gate CI with the GitHub Action?

Add the action to a workflow. Set `fail-under` to the lowest score you accept. The job fails below that number, and the action uploads SARIF so each failing check shows up as a code scanning alert in the repository. Set `upload-sarif` to false if you only want the exit code.

```yaml
name: findable
on: [push]
jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: findable-page/findable@v0.2.0
        with:
          url: https://example.com
          fail-under: 80
          args: -site
```

You can get the same gate locally with `findable -fail-under 80 https://example.com`. It exits 1 below 80. Start with a threshold near your current score, then raise it as you fix more checks. The Site checks are unscored, so `-site` in `args` adds rows to the report but never changes whether the gate passes.

## Key takeaways

- Install with one `go install` line. The binary is `findable`.
- Run `findable -k "phrase" -site <url>` and read the failing check ids first.
- Fix the top three failures, then re-run to confirm.
- Use `-mcp` for editors, `-serve 127.0.0.1:8080` for a local UI, and POST `/api/scan` for scripts.
- Gate CI with `-fail-under N` or the `findable-page/findable@v0.2.0` action.
