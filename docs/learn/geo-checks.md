---
title: "Every findable GEO Check ID, Plus AI Crawler Checks"
description: "All 27 GEO checks in findable explained, plus the unscored Site checks for GPTBot, ClaudeBot and llms.txt. GEO is 25% of the score. Fixes included."
slug: geo-checks
date: 2026-10-14
updated: 2026-10-14
author: "Nadim Tuhin"
faq:
  - q: "What is GEO in findable?"
    a: "GEO means generative engine optimization. The checks look for things that make a page easy to cite in a generated answer: sources, numbers, quotes, authorship, entity markup and fresh dates."
  - q: "How much of the score is GEO?"
    a: "GEO is 25% of the findable score. SEO is 50% and AEO is 25%."
  - q: "Are the Site checks part of the score?"
    a: "No. The Site checks run only with the -site flag and are unscored. They test robots.txt, AI bot access and llms.txt, which still decide whether AI crawlers can read you."
  - q: "Does llms.txt make AI tools cite me?"
    a: "No promise. llms.txt is a proposal, not a standard every crawler follows. findable only checks that the file exists."
---

# Every GEO check findable runs, and how to fix each one

AI tools now write answers and cite pages. GEO is the work of making your page worth citing. findable runs 27 GEO checks, and a short set of unscored Site checks for AI crawlers. This guide covers both. Remember that these are heuristics. They test signals, not quality.

## What does the GEO family test in findable?

GEO is 25% of your findable score. It has 27 checks on citability: numbers, tables, quotes, outbound sources, clear definitions, author and organization markup, trust pages, fresh dates and content length. Each check is a pattern match, so a pass shows the signal exists, not that it is good.

| Check id | What it checks | Fix in one line |
| --- | --- | --- |
| `geo_statistics_density` | 3 or more numbers with units | State real figures with units |
| `geo_data_viz` | A table is present | Show data in a table |
| `geo_comparison` | A table or " vs " in text | Compare options side by side |
| `geo_expert_quotes` | A blockquote or expert quote | Quote a named source |
| `geo_citations` | 1 or more external link | Link to your sources |
| `geo_definitions` | " is a ", " is the " or " defined as " | Define key terms plainly |
| `geo_passage_citability` | A paragraph of 40 to 120 words | Write self-contained paragraphs |
| `geo_author_schema` | Person schema or author in schema | Add author to JSON-LD |
| `geo_author_credentials` | A visible "by " byline or schema author | Show who wrote it |
| `geo_org_schema` | Organization or LocalBusiness schema | Add Organization JSON-LD |
| `geo_knowledge_graph` | `sameAs` links in schema | Link your profiles |
| `geo_trust` | Internal links to about and contact | Link both pages |
| `geo_entity_markup` | 2 or more schema types | Mark up more entities |
| `geo_freshness` | `<time>`, a year in text, or `dateModified` | Show a date |
| `geo_time_element` | A `<time>` element | Wrap dates in `<time>` |
| `geo_content_length` | 300 or more words | Write enough to be useful |
| `geo_intent_clarity` | 2 or more title words (over 3 letters) in the first heading | Match h1 to title |
| `geo_format_variety` | 3 or more of lists, tables, quotes, code, dl | Mix formats |
| `geo_snippet_control` | No `nosnippet` in the robots meta | Remove `nosnippet` |
| `geo_date_published` | A valid publish date in schema or meta (N/A unless the page is an article) | Add `datePublished` |
| `geo_date_modified` | A valid modified date in schema or meta (N/A unless the page is an article) | Add `dateModified` |
| `geo_modified_after_published` | Modified is not before published (N/A without both) | Fix the date order |
| `geo_visible_date` | A visible "Updated" or "Published" date (N/A unless the page is an article) | Show the date in the text |
| `geo_authoritative_links` | A link to a .gov, .edu or reference site (N/A unless the page is an article) | Cite primary sources |
| `geo_source_diversity` | 3 or more distinct external domains (N/A unless the page is an article) | Cite more than one source |
| `geo_quote_attribution` | Blockquotes are attributed (N/A without blockquotes) | Add `cite` or a `cite` element |
| `geo_policy_links` | A privacy or terms link | Link your policies |

## How do I fix the evidence checks?

Five checks reward evidence: `geo_statistics_density`, `geo_data_viz`, `geo_comparison`, `geo_expert_quotes` and `geo_citations`. Add real numbers with units, a table, a named quote and links to sources. Never invent a figure to pass. A fake number hurts trust. The stat matcher wants a unit it knows, such as `%`, `ms`, `MB`, `px`, `hours` or a `$` amount; "300 words" does not count.

```html
<blockquote>
  <p>Quote text from a named person.</p>
  <footer>Name, role, <a href="https://example.org/source">source</a></footer>
</blockquote>
<p>Source: <a href="https://arxiv.org/abs/2311.09735">GEO: Generative Engine Optimization</a>.</p>
```

The [GEO paper](https://arxiv.org/abs/2311.09735) by Aggarwal et al. tested ideas like adding citations, quotations and statistics. In their tests, these helped visibility. Results will vary by engine and topic. `geo_expert_quotes` is satisfied by any `<blockquote>` with text in it; nothing checks that a named person said it.

## How do I fix the passage and definition checks?

Three checks cover passages: `geo_definitions`, `geo_passage_citability` and `geo_content_length`. State what a thing is using "is a" or "is the". Write at least one paragraph of 40 to 120 words that stands alone. Reach 300 words in total. Only `<p>` text counts toward the passage check, so an answer written as a list or inside a `<div>` is never seen.

A passage is citable when you can lift it out and it still makes sense. Name the subject in the first sentence. Avoid "this" and "it" with no clear referent. The `geo_definitions` test only looks for the phrases " is a ", " is the " or " defined as ", so it is easy to pass and easy to fool.

## How do I fix the author and organization checks?

Six checks cover who is behind the page: `geo_author_schema`, `geo_author_credentials`, `geo_org_schema`, `geo_knowledge_graph`, `geo_trust` and `geo_entity_markup`. Show a byline, add Person and Organization JSON-LD with `sameAs` links, link to your about and contact pages, and use at least 2 schema types.

```html
<p>By Nadim Tuhin</p>
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@graph": [
    {"@type": "Organization", "name": "Example", "url": "https://example.com",
     "sameAs": ["https://github.com/example"]},
    {"@type": "Person", "name": "Nadim Tuhin"}
  ]
}
</script>
```

The `geo_trust` check looks for the strings "about" and "contact" in your internal link URLs, so `/about-us` and `/contact-sales` both count. The schema author only counts when `author` is an object with a `name`; a bare string is ignored. See [schema.org](https://schema.org/) for types, and Google's [structured data intro](https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data).

## How do I fix the freshness, intent and format checks?

Five checks cover shape and dates: `geo_freshness`, `geo_time_element`, `geo_intent_clarity`, `geo_format_variety` and `geo_snippet_control`. Show a date in a `<time>` element, share words between title and first heading, mix 3 formats, and avoid `nosnippet` in the robots meta. The five formats counted are lists, tables, blockquotes, code and definition lists.

```html
<time datetime="2026-10-14">October 14, 2026</time>
```

`geo_freshness` passes on a `<time>` element, a year from 2024 to 2029 in the text, or a `dateModified` in schema. A year in the text is a weak signal, and a 2023 date does not count. `geo_intent_clarity` counts shared words longer than 3 letters between the title and the first heading. `geo_snippet_control` fails if `nosnippet` limits what can be shown.

## What are the Site checks and why do AI crawlers care?

The Site checks run only with the `-site` flag and are unscored. They fetch `/robots.txt` and `/llms.txt` and test whether named AI bots may read your page. They matter because a blocked crawler cannot read even a perfect page. `-site` needs an http or https URL; it refuses a local HTML file.

```
findable -site https://example.com
```

| Check id | What it checks |
| --- | --- |
| `robots-txt` | `/robots.txt` exists |
| `robots-page` | robots.txt allows `*` to fetch the page |
| `robots-gptbot` | robots.txt allows GPTBot |
| `robots-claudebot` | robots.txt allows ClaudeBot |
| `robots-perplexitybot` | robots.txt allows PerplexityBot |
| `robots-google-extended` | robots.txt allows Google-Extended |
| `llms-txt` | `/llms.txt` exists |

## How do I let AI bots in, or keep them out?

Edit `/robots.txt`. A bot you name with `Disallow: /` is blocked. A bot you do not mention follows the `*` rules. The choice is yours. A bot that cannot fetch the page cannot cite it. Blocking one keeps your text out of that vendor's crawl. findable only reports the current state.

```
User-agent: GPTBot
Allow: /

User-agent: ClaudeBot
Allow: /

User-agent: PerplexityBot
Allow: /

User-agent: Google-Extended
Allow: /
```

Official docs: [OpenAI crawlers](https://platform.openai.com/docs/bots), [Anthropic ClaudeBot](https://support.anthropic.com/en/articles/8896518-does-anthropic-crawl-data-from-the-web-and-how-can-site-owners-block-the-crawler), [Perplexity bots](https://docs.perplexity.ai/guides/bots), [Google common crawlers](https://developers.google.com/search/docs/crawling-indexing/google-common-crawlers) and the [robots.txt intro](https://developers.google.com/search/docs/crawling-indexing/robots/intro). Note that robots.txt is a request, not a lock.

For `llms-txt`, add a plain Markdown file at `/llms.txt` that lists your key pages. See the [llms.txt proposal](https://llmstxt.org/). It is a proposal, and no one promises that every AI tool reads it.

## Key takeaways

- GEO is 25% of the findable score and has 27 checks.
- Real sources, quotes, tables and numbers with units make a page easier to cite.
- Never invent a statistic to pass `geo_statistics_density`.
- Show a byline, add Person and Organization JSON-LD, and link about and contact pages.
- Site checks are unscored. Run them with `-site`.
- Check `robots-gptbot`, `robots-claudebot`, `robots-perplexitybot` and `robots-google-extended`, then decide who you want in.
- Add `/llms.txt` if you want to, knowing it is only a proposal.
