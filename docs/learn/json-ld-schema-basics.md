---
title: "JSON-LD Schema Basics: What findable Checks and Why"
description: "Learn what JSON-LD is, which schema.org types findable validates, and how to fix jsonld_valid, jsonld_context and empty or relative values."
slug: json-ld-schema-basics
date: 2026-10-15
updated: 2026-10-15
author: "findable team"
faq:
  - q: "What is JSON-LD?"
    a: "JSON-LD is a block of JSON in a script tag that describes the page using schema.org terms. Machines read it so they do not have to guess what the page is about."
  - q: "Where do I put the script tag?"
    a: "Anywhere in the HTML is valid. The head is the usual place. findable reads every application/ld+json block it finds."
  - q: "Does JSON-LD guarantee rich results?"
    a: "No. It makes your page eligible and easier to read. Search engines decide what to show."
  - q: "Do I need a separate block for each type?"
    a: "No. You can use several blocks, one object with a @graph array, or both. findable handles each."
---

# JSON-LD schema basics: what findable checks and why

JSON-LD tells machines what a page is. The text says "Alex Morgan wrote this on 15 October". JSON-LD says it in a form a program can read without guessing.

## What does a JSON-LD block look like?

It is a script tag with a JSON object inside. The object needs a `@context` and a `@type`.

```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "Article",
  "headline": "JSON-LD schema basics",
  "author": { "@type": "Person", "name": "Alex Morgan" },
  "datePublished": "2026-10-15",
  "image": "https://example.com/cover.jpg"
}
</script>
```

## Which checks cover the basics?

| Check id | What it tests | Fix |
| --- | --- | --- |
| `aeo_structured_data` | At least one JSON-LD block exists | Add a block |
| `jsonld_valid` | Every block parses as JSON | Fix the syntax error |
| `jsonld_context` | Each root has a schema.org `@context` | Add `"@context": "https://schema.org"` |
| `schema_empty_values` | No empty string values | Fill the field or remove it |
| `schema_urls_absolute` | `url`, `logo` and `image` are absolute http(s) URLs | Use the full URL |

The usual cause of `jsonld_valid` is a trailing comma or an unescaped quote, often from a template that pastes page text into the JSON.

## Which types does findable validate?

If a type is present, findable checks its required fields. If it is absent, those checks are skipped, so you are never punished for a type you do not need.

| Type | What findable looks for |
| --- | --- |
| `Organization` | name, logo and url |
| `Article` | headline of 110 characters or fewer, named author, valid ISO dates, an image |
| `FAQPage` | each question has a name and an answer text |
| `Product` | offers, plus a price and currency on the offers |
| `WebSite` | name and url, and a SearchAction with a target and query-input |
| `VideoObject` | name, thumbnailUrl and uploadDate |
| `BreadcrumbList` | 2 or more positioned, named items |
| `LocalBusiness` | address and telephone |
| `aggregateRating` | a ratingValue and a review or rating count |

Pick the type that is true. Do not add `Article` to a product page to collect a check.

## What are the common mistakes?

1. Marking up text that is not visible on the page.
2. Leaving a field as an empty string.
3. Using a relative path like `/logo.png` for `logo` or `image`.
4. Forgetting `@context`, so the block has no meaning to a parser.
5. Copying a plugin's output for a type that does not describe your page.

## Where do you go next?

FAQ markup has its own guide: [FAQ schema for answer engines](/learn/faq-schema-answer-engines). The full list of AEO ids is in [every AEO check](/learn/aeo-checks).

## Key takeaways

- Add at least one valid JSON-LD block with a schema.org `@context`.
- Use absolute URLs and no empty values.
- Mark up only what is visible and true.
- Missing types are skipped, not failed.
