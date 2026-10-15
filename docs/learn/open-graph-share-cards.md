---
title: "Open Graph and Twitter Cards: Fix Your Share Previews"
description: "Add og:title, og:description, og:image, og:url and twitter:card so links unfurl well. See every social check findable runs and how to pass it."
slug: open-graph-share-cards
date: 2026-10-15
updated: 2026-10-15
author: "Nadim Tuhin"
faq:
  - q: "What is Open Graph?"
    a: "Open Graph is a set of meta tags that tell chat apps and social networks which title, summary and image to show when someone shares your link."
  - q: "Do I need both Open Graph and Twitter tags?"
    a: "Most apps read Open Graph. The twitter:card tag chooses the card layout on X. findable checks that both are present."
  - q: "Why does my old preview not update?"
    a: "Platforms cache previews. Fix the tags first, then use the platform's own debugger or sharing tool to refresh the cache."
  - q: "What image size should I use?"
    a: "findable only checks that og:image exists and is an absolute URL. A wide image around 1200 by 630 pixels is a common choice, but check each platform's guidance."
---

# Open Graph and Twitter cards: fix your share previews

When someone pastes your link into a chat or a feed, the app fetches your page and reads a few meta tags. If they are missing, it guesses, and the preview looks bad. findable tests the tags below.

## Which checks cover social tags?

| Check id | What it tests |
| --- | --- |
| `og_title` | `og:title` is present |
| `og_desc` | `og:description` is present |
| `og_image` | `og:image` is present |
| `og_url` | `og:url` is present |
| `twitter_card` | A `twitter:card` meta is present |
| `og_type` | `og:type` is set |
| `og_site_name` | `og:site_name` is set |
| `og_image_absolute` | `og:image` is an absolute http(s) URL |
| `og_url_matches_canonical` | `og:url` matches the canonical URL |

The last two are skipped when there is nothing to compare.

## What is the minimum set of tags?

```html
<meta property="og:type" content="article">
<meta property="og:site_name" content="Example Co">
<meta property="og:title" content="How to care for a ceramic mug">
<meta property="og:description" content="Wash, dry and store a hand-made mug so it lasts for years.">
<meta property="og:url" content="https://example.com/care">
<meta property="og:image" content="https://example.com/care-cover.jpg">
<meta name="twitter:card" content="summary_large_image">
```

Use `summary_large_image` for a wide picture and `summary` for a small one.

## What goes wrong most often?

1. A relative image path such as `/cover.jpg`. Scrapers do not resolve it the way your browser does. Use the full URL.
2. `og:url` pointing at a different address than the canonical. Keep them the same. See [canonical and duplicate content](/learn/canonical-and-duplicate-content).
3. Reusing one image and one description on every page.
4. Setting the tags in JavaScript. Share scrapers usually do not run it, so put the tags in the served HTML.
5. Forgetting that a platform cached an old preview.

## How do descriptions fit in?

`og:description` can differ from your meta description, but they should agree. Write the meta description first: [meta descriptions](/learn/meta-descriptions).

## Key takeaways

- Put the five core tags plus `og:type` and `og:site_name` in the served HTML.
- Make `og:image` an absolute URL.
- Make `og:url` equal to the canonical.
- Refresh the platform cache after you fix things.
- The full SEO list is in [every SEO check](/learn/seo-checks).
