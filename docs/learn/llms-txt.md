---
title: "llms.txt: What It Is and How findable Checks It"
description: "llms.txt is a plain Markdown file at your site root that points AI tools to your key pages. See the llms-txt check and how to write the file."
slug: llms-txt
date: 2026-10-15
updated: 2026-10-15
author: "Nadim Tuhin"
faq:
  - q: "Is llms.txt a standard?"
    a: "No. It is a proposal. Not every AI tool reads it, and findable does not claim it raises citations."
  - q: "What does the llms-txt check test?"
    a: "Only that /llms.txt on your site's origin returns a 200 response. It does not grade the content."
  - q: "Does llms.txt replace robots.txt or sitemap.xml?"
    a: "No. robots.txt controls crawler access. sitemap.xml lists URLs for search engines. llms.txt is a short human-readable guide for language models."
  - q: "Where does the file go?"
    a: "At the root of the host, for example https://example.com/llms.txt. A file in a subfolder is not found."
---

# llms.txt: what it is and how findable checks it

llms.txt is a plain text file in Markdown that sits at the root of your site. It gives an AI tool a short map: who you are, and which pages matter most.

## What does findable check?

One Site check, `llms-txt`. It fetches `/llms.txt` from the origin of the page you scanned and passes when the response status is 200. The hosted API reads at most the first 512 KB; the CLI reads up to 10 MB. It does not judge the content.

The check is unscored and runs with the `-site` flag or the `site` option in the API. See [robots.txt and AI crawlers](/learn/robots-txt-ai-crawlers) for the other Site checks.

## What goes in the file?

A title, a one-line summary in a blockquote, and lists of links with a short note on each. This is the shape findable uses for its own file:

```markdown
# Example Co

> Example Co sells hand-made ceramic mugs and ships worldwide.

## Pages

- [Shop](https://example.com/shop): every mug we sell, with sizes and prices
- [Shipping](https://example.com/shipping): countries, costs and delivery times
- [Care guide](https://example.com/care): how to wash and store a mug

## Optional

- [Full text](https://example.com/llms-full.txt): every page described in full
```

## How do you write a useful one?

1. Open with a summary that states what you do in one sentence.
2. List the pages you would want quoted, not every page.
3. Describe each link in plain words. The note is the part a model reads.
4. Use absolute URLs.
5. Keep the file current. Regenerate it when pages change, or build it from your page list.

## What are the limits?

Nothing here guarantees that an AI tool will cite you. A page still has to be crawlable, answerable and trustworthy. The answer shape matters more: see [every AEO check](/learn/aeo-checks) and [every GEO check](/learn/geo-checks).

## Key takeaways

- Put `llms.txt` at the site root and make sure it returns 200.
- Write Markdown: title, summary, links with notes.
- It is a cheap proposal, not a ranking factor.
- findable's `llms-txt` check tests presence only.
