---
title: "Every findable SEO Check ID and How to Fix It"
description: "A plain guide to all 98 SEO checks in findable, from title_length to csp. See what each id tests, when it passes and the fix. SEO is 50% of your score."
slug: seo-checks
date: 2026-10-14
updated: 2026-10-14
author: "findable team"
faq:
  - q: "How much of the findable score is SEO?"
    a: "SEO is 50% of the score. AEO and GEO are 25% each. The SEO family has the most checks and covers metadata, headings, links, images, social tags, security headers and markup."
  - q: "Do I need to pass every SEO check?"
    a: "No. Some checks are rules of thumb, like word_count and dom_size. Treat them as prompts to look at the page, not as laws. Fix the failing basics first, then decide on the rest."
  - q: "How do I see only SEO failures for my page?"
    a: "Run findable on the URL and read the SEO family card. The issue list puts failing checks first, and each row shows the check id from this guide."
  - q: "Can I run these checks in CI?"
    a: "Yes. Use the -fail-under flag, for example findable -fail-under 80 https://example.com. The command exits with code 1 when the score is below your number."
---

# Every SEO check findable runs, and how to fix each one

You ran an audit and got a list of ids like `title_length` and `img_dimensions`. This guide explains every one of them. Each check is a simple test on your HTML or response headers. Fix the failing ones and your score goes up.

## What does the SEO family cover in findable?

SEO is 50% of your findable score, the biggest share. It has 98 checks covering metadata, structured data, performance and accessibility - the basics: metadata, page setup, headings, content size, links, images, social tags, security headers and markup. Each check has a fixed id, so you can search for it in any report.

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
| `canonical_self` | Canonical points to this page | Point canonical at the page itself |
| `canonical_absolute` | Canonical is an absolute URL | Use a full https URL |
| `meta_robots_conflict` | Robots meta has no contradictory directives | Drop the conflicting index or follow token |
| `robots_nofollow` | Robots meta does not nofollow every link | Remove nofollow or none |
| `title_single` | At most one title tag in head | Keep a single title |
| `desc_single` | At most one meta description | Keep a single description |
| `url_length` | Path and query of 100 characters or fewer | Shorten the URL |
| `url_query_params` | 2 or fewer query parameters | Use clean paths over parameters |
| `url_depth` | 5 or fewer path levels | Flatten the URL structure |
| `favicon` | A favicon link | Add a `rel=icon` link |
| `viewport_device_width` | Viewport uses width=device-width (N/A without a viewport) | Set `width=device-width` |
| `viewport_scalable` | Viewport allows pinch zoom (N/A without a viewport) | Remove `user-scalable=no` and low maximum-scale |
| `lang_valid` | The html lang value is a valid tag (N/A without lang) | Use a code like `en` or `pt-BR` |
| `hreflang_valid` | Hreflang codes are valid (N/A without hreflang) | Use language-region codes or `x-default` |
| `hreflang_absolute` | Hreflang URLs are absolute (N/A without hreflang) | Use full URLs |
| `charset_early` | Charset declared in the first 1024 bytes (N/A without charset) | Put the charset meta first in head |
| `img_alt_quality` | Alt text is not a filename or placeholder (N/A without images) | Describe the image |
| `img_alt_length` | Alt text of 125 characters or fewer (N/A without images) | Shorten long alt text |
| `img_lazy_loading` | At most 5 images load eagerly; lazy-load the rest (N/A without images) | Add `loading=lazy` |
| `img_modern_format` | Half or more images are WebP, AVIF or SVG (N/A without images) | Convert to WebP or AVIF |
| `nofollow_internal` | No internal link marked nofollow | Remove nofollow from your own links |
| `generic_anchor_text` | No "click here" style anchors | Write descriptive link text |
| `empty_anchor_text` | No links without text, label or image | Add text or an aria-label |
| `javascript_hrefs` | No `javascript:` hrefs | Use real URLs |
| `links_count` | 200 or fewer links | Trim link-heavy navigation |
| `mixed_content` | No http resources on an https page (N/A on http) | Load everything over https |
| `render_blocking_scripts` | No blocking scripts in head | Add `defer` or `async` |
| `script_count` | 20 or fewer external scripts | Remove or bundle scripts |
| `stylesheet_count` | 6 or fewer stylesheets | Combine stylesheets |
| `inline_css_size` | 20000 bytes or less of inline CSS | Move large CSS to a file |
| `resource_hints` | Third-party resources have preconnect or preload | Add `rel=preconnect` |
| `html_weight` | HTML of 500000 bytes or less | Cut inline data and markup |
| `skip_link` | A skip-to-content link | Add a "Skip to content" link |
| `button_names` | Every button has an accessible name | Add button text or aria-label |
| `table_headers` | At least one `th` per table | Use `th` header cells |
| `iframe_title` | Every iframe has a title | Add a `title` attribute |
| `color_contrast` | Inline text and background colors reach WCAG AA contrast, 4.5:1 or 3:1 for h1-h3 (N/A without inline color pairs) | Darken the text or lighten the background |
| `jsonld_valid` | All JSON-LD blocks parse (N/A without JSON-LD) | Fix the JSON syntax |
| `jsonld_context` | Every JSON-LD root has a schema.org @context | Add `@context` |
| `schema_empty_values` | No empty strings in structured data | Fill or remove empty fields |
| `schema_urls_absolute` | url, logo and image values are absolute | Use full http(s) URLs |
| `org_name` | Organization has a name (N/A without Organization) | Add `name` |
| `org_logo` | Organization has a logo (N/A without Organization) | Add `logo` |
| `org_url` | Organization has a url (N/A without Organization) | Add `url` |
| `website_name_url` | WebSite has name and url (N/A without WebSite) | Add both fields |
| `searchaction_valid` | SearchAction has a target template (N/A without it) | Add a `target` with `{search_term_string}` |
| `breadcrumb_valid` | BreadcrumbList items are complete (N/A without it) | Give each item a position and name |
| `faq_valid` | FAQPage questions have a name and answer text | Fill every question and answer |
| `article_headline` | Article headline of 110 characters or fewer | Add a short `headline` |
| `article_author` | Article names an author | Add `author` with a name |
| `article_dates` | Article dates are valid ISO dates | Use `2025-01-31` style dates |
| `article_image` | Article has an image | Add `image` |
| `product_offers` | Product has offers (N/A without Product) | Add `offers` |
| `product_price` | Product offer has a price and currency (N/A without Product) | Add price and priceCurrency |
| `product_rating` | aggregateRating is complete (N/A without it) | Add ratingValue and a count |
| `localbusiness_address` | LocalBusiness has an address (N/A without it) | Add `address` |
| `localbusiness_phone` | LocalBusiness has a telephone (N/A without it) | Add `telephone` |
| `video_fields` | VideoObject has name, thumbnail and date (N/A without it) | Fill the required fields |
| `og_type` | `og:type` is set | Add `og:type` |
| `og_site_name` | `og:site_name` is set | Add `og:site_name` |
| `og_image_absolute` | `og:image` is an absolute URL | Use a full URL |
| `og_url_matches_canonical` | `og:url` matches canonical | Make both URLs the same |

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

Five checks look at transport and headers: `https`, `hsts`, `x_content_type`, `x_frame_options` and `csp`. They test your server response, not your HTML. A check passes when the header is present. It does not judge how strong the value is. If you audit a local file with `findable ./page.html`, there are no response headers, so expect these five to fail.

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
- There are 98 SEO checks. Each has a fixed id and a simple pass test.
- Fix the head first: title, description, charset, viewport, lang, doctype.
- `word_count`, `text_html_ratio` and `dom_size` are heuristics. Use judgment.
- Security header checks pass when the header exists. They do not rate its strength.
- Run `findable -fail-under 80 URL` in CI to catch regressions.
