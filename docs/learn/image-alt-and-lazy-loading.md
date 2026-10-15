---
title: "Image Alt Text, Dimensions and Lazy Loading Explained"
description: "Fix img_alt, img_dimensions, img_alt_quality, img_lazy_loading and img_modern_format. See what each image check in findable tests and how to pass."
slug: image-alt-and-lazy-loading
date: 2026-10-15
updated: 2026-10-15
author: "Nadim Tuhin"
faq:
  - q: "Should decorative images have alt text?"
    a: "Yes, for truly decorative images an empty alt is correct under WCAG and screen readers skip them. findable still counts an empty alt as missing, so img_alt fails. Prefer short real text, or move decoration to CSS."
  - q: "Should I lazy load every image?"
    a: "No. Lazy load images below the first screen. The main image at the top should load right away, and fetchpriority high marks it as important."
  - q: "Why do I need width and height?"
    a: "They let the browser reserve space before the image arrives, so the page does not jump while it loads."
  - q: "Is WebP required?"
    a: "No. The check passes when at least half of your recognised images are WebP, AVIF or SVG."
---

# Image alt text, dimensions and lazy loading explained

Images affect three things findable cares about: whether a machine can understand them, whether the page jumps as it loads, and how much data loads up front.

## Which checks cover images?

| Check id | What it tests | Passes when |
| --- | --- | --- |
| `img_alt` | Every image has non-empty alt text | None missing or blank (findable flags `alt=""` even where valid for decoration) |
| `img_dimensions` | Every image has width and height | None missing |
| `img_alt_quality` | Alt is not a filename or placeholder | No alt like `IMG_1234` or `photo.jpg` |
| `img_alt_length` | Alt is not too long | 125 characters or fewer |
| `img_lazy_loading` | Few images load eagerly | 5 or fewer without `loading="lazy"` or `fetchpriority="high"` |
| `img_modern_format` | Modern file types | At least half are WebP, AVIF or SVG |

When a page has no images, the first two pass and the others are skipped.

## How do you write good alt text?

1. Describe what the image shows or means in this spot.
2. Be short. Stay under 125 characters.
3. Do not start with "image of". The reader already knows.
4. Never use the filename. `IMG_2043.jpg` or `screenshot-3` fails `img_alt_quality`.
5. For truly decorative images an empty `alt=""` is correct WCAG, but findable counts it as missing and `img_alt` will fail. Prefer real text, or move decoration to CSS backgrounds.

```html
<img src="/mug.webp" alt="Blue ceramic mug on a wooden table" width="800" height="600">
```

## How does lazy loading work?

Add `loading="lazy"` to images that are off screen at first. The browser fetches them only when the reader scrolls near. Leave the main image at the top without it, and consider `fetchpriority="high"` for that one. findable does not count a high priority image as eager.

```html
<img src="/hero.webp" alt="Mug collection" width="1200" height="600" fetchpriority="high">
<img src="/detail.webp" alt="Glaze close-up" width="800" height="600" loading="lazy">
```

## Why add width and height?

With both set, the browser knows the shape before the file arrives and holds the space. Without them, text jumps down when the image appears. That jump hurts the reader. The full SEO list is in [every SEO check](/learn/seo-checks).

## What file types pass?

`img_modern_format` looks at the file extension of each recognised image (jpg, png, gif, webp, avif, svg). It passes when WebP, AVIF or SVG make up half or more. Photos usually shrink well when converted to WebP or AVIF.

## Key takeaways

- Every image needs non-empty `alt`, plus `width` and `height`.
- Keep alt under 125 characters and never use a filename.
- Lazy load below-the-fold images, not the main one.
- Prefer WebP, AVIF or SVG.
- The wider list is in [every SEO check](/learn/seo-checks).
