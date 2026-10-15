---
title: "How to Write a Meta Description That Passes findable"
description: "Write a meta description of 120 to 160 characters, one per page, that matches the page. See the checks desc_exists, desc_length and desc_single."
slug: meta-descriptions
date: 2026-10-15
updated: 2026-10-15
author: "Nadim Tuhin"
faq:
  - q: "Does a meta description change my ranking?"
    a: "Not directly. It is a summary that search results and link previews often show. A clear one can earn more clicks, but findable only tests that it exists, has a sane length and appears once."
  - q: "Will Google always show my description?"
    a: "No. Search engines may rewrite it from the page text. Write it as if it will be shown, and make the first sentence of the page a good fallback."
  - q: "Should every page have its own description?"
    a: "Yes. Copying one description across pages gives each page the same summary in results, so none of them stand out."
---

# How to write a meta description that passes findable

A meta description is one sentence or two in the head of your page. Search results and link previews often use it as the summary under your title. findable runs three checks on it.

## What does findable check about meta descriptions?

| Check id | What it tests | Passes when |
| --- | --- | --- |
| `desc_exists` | A meta description is present | The tag exists and has content |
| `desc_length` | The length in characters | 120 to 160 characters |
| `desc_single` | How many description tags there are | Zero or one, never two |

The length range is a rule of thumb. Longer text is usually cut off in results, and very short text wastes the space. The check counts characters, not pixels, so treat it as a guide.

## What does the tag look like?

Put one tag inside `<head>`:

```html
<meta name="description" content="Plain guides to SEO, AEO and GEO and to every check findable runs, so you can see why a page is hard to find and fix it.">
```

That example is 120 characters, inside the range.

## How do you write a good one?

1. Say what the page offers in the first half. Readers skim.
2. Use the words someone would type to find the page. Do not stuff them.
3. Write one description per page. Do not reuse the site tagline everywhere.
4. Do not repeat the title word for word. Add something the title leaves out.
5. Count characters before you ship. Most editors show a count.

## Why does desc_single fail?

It fails when a page has two description tags. This usually happens when a theme adds one and an SEO plugin adds another. Search engines pick one of them, and you lose control of which. Remove one source.

## Where else does the description get used?

Open Graph and Twitter tags have their own description fields. See [Open Graph and share cards](/learn/open-graph-share-cards). The `description` in your JSON-LD is a third copy, and it should say the same thing. See [JSON-LD basics](/learn/json-ld-schema-basics).

## Key takeaways

- Add exactly one `<meta name="description">` per page.
- Keep it between 120 and 160 characters.
- Write a different one for each page.
- Fix `desc_single` by removing the duplicate source, not by editing text.
- The other basic metadata checks are in [every SEO check](/learn/seo-checks).
