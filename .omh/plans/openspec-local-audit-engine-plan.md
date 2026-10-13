---
title: "Implementation Plan: Local Offline Audit Engine"
status: confirmed
date: "2026-10-11"
spec: ".omh/specs/openspec-local-audit-engine.md"
---

# Implementation Plan: Local Offline Audit Engine

## Step 1: Signal Extraction Engine (`signals.go`)
- Create `type PageSignals struct` to capture:
  - Document metadata: Title, Description, Canonical, Lang, Viewport, Charset, Robots, MaxImagePreview.
  - Social metadata: OG tags, Twitter Card tags.
  - Heading outline: List of headings with level, text, and order.
  - Content metrics: Raw text, word count, sentence count, paragraphs, reading metrics.
  - Structured data: List of parsed JSON-LD objects.
  - Structural elements: Counts of `<article>`, `<section>`, `<nav>`, `<main>`, `<header>`, `<footer>`, `<time>`, `<dl>`, `<ol>`, `<ul>`, `<table>`, `<blockquote>`, `<pre>`, `<code>`, `<details>`.
  - Links: List of internal and external URLs, anchor texts.
  - Media: Images with alt texts, widths, heights.
  - HTTP headers: Status code, Content-Type, Content-Security-Policy, HSTS, X-Content-Type-Options, X-Frame-Options, Cache-Control.
- Implement `extractSignals(rawHTML string, headers http.Header, targetURL string) (PageSignals, error)` using `golang.org/x/net/html`.

## Step 2: Rules Engine (`engine.go`)
- Define `type ruleFunc func(s PageSignals) check`
- Implement rule functions:
  - `evalSEO(s PageSignals) category`
  - `evalAEO(s PageSignals) category`
  - `evalGEO(s PageSignals) category`
- Implement `Audit(rawHTML string, headers http.Header, targetURL string) scan`
  - Computes `Passed`, `Total`, and `Score = round(float64(passed) / float64(applicable) * 100)`.

## Step 3: CLI Integration (`main.go`)
- Add `-remote` flag (boolean, default `false`).
- When `-remote` is false (default):
  - Directly fetch the URL using standard Go `http.Get` (with user-agent, follow redirects).
  - Also support file paths (e.g. `file://...` or path to `.html` file) for completely offline file testing!
  - Run in-process `Audit()`.
- When `-remote` is true:
  - Call `https://seoscore.tools/api/scan` as before.
- Support crawl mode `-crawl` in offline mode as well!

## Step 4: Verification & Benchmarking
- Add comprehensive unit tests in `engine_test.go`:
  - Test signal extraction.
  - Test individual check outcomes.
  - Run local engine across all 10 archetypes in `testdata/site/`.
  - Assert that scores reflect expected quality tiers.
