---
title: "Canonical Tags and Duplicate Content: How to Get Them Right"
description: "Use one canonical link per page that points to its own absolute URL. Learn the checks canonical, canonical_self and canonical_absolute and fix them."
slug: canonical-and-duplicate-content
date: 2026-10-15
updated: 2026-10-15
author: "findable team"
faq:
  - q: "What is a canonical tag?"
    a: "It is a link in the head that names the preferred address for a page. Search engines use it to pick one URL when the same content is reachable at several."
  - q: "Should a page point its canonical at itself?"
    a: "Usually yes. A self-referencing canonical is the safe default. Point elsewhere only when the page really is a copy of another."
  - q: "Is the canonical a command?"
    a: "No. It is a strong hint. Search engines can pick a different URL if other signals disagree."
  - q: "Why does findable skip canonical_self?"
    a: "It needs a canonical to compare against. With no canonical tag, canonical_self and canonical_absolute are skipped and canonical fails."
---

# Canonical tags and duplicate content: how to get them right

The same page is often reachable at many addresses. With and without `www`, with tracking parameters, with a trailing slash, over http and https. To a crawler these can be separate pages with identical content. A canonical tag says which one counts.

## Which checks cover canonical?

| Check id | What it tests | Passes when |
| --- | --- | --- |
| `canonical` | A canonical link is present | The tag exists |
| `canonical_self` | The canonical points to this page | The canonical resolves to the scanned URL |
| `canonical_absolute` | The canonical is an absolute URL | It starts with a scheme like `https:` |
| `og_url_matches_canonical` | `og:url` agrees with the canonical | Both name the same page |
| `meta_robots_conflict` | Robots directives do not contradict | No `index` with `noindex` |

## What does the tag look like?

```html
<link rel="canonical" href="https://example.com/care">
```

Use the full URL, with the scheme and host. A relative value like `/care` is easy to break and fails `canonical_absolute`.

## When should the canonical point elsewhere?

When the page is a deliberate copy: a print version, a filtered list, or a tracking-parameter variant. Then `canonical_self` will fail, and that is correct. Decide on purpose. If a page is meant to rank by itself, it should point to itself.

## What are the common mistakes?

1. Canonical pointing at the home page from every page.
2. A canonical that differs from the redirect target or from `og:url`.
3. Two canonical tags that disagree.
4. A canonical pointing at a URL that is blocked or marked `noindex`. See [robots.txt and AI crawlers](/learn/robots-txt-ai-crawlers).
5. Mixing `index` and `noindex` in one robots meta.

## How do you reduce duplicate content in the first place?

1. Redirect `http` to `https` and pick one hostname.
2. Pick a trailing slash policy and keep it.
3. Keep parameters out of internal links.
4. Give each page its own title and description. See [meta descriptions](/learn/meta-descriptions).

## Key takeaways

- One absolute canonical per page, normally pointing at itself.
- Point it elsewhere only for true copies.
- Keep `og:url` equal to the canonical.
- More SEO ids are in [every SEO check](/learn/seo-checks).
