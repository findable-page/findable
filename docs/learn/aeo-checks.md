---
title: "Every findable AEO Check ID and How to Fix It"
description: "All 35 AEO checks in findable explained, from aeo_faq to aeo_summary. See what each id tests, how loose the test is and how to fix it. AEO is 25% of the score."
slug: aeo-checks
date: 2026-10-14
updated: 2026-10-14
author: "findable team"
faq:
  - q: "What is AEO in findable?"
    a: "AEO means answer engine optimization. The checks test whether your page is easy to lift an answer from: clear questions, short answers, lists, tables, structured data and plain sentences."
  - q: "How much of the score is AEO?"
    a: "AEO is 25% of the findable score. SEO is 50% and GEO is 25%."
  - q: "Is aeo_summary a smart check?"
    a: "No. It only looks for the words summary, key takeaway or conclusion in your text. It cannot judge whether your summary is good, so write a real one."
  - q: "Does passing AEO checks guarantee I get quoted by an AI answer?"
    a: "No. The checks measure structure that helps, not a promise. Nobody outside the search companies can guarantee a citation, and findable does not claim to."
---

# Every AEO check findable runs, and how to fix each one

AEO is about making your page easy to answer from. People ask a question. A search feature or an assistant looks for a clean passage that answers it. These 35 checks test whether your page offers that. Most are simple pattern matches on your HTML or text.

## What does the AEO family test in findable?

AEO is 25% of your findable score. It has 35 checks on answer shape: structured data, FAQ markup, short direct paragraphs, question headings, lists, tables, plain language and clean markup. Every check is a pattern test on your HTML and text, so each one is only a proxy.

| Check id | What it checks | Fix in one line |
| --- | --- | --- |
| `aeo_structured_data` | Any JSON-LD is present | Add a JSON-LD block |
| `aeo_faq` | FAQPage schema type | Add FAQPage markup |
| `aeo_direct_answers` | A paragraph of 30 to 80 words | Answer a question in one tight paragraph |
| `aeo_question_headings` | A heading starts with What, How or Why, or ends with ? | Phrase an h2 as a question |
| `aeo_definition_lists` | A `<dl>` is present | Use a definition list for terms |
| `aeo_numbered_steps` | An `<ol>` is present | Write steps as an ordered list |
| `aeo_tables` | A `<table>` is present | Put comparisons in a table |
| `aeo_lists` | A `<ul>` or `<ol>` is present | Use a list for items |
| `aeo_code_blocks` | `<code>` or `<pre>` is present | Mark code as code |
| `aeo_collapsible` | A `<details>` is present | Use details for extras |
| `aeo_conversational` | "you", "your" or "we" 3 or more times | Address the reader |
| `aeo_reading_level` | Average sentence is 22 words or fewer | Shorten sentences |
| `aeo_short_sentences` | Average sentence is 20 words or fewer | Shorten them more |
| `aeo_summary` | Text has "key takeaway", "summary" or "conclusion" | Add a summary section |
| `aeo_clean_url` | Path is lowercase with no underscores | Use lowercase hyphen URLs |
| `aeo_semantic_html` | 4 or more semantic HTML5 elements | Use header, main, article, nav |
| `aeo_ai_robots` | No `noindex` in the robots meta | Remove `noindex` |
| `aeo_speakable` | Speakable markup | Add `speakable` to Article schema |
| `aeo_answer_after_question` | Question headings are followed by a paragraph | Answer right under the question |
| `aeo_answer_length_ideal` | Answers of 20 to 80 words | Tighten or expand the answer |
| `aeo_faq_count` | 3 or more FAQ questions (N/A without FAQPage) | Add more questions |
| `aeo_howto_valid` | HowTo steps are complete (N/A without HowTo) | Name and describe every step |
| `aeo_speakable_valid` | Speakable selectors are valid (N/A without speakable) | Add cssSelector or xpath |
| `aeo_list_after_heading` | A heading is followed by a list | Lead a section with a list |
| `aeo_list_item_length` | List items average 25 words or fewer | Shorten list items |
| `aeo_paragraph_length` | No paragraph over 150 words | Split long paragraphs |
| `aeo_section_length` | 350 or fewer words between headings | Add subheadings |
| `aeo_unique_headings` | No duplicate heading text | Make each heading distinct |
| `aeo_heading_length` | Headings of 100 characters or fewer | Shorten headings |
| `aeo_heading_anchors` | h2 and h3 headings have an id | Add `id` for deep links |
| `aeo_toc_links` | 3 or more in-page links on long pages | Add a table of contents |
| `aeo_definition_lead` | A section opens with a definition | Start with "X is ..." |
| `aeo_table_captions` | Tables have a caption (N/A without tables) | Add `caption` |
| `aeo_figure_captions` | Figures have a figcaption (N/A without figures) | Add `figcaption` |
| `aeo_details_summary` | Each details has a summary (N/A without details) | Add `summary` |

## How do I fix the structured data and FAQ checks?

Two checks cover markup: `aeo_structured_data` passes with any JSON-LD block, and `aeo_faq` passes only with a FAQPage type. Add JSON-LD that matches content people can see on the page. Do not mark up questions that are not visible. A block that is not valid JSON is dropped, so a trailing comma fails both checks at once.

```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "FAQPage",
  "mainEntity": [{
    "@type": "Question",
    "name": "What is AEO?",
    "acceptedAnswer": {"@type": "Answer", "text": "Making a page easy to answer from."}
  }]
}
</script>
```

See [schema.org FAQPage](https://schema.org/FAQPage) and Google's [structured data intro](https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data). Markup helps machines read the page. It does not force a rich result.

## How do I fix the direct answer and question heading checks?

Two checks cover answer shape: `aeo_direct_answers` wants one paragraph of 30 to 80 words, and `aeo_question_headings` wants a heading that starts with What, How or Why, or ends with a question mark. Put the answer right under the heading. Only `<p>` elements count as paragraphs, so a 50 word answer inside a `<div>` is invisible to the check.

```html
<h2>How do I add a canonical link?</h2>
<p>Put a link tag in the head with rel set to canonical and your preferred URL.
Every duplicate version of the page should point at that same URL. Search engines
then treat that address as the main one.</p>
```

The word range is a guess at a good snippet size. It is not a published rule.

## How do I fix the list, table and code checks?

Six checks reward structure: `aeo_definition_lists`, `aeo_numbered_steps`, `aeo_tables`, `aeo_lists`, `aeo_code_blocks` and `aeo_collapsible`. Each one passes when the element exists. Use the right element for the job and these pass on their own. An `<ol>` satisfies both `aeo_numbered_steps` and `aeo_lists`, so a page with one ordered list and no bullets still passes both.

```html
<ol><li>Open the file.</li><li>Add the tag.</li><li>Save.</li></ol>
<dl><dt>AEO</dt><dd>Answer engine optimization.</dd></dl>
<details><summary>More detail</summary><p>Extra text.</p></details>
```

Do not add an empty table just to pass. These checks only see that the tag exists. They cannot tell if it helps the reader. A pageful of prose with no steps or data may not need all six.

## How do I fix the readability checks?

Three checks cover plain writing: `aeo_conversational`, `aeo_reading_level` and `aeo_short_sentences`. Use "you" or "we" at least 3 times. Keep the average sentence at 22 words or fewer for reading level, and 20 or fewer for the short sentence check. The pronoun count needs a space on each side, so "you," and "you." are not counted.

Cut filler. Split long sentences at "and" or "but". Prefer plain words. The average is computed by splitting text on periods, so a page full of bullet points with no periods looks like one giant sentence and fails. End list items with a period if that bites you.

## How do I fix the summary check?

The `aeo_summary` check is simple. It passes when your text contains the phrase "key takeaway", "summary" or "conclusion". That is all. It does not read your summary or judge it. A page that says "no summary" would pass. The match is case-insensitive and runs over the whole page text, so a nav link called "Summary" passes it too.

Do the real thing anyway. End the page with a "Key takeaways" section of 3 to 6 bullets. Readers who skim land there first. The check is a nudge to write one, not proof that you wrote a good one.

## How do I fix the URL, semantic HTML and robots checks?

Three checks cover housekeeping: `aeo_clean_url` wants a lowercase path with no underscores, `aeo_semantic_html` wants 4 or more semantic HTML5 elements, and `aeo_ai_robots` fails if the robots meta contains `noindex`. Use hyphens, landmark elements and no noindex. The elements counted are article, section, main, nav, aside, header, footer, time, figure and details.

```html
<!-- good path: /learn/aeo-checks  bad path: /Learn/AEO_Checks -->
<header>...</header>
<nav>...</nav>
<main><article>...</article></main>
<footer>...</footer>
```

Despite the name, `aeo_ai_robots` only checks the robots meta tag for `noindex`. It does not read robots.txt, and it ignores `nofollow` and `nosnippet`. For AI crawler rules in robots.txt, run with `-site` and read the unscored Site checks in the GEO guide.

## Key takeaways

- AEO is 25% of the findable score and has 35 checks.
- Most checks only test that an element or phrase exists. They are proxies.
- Put a 30 to 80 word answer right under a question heading.
- Use lists, tables, steps and code blocks where the content calls for them.
- `aeo_summary` only looks for the words summary, key takeaway or conclusion. Write a real one.
- Keep sentences short and speak to the reader as "you".
- A good score helps, but it does not guarantee a citation anywhere.
