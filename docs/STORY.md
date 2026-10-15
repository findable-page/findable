# The findable story

People used to ask one question: does my site rank on Google? Now they also ask ChatGPT, Perplexity and Claude, and those tools answer in a paragraph instead of a list of links. If your page is not in that paragraph, it may not get seen. I wanted one tool that tells me if a page is ready for both worlds.

## An engine that runs offline

findable started as a small Go CLI that scored a page against a fixed set of checks. The first version leaned on a hosted service. I removed that. Today the engine runs on your machine: it reads a URL or a local HTML file, extracts signals once, and runs 160 page checks plus 7 optional site checks. No account, no rate limit. Scoring a local file is fully offline; fetching a URL or running site checks makes network requests. The CLI, the local API and the MCP server share that engine. The edge API at findable.page runs a TypeScript port of it: the same 160 page checks. Keyphrase scoring is the exception: only the edge API does it today.

## Three families, one score

Classic SEO is half the score: titles, meta tags, headings, mobile, social cards and security headers. Answer engine optimization (AEO) is a quarter: does the page give short, direct answers, use question headings and carry FAQ markup that a machine can lift out. Generative engine optimization (GEO) is the last quarter: does the page show who wrote it, carry dates, organization markup and about and contact links, and avoid blocking snippets. Site checks add robots.txt for AI crawlers and llms.txt. Every failing check says what to change.

## Built for agents

You can run findable as an MCP server. Add one block to Claude Desktop or Cursor, and your agent can audit a page, read the failing checks and edit your content in the same loop. The tools take a URL, a file path or raw HTML.

## An edge API that is safe to expose

The Cloudflare Worker serves a web UI and a JSON endpoint. Letting strangers pick a URL to fetch is a classic SSRF risk, so the guard is strict. It refuses private and reserved addresses, including tricks like 6to4, Teredo and NAT64. It only allows ports 80 and 443, re-checks every redirect, stops after 5 redirects, times out at 10 seconds, caps the page size at 5 MB and rate limits each client.

## A gate for CI

The CLI can fail a build. Pass `-fail-under 70` and it exits 1 when any page scores lower. The GitHub Action wraps this, writes the failing checks as SARIF and uploads them to code scanning, so problems show up as annotations in the pull request, next to your other security findings.

## What is next

More checks as the AI engines change, and better fix guidance. The scoring weights live in one place in the code, so they are easy to argue with. Open an issue if you disagree with one.
