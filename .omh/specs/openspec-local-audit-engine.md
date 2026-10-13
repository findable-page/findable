---
title: "OpenSpec: Unified Local SEO/AEO/GEO Audit Engine"
status: confirmed
version: "2.0.0"
date: "2026-10-11"
authors: ["Nadim Tuhin"]
type: "openspec"
sources:
  - "SiteOne Crawler (Rust, janreges/siteone-crawler)"
  - "Booltools SEO Crawler (Go, booltools/booltools-seo-crawler)"
  - "GEO Optimizer Audit (Python, Auriti-Labs/geo-optimizer-skill)"
  - "seo-aeo-audit (Python, openairlabs/seo-aeo-audit)"
  - "seoscore.tools (Cloudflare Worker, seoscore.tools)"
---

# OpenSpec: Unified Local SEO/AEO/GEO Audit Engine

## 1. Executive Summary

This specification merges the audit rules, heuristics, and scoring models of the leading open-source website graders into a single, unified, dependency-light Go audit engine. It provides 100% offline and in-process execution, eliminating remote API dependencies and rate-limiting bottlenecks while matching the output schema of `seoscore.tools`.

### Synthesis of Upstream Projects (via Perplexity Research):
1. **SiteOne Crawler:** Deterministic category weighting, strict security headers (CSP, HSTS, X-Frame-Options, X-Content-Type), DOM depth limits, ARIA accessibility landmarks, and HTML doctype hygiene.
2. **Booltools SEO Crawler:** 15-category taxonomy spanning technical SEO, on-page SEO, and GEO rules for AI-search readiness.
3. **GEO Optimizer Audit:** Google AI Overviews readiness (`G-INDEX`, `G-SNIPPET`, `G-CANONICAL`, `G-DATES`, `G-BYLINE`), AI crawler permissions (GPTBot, ClaudeBot, PerplexityBot), `llms.txt` discovery, knowledge graph `sameAs` entity verification, and quantitative data density.
4. **seo-aeo-audit:** Answer Engine Optimization heuristics including early extractability (front-loaded answer budget), question-style headings, definition lists (`<dl>`), conversational pronoun density, and active voice dominance.
5. **seoscore.tools:** Tri-part scoring taxonomy: SEO (50%), AEO (25%), GEO (25%) yielding a composite 0–100 overall score and standard CSV export.

---

## 2. Merged Rule Taxonomy & Signal Matrix

The merged engine organizes checks into three primary families:

### Category 1: Traditional SEO (Weight: 50%)

#### A. Meta & Document Signals
- `title_exists`: `<title>` tag exists and is non-empty.
- `title_length`: 30–60 characters (optimal for Google SERP truncation).
- `desc_exists`: `<meta name="description">` exists and is non-empty.
- `desc_length`: 120–160 characters (optimal snippet size).
- `viewport`: `<meta name="viewport">` present with `width=device-width`.
- `charset`: `<meta charset="utf-8">` declared.
- `lang_attr`: `<html lang="...">` defined on root tag.
- `canonical`: `<link rel="canonical">` present with absolute URL.
- `canonical_self`: Canonical matches target URL or provides clean canonicalization.
- `html_doctype`: `<!DOCTYPE html>` correctly declared.
- `meta_refresh`: No client-side `<meta http-equiv="refresh">` redirect traps.
- `robots_meta`: No `noindex` or `nofollow` directives preventing crawling.
- `max_image_preview`: `max-image-preview:large` included for Google Discover.

#### B. Content & Hierarchy
- `h1_exists`: Exactly one `<h1>` tag present on the page.
- `h1_length`: 20–70 characters for optimal topic definition.
- `title_h1_alignment`: Keywords or concept in title aligned with H1.
- `h2_exists`: At least one `<h2>` tag present.
- `heading_hierarchy`: Strict hierarchical order (no skipping levels, e.g., H1 -> H3).
- `word_count`: Main body contains at least 300 words (flags thin content).
- `text_html_ratio`: Text-to-HTML ratio >= 15%.
- `dom_size`: Total DOM elements < 1,500 elements.

#### C. Links & Media
- `internal_links`: Page contains internal navigation links.
- `external_links`: Page cites external authoritative sources.
- `empty_hrefs`: No empty or dangling `href=""` or `href="#"` links.
- `link_noopener`: External `target="_blank"` links include `rel="noopener"`.
- `descriptive_anchors`: Anchor texts avoid generic "click here" or "read more".
- `img_alt`: All `<img>` tags possess descriptive `alt` attributes.
- `img_dimensions`: Images declare explicit `width` and `height` attributes to avoid Cumulative Layout Shift (CLS).

#### D. Social & Branding
- `og_tags`: Complete OpenGraph suite (`og:title`, `og:description`, `og:image`, `og:url`, `og:type`).
- `twitter_card`: Twitter cards declared (`twitter:card`, `twitter:title`, `twitter:description`, `twitter:image`).
- `favicon`: Favicon or SVG icon declared.
- `web_manifest`: Web App Manifest (`manifest.json`) linked.

#### E. Technical Security & Headers (SiteOne Integration)
- `https`: Target serves over TLS/HTTPS.
- `hsts`: `Strict-Transport-Security` header enforced.
- `x_content_type`: `X-Content-Type-Options: nosniff` header present.
- `x_frame_options`: `X-Frame-Options` (DENY/SAMEORIGIN) or CSP frame-ancestors.
- `csp`: `Content-Security-Policy` header declared.
- `referrer_policy`: `Referrer-Policy` declared.
- `permissions_policy`: `Permissions-Policy` restricts sensitive browser capabilities.

#### F. Accessibility & Structure
- `aria_landmarks`: Essential landmarks (`<main>`, `<nav>`, `<header>`, `<footer>`) present.
- `form_labels`: All input fields have associated `<label>` or `aria-label`.
- `duplicate_ids`: No duplicate DOM `id` attributes.

---

### Category 2: AEO — Answer Engine Optimization (Weight: 25%)

#### A. Direct Extraction & Structure
- `aeo_structured_data`: Valid Schema.org JSON-LD found.
- `aeo_faq`: `FAQPage` schema or Q&A section detected.
- `aeo_direct_answers`: Concise 40–80 word direct answer paragraphs placed immediately below H2/H3 headings.
- `aeo_question_headings`: Headings formatted as questions ("What is...", "How to...", "Why does...").
- `aeo_definition_lists`: `<dl>` elements with `<dt>` and `<dd>` used for terminology and glossaries.
- `aeo_numbered_steps`: Ordered lists `<ol>` used for procedural/instructional tasks.
- `aeo_tables`: `<table>` elements with structured comparative data.
- `aeo_lists`: Bullet lists `<ul>` present for itemized scannability.
- `aeo_code_blocks`: `<code>` or `<pre>` present for technical implementation clarity.
- `aeo_collapsible`: `<details>` and `<summary>` elements used for expandable Q&A content.

#### B. Voice, Readability & Tone
- `aeo_conversational`: Conversational pronoun density ("you", "your", "we") matching spoken query phrasing.
- `aeo_reading_level`: Average sentence length under 20 words for easy extraction.
- `aeo_active_voice`: Dominant active voice construction (>80%).
- `aeo_short_sentences`: High ratio of concise sentences.
- `aeo_concise_paragraphs`: Paragraphs maintained between 30 and 90 words.
- `aeo_summary`: Dedicated summary section (Key Takeaways, TL;DR, or Conclusion).

#### C. AI Discoverability & Crawlability (GEO Optimizer Integration)
- `aeo_ai_robots`: Robots directives allow general AI retrieval.
- `aeo_ai_crawlers`: Major citation bots allowed (`GPTBot`, `ClaudeBot`, `PerplexityBot`).
- `aeo_clean_url`: URL path is clean, lowercase, and descriptive.
- `aeo_semantic_html`: Rich HTML5 semantic tag utilization (`<article>`, `<section>`, `<main>`).

---

### Category 3: GEO — Generative Engine Optimization (Weight: 25%)

#### A. Empirical Evidence & Citability
- `geo_statistics_density`: Quantitative data points with numbers and units (e.g., %, ms, $).
- `geo_data_viz`: Comparative data tables or structured benchmarks.
- `geo_expert_quotes`: `<blockquote>` or attributed quotes from subject-matter experts.
- `geo_citations`: External outbound links to primary sources (academic papers, docs, standards).
- `geo_definitions`: Clear definition patterns ("X is a...", "defined as...") in top sections.
- `geo_passage_citability`: Self-contained 100–180 word answer blocks capable of zero-context citation.

#### B. E-E-A-T & Entity Authority
- `geo_author_schema`: `Person` author markup in JSON-LD.
- `geo_author_credentials`: Explicit byline stating author name, title, or credentials.
- `geo_org_schema`: `Organization` schema declared.
- `geo_knowledge_graph`: Authoritative `sameAs` links pointing to Wikidata, Wikipedia, LinkedIn, or GitHub.
- `geo_trust`: Presence of trust links (About Us, Contact, Privacy Policy, Terms).
- `geo_entity_markup`: Rich multi-entity schema graph (`Article`, `Organization`, `Product`, `FAQPage`).

#### C. Freshness & Content Depth
- `geo_freshness`: Machine-readable `<time datetime="...">` tags or visible current-year updates.
- `geo_content_depth`: Substantial content volume (>300 words, multi-section coverage).
- `geo_intent_clarity`: Title keywords prominently featured in H1 and first 200 words.
- `geo_format_variety`: Content combines at least 4 formatting types (bold, lists, tables, quotes, code).
- `geo_snippet_control`: Content avoids `data-nosnippet` over-blocking.

---

## 3. Scoring Engine Specification

### Category Calculation:
For each category $C \in \{\text{SEO}, \text{AEO}, \text{GEO}\}$:
$$
\text{Total}_C = \sum_{k \in C} [\text{applicable}_k \neq \text{false}]
$$
$$
\text{Passed}_C = \sum_{k \in C} [\text{pass}_k == \text{true} \land \text{applicable}_k \neq \text{false}]
$$
$$
\text{Score}_C = \text{round}\left( \frac{\text{Passed}_C}{\text{Total}_C} \times 100 \right)
$$

### Overall Composite Score:
$$
\text{Overall} = \text{round}(0.50 \times \text{Score}_{\text{SEO}} + 0.25 \times \text{Score}_{\text{AEO}} + 0.25 \times \text{Score}_{\text{GEO}})
$$

---

## 4. Implementation Structure

1. **`signals.go`:** Extract document signals from DOM (`golang.org/x/net/html`) and HTTP response headers.
2. **`engine.go`:** Evaluates the rule registry, maps check results, and computes scores.
3. **`main.go`:** Dispatches to the local audit engine by default; supports `-remote` for proxying through `seoscore.tools/api/scan`.
4. **`main_test.go` & `engine_test.go`:** Unit tests verifying rules and benchmark parity across all 10 synthetic archetypes.
