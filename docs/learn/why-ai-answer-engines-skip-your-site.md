---
title: "Why AI Answer Engines Skip Your Site and How to Fix It"
description: "ChatGPT, Perplexity and Claude may ignore your pages for five fixable reasons, from robots.txt blocks to missing answers. Learn each cause and its check id."
slug: why-ai-answer-engines-skip-your-site
date: 2026-10-14
updated: 2026-10-14
author: "Nadim Tuhin"
faq:
  - q: "Does blocking GPTBot hide my site from ChatGPT?"
    a: "It stops the GPTBot crawler from fetching your pages. OpenAI documents separate bots for training and for search, so read their bot page and decide each one on purpose."
  - q: "Is llms.txt required?"
    a: "No. It is a proposal, not a standard that every engine reads. It is cheap to add, and findable reports it as the llms-txt site check."
  - q: "How do I see which bots my robots.txt blocks?"
    a: "Run findable with -site. It reports robots-gptbot, robots-claudebot, robots-perplexitybot and robots-google-extended as separate results."
---

# Why AI Answer Engines Skip Your Site and How to Fix It

You wrote the page. It ranks fine. Then you ask ChatGPT, Perplexity or Claude about your topic and your site never shows up. This post lists the five most common causes and the findable check that catches each one.

## Why does robots.txt keep AI bots out?

It is easy to block AI crawlers without noticing. A broad `Disallow` rule, or a rule copied from a template, tells bots like GPTBot, ClaudeBot and PerplexityBot to stay away. If a bot cannot fetch your page, it cannot use it.

Check your file at `/robots.txt`. The [Google robots.txt intro](https://developers.google.com/search/docs/crawling-indexing/robots/intro) explains the format. Each bot has its own docs: [OpenAI crawlers](https://platform.openai.com/docs/bots), [ClaudeBot](https://support.anthropic.com/en/articles/8896518-does-anthropic-crawl-data-from-the-web-and-how-can-site-owners-block-the-crawler), [PerplexityBot](https://docs.perplexity.ai/guides/bots) and [Google-Extended](https://developers.google.com/search/docs/crawling-indexing/google-common-crawlers).

This snippet allows all four bots:

```text
User-agent: GPTBot
Allow: /

User-agent: ClaudeBot
Allow: /

User-agent: PerplexityBot
Allow: /

User-agent: Google-Extended
Allow: /

User-agent: *
Allow: /
```

If you do want to block one of them, that is your call. Just make it a choice. With `-site`, findable reports `robots-txt`, `robots-page`, `robots-gptbot`, `robots-claudebot`, `robots-perplexitybot` and `robots-google-extended`.

## Why does a missing llms.txt matter?

An `llms.txt` file is a plain Markdown file at your site root that points language models to your best pages. It is a proposal, so no engine promises to read it. It costs little to add and gives tools a clean map of your site.

Read the [llms.txt proposal](https://llmstxt.org/) for the format. findable checks for it with the `llms-txt` site check. It is unscored, so a miss will not drop your grade. The check only tests that the file exists; it does not parse the Markdown or follow the links in it.

## Why can't bots read JavaScript-only content?

If your text only appears after a script runs, some crawlers will see an empty shell. Google can render JavaScript, but you should not assume every AI bot does. Plain HTML in the first server response is the safe path for every reader.

Read [Google's JavaScript SEO basics](https://developers.google.com/search/docs/crawling-indexing/javascript/javascript-seo-basics) for how rendering works. In findable, a thin shell shows up as a failing `word_count` (300+ words), a low `text_html_ratio` (15% or more) and a missing `h1_exists`. If those fail on a page that looks full in your browser, you are probably rendering on the client. Move the key text into server-rendered HTML.

## Why does missing structured data hurt?

Structured data tells a machine what a block of your page is: an article, an FAQ, a person, an organization. Without it, a model has to guess what you meant. With it, your meaning is explicit, and the machine does not need to infer it.

The [structured data intro](https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data) shows how to add JSON-LD, and [schema.org](https://schema.org/) lists the types. findable checks `jsonld`, `aeo_structured_data`, `aeo_faq`, `geo_org_schema`, `geo_author_schema`, `geo_knowledge_graph` and `geo_entity_markup`. Start with one JSON-LD block, then add FAQPage and Organization.

## Why don't engines find a direct answer on your page?

A short passage that stands alone is the easiest thing to quote. If your page buries the point under a long intro, there is nothing clean to lift. Put the answer first, right under a question heading, and add detail after it.

Write a paragraph of 30 to 80 words for `aeo_direct_answers`, and a question heading for `aeo_question_headings`. Add `aeo_lists`, `aeo_tables` and `aeo_numbered_steps` where they fit. For citability, `geo_passage_citability` wants a paragraph of 40 to 120 words, and `geo_citations` wants at least one external link.

## How do you find all five problems at once?

Run one command against your home page with `-site`. You get the page checks plus the robots.txt and llms.txt results in one report. Each bot gets its own row, so you can see that GPTBot is allowed while ClaudeBot is blocked.

```bash
findable -site https://example.com
```

Fix the robots.txt rules first, since a blocked bot ignores everything else. Then server-render your text, add JSON-LD, and rewrite your top sections as question plus answer.

## Key takeaways

- Bots skip sites for five common reasons, and all five are fixable.
- Check robots.txt first. Allow GPTBot, ClaudeBot, PerplexityBot and Google-Extended if you want to be found.
- Add `/llms.txt`, put your text in server-rendered HTML, and add JSON-LD.
- Lead each section with a direct answer.
- Run `findable -site <url>` to see every one of these as a check id.
