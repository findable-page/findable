---
title: "Every findable SEO Check ID and How to Fix It"
description: "A plain guide to all 36 SEO checks in findable, from title_length to csp. See what each id tests, when it passes and the fix. SEO is 50% of your score."
slug: seo-checks
date: 2026-10-14
updated: 2026-10-14
author: "Nadim Tuhin"
faq:
  - q: "How much of the findable score is SEO?"
    a: "SEO is 50% of the score. AEO and GEO are 25% each. The SEO family has the most checks and covers metadata, headings, links, images, social tags, security headers and markup."
  - q: "Do I need to pass every SEO check?"
    a: "No. Some checks are rules of thumb, like word_count and dom_size. Treat them as prompts to look at the page, not as laws. Fix the failing basics first, then decide on the rest."
  - q: "How do I see only SEO failures for my page?"
    a: "Run seoscore on the URL and read the SEO family card. The issue list puts failing checks first, and each row shows the check id from this guide."
  - q: "Can I run these checks in CI?"
    a: "Yes. Use the -fail-under flag, for example seoscore -fail-under 80 https://example.com. The command exits with code 1 when the score is below your number."
---

# Every SEO check findable runs, and how to fix each one

You ran an audit and got a list of ids like `title_length` and `img_dimensions`. This guide explains every one of them. Each check is a simple test on your HTML or response headers. Fix the failing ones and your score goes up.

## What does the SEO family cover in findable?

SEO is 50% of your findable score, the biggest share. It has 36 checks on the basics: metadata, page setup, headings, content size, links, images, social tags, security headers and markup. Each check has a fixed id, so you can search for it in any report.

| Check id | What it checks | Fix in one line |
| --- | --- | --- |
| `title_exists` | A title tag is present | Add `<title>` in the head |
| `title_length` | Title is 30 to 60 characters | Rewrite the title to fit |
| `desc_exists` | A meta description is present | Add `<meta name="description">` |
| `desc_length` | Description is 120 to 160 characters | Trim or expand it |
| `viewport` | A viewport meta tag is present | Add the viewport tag |
| `charset` | UTF-8 is declared | Add `<meta charset="utf-8">` |
| `lang_attr` | `<html>` has a lang attribute | Set `lang="en"` or your language |
| `html_doctype` | `<!DOCTYPE html>` is present | Put it on line one |
| `meta_refresh` | No meta refresh tag | Use a server redirect instead |
| `robots_meta` | No `noindex` in the robots meta | Remove `noindex` |
| `h1_exists` | Exactly one h1 | Keep one h1 per page |
| `h2_exists` | At least one h2 | Split content into sections |
| `heading_hierarchy` | No skipped heading levels | Go h1, h2, h3 in order |
| `word_count` | 300 or more words | Add useful body text |
| `text_html_ratio` | Text is at least 15% of the HTML | Cut markup bloat |
| `dom_size` | Fewer than 1500 elements | Simplify the markup |
| `internal_links` | 2 or more internal links | Link to related pages |
| `external_links` | 1 or more external link | Cite a source |
| `empty_hrefs` | No empty href values | Fix or remove the link |
| `link_noopener` | `target="_blank"` links have `rel="noopener"` | Add the rel attribute |
| `img_alt` | Every image has alt text | Add alt to each img |
| `img_dimensions` | Every image has width and height | Add both attributes |
| `canonical` | A canonical link is present | Add `<link rel="canonical">` |
| `og_title` | Open Graph title is present | Add `og:title` |
| `og_desc` | Open Graph description is present | Add `og:description` |
| `og_image` | Open Graph image is present | Add `og:image` |
| `og_url` | Open Graph URL is present | Add `og:url` |
| `twitter_card` | A `twitter:card` meta is present | Add the card tag |
| `https` | Page is served over HTTPS | Install a certificate |
| `hsts` | Strict-Transport-Security header is set | Send the header |
| `x_content_type` | `X-Content-Type-Options: nosniff` is set | Send the header |
| `x_frame_options` | X-Frame-Options header is set | Send the header |
| `csp` | Content-Security-Policy header is set | Send a policy |
| `jsonld` | At least one JSON-LD block | Add structured data |
| `aria_landmarks` | `<main>` and `<nav>` are present | Use landmark elements |
| `duplicate_ids` | No duplicate id attributes | Make each id unique |

## How do I fix the title and description checks?

Four checks cover this: `title_exists`, `title_length`, `desc_exists` and `desc_length`. The title must exist and be 30 to 60 characters. The description must exist and be 120 to 160 characters. Search engines use both to build your result snippet, so write them for people.

```html
<title>Every SEO Check ID and How to Fix It</title>
<meta name="description" content="A plain guide to the SEO checks, with the fix for each one.">
```

The length ranges are common guidance, not a hard rule from Google. Google may rewrite your snippet anyway. See the [SEO starter guide](https://developers.google.com/search/docs/fundamentals/seo-starter-guide) for what it says about titles and snippets.

## How do I fix the page setup checks?

Six checks cover basic page setup: `viewport`, `charset`, `lang_attr`, `html_doctype`, `meta_refresh` and `robots_meta`. Most are one-line additions to the head. The last two are about things you must not have: a meta refresh and a `noindex` directive. `html_doctype` ignores leading whitespace but fails if a byte order mark comes before the doctype.

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
```

If `robots_meta` fails, look for `noindex` in `<meta name="robots">`. It tells crawlers to skip the page. If `meta_refresh` fails, replace the tag with a 301 redirect on the server.

## How do I fix the heading and content checks?

Six checks look at structure and size: `h1_exists`, `h2_exists`, `heading_hierarchy`, `word_count`, `text_html_ratio` and `dom_size`. Use one h1, then h2 sections, then h3 inside them. Never jump from h2 to h4. Write at least 300 words of real text. A page with two h1 tags fails `h1_exists` just like a page with none.

Be honest about three of these. `word_count`, `text_html_ratio` (15%) and `dom_size` (1500 elements) are heuristics. Google does not publish these cutoffs. A short page can still be great. Use them as a smell test for thin or bloated pages.

## How do I fix the link checks?

Four checks cover links: `internal_links`, `external_links`, `empty_hrefs` and `link_noopener`. You need at least 2 internal links and 1 external link. No link may have an empty href. Every `target="_blank"` link needs `rel="noopener"`. A link to another host counts as external, so a link to your own subdomain is external too.

```html
<a href="https://example.org/source" target="_blank" rel="noopener">Source</a>
```

The `noopener` value stops the new page from reaching your page through `window.opener`. `rel="noreferrer"` also passes the check, since it implies noopener. Read the [MDN page on rel=noopener](https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/rel/noopener) for details.

## How do I fix the image checks?

Two checks cover images: `img_alt` and `img_dimensions`. Every image needs an alt attribute and explicit width and height. The alt check passes when the page has no images. The size attributes let the browser reserve space, so the layout does not jump.

```html
<img src="/chart.png" alt="Bar chart of audit scores by family" width="800" height="450">
```

One catch: findable treats `alt=""` as missing, because the check trims whitespace and tests for empty. Decorative images with an empty alt are correct HTML, but they will still fail `img_alt`. Decide which matters more to you. Do not stuff keywords into alt text. Describe the image.

## How do I fix the canonical and social tag checks?

Six checks cover sharing and duplicates: `canonical`, `og_title`, `og_desc`, `og_image`, `og_url` and `twitter_card`. The canonical link names the preferred URL. The Open Graph and Twitter tags control how your link looks when shared. See the [Open Graph protocol](https://ogp.me/). Each check only tests that the tag has a non-empty value, so a broken `og:image` URL still passes.

```html
<link rel="canonical" href="https://example.com/page">
<meta property="og:title" content="Page title">
<meta property="og:description" content="Short summary.">
<meta property="og:image" content="https://example.com/card.png">
<meta property="og:url" content="https://example.com/page">
<meta name="twitter:card" content="summary_large_image">
```

## How do I fix the security header checks?

Five checks look at transport and headers: `https`, `hsts`, `x_content_type`, `x_frame_options` and `csp`. They test your server response, not your HTML. A check passes when the header is present. It does not judge how strong the value is. If you audit a local file with `seoscore ./page.html`, there are no response headers, so expect these five to fail.

```
Strict-Transport-Security: max-age=31536000
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Content-Security-Policy: default-src 'self'
```

Set these in your web server or CDN. Start the CSP in report-only mode so you do not break scripts. The [MDN CSP guide](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP) explains the options.

## How do I fix the structured data and markup checks?

Three checks cover markup: `jsonld`, `aria_landmarks` and `duplicate_ids`. Add at least one JSON-LD block, use both `<main>` and `<nav>`, and keep every id unique. Duplicate ids break anchors and scripts, and they confuse screen readers. The `jsonld` check drops any block that is not valid JSON, so one stray trailing comma makes it fail.

```html
<script type="application/ld+json">
{"@context":"https://schema.org","@type":"Article","headline":"Every SEO check"}
</script>
```

Read Google's [structured data intro](https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data) to pick a type that fits your page.

## Key takeaways

- SEO is 50% of the findable score, so start here.
- There are 36 SEO checks. Each has a fixed id and a simple pass test.
- Fix the head first: title, description, charset, viewport, lang, doctype.
- `word_count`, `text_html_ratio` and `dom_size` are heuristics. Use judgment.
- Security header checks pass when the header exists. They do not rate its strength.
- Run `seoscore -fail-under 80 URL` in CI to catch regressions.
