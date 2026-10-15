---
title: "FAQ Schema for Answer Engines: Markup and Checks"
description: "How to add FAQPage JSON-LD that answer engines can lift, and how findable tests aeo_faq, aeo_faq_count and faq_valid. Includes a working example."
slug: faq-schema-answer-engines
date: 2026-10-15
updated: 2026-10-15
author: "Nadim Tuhin"
faq:
  - q: "Do I need FAQ markup if I have a visible Q&A section?"
    a: "No. The aeo_faq check passes only on FAQPage schema, so a visible Q&A section alone does not pass it. Add the markup, and keep the same questions visible on the page."
  - q: "How many questions should I include?"
    a: "findable recommends 3 or more in the aeo_faq_count check. Add as many as you genuinely answer on the page."
  - q: "Can the markup differ from the visible text?"
    a: "It should not. Mark up the same questions and answers a reader can see. Hidden or invented answers are a trust problem."
  - q: "Should answers be long?"
    a: "Keep them short and direct. A few sentences that answer the question on their own work best."
---

# FAQ schema for answer engines: markup and checks

An answer engine wants a question and a clean answer it can lift. FAQ markup hands it both in a form it does not have to guess at.

## Which checks cover FAQ content?

| Check id | What it tests | Fix |
| --- | --- | --- |
| `aeo_faq` | FAQPage schema is detected | Add FAQPage markup |
| `aeo_faq_count` | The FAQPage lists 3 or more questions | Add more real questions |
| `faq_valid` | Each question has a name and answer text | Fill both fields |
| `aeo_question_headings` | A heading starts with What, How or Why, or ends with ? | Phrase an h2 as a question |
| `aeo_answer_after_question` | A paragraph directly follows each question heading | Put the answer right under the heading |
| `aeo_answer_length_ideal` | Half or more of the answers after question headings are 20 to 80 words | Tighten or expand |

`aeo_faq_count`, `faq_valid` and the answer checks are skipped when there is nothing to test, so they only apply once you add FAQ content.

## What does the markup look like?

```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "FAQPage",
  "mainEntity": [
    {
      "@type": "Question",
      "name": "How long does shipping take?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Orders ship within two working days. Delivery then takes three to seven working days inside the country, and you get a tracking link by email."
      }
    },
    {
      "@type": "Question",
      "name": "Can I return a mug?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Yes. Send it back within 30 days in its original box and we will refund the full price once it arrives at our warehouse."
      }
    },
    {
      "@type": "Question",
      "name": "Are the mugs dishwasher safe?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Yes. Use the top rack of the dishwasher and a gentle cycle, and the glaze will last for many years."
      }
    }
  ]
}
</script>
```

## How should the visible page match?

Show the same questions as headings and put each answer in the paragraph below. Then the schema and the page agree, and the page passes the question heading checks too:

```html
<h2>How long does shipping take?</h2>
<p>Orders ship within two working days. Delivery then takes three to seven working days inside the country, and you get a tracking link by email.</p>
```

## What are the common mistakes?

1. Marking up questions nobody asked. Use real ones.
2. Putting a sales pitch in the answer.
3. Leaving `text` empty, which fails `faq_valid`.
4. One huge answer. Split it into separate questions.
5. Schema that does not match the page text.

## Key takeaways

- Add FAQPage JSON-LD with 3 or more questions you really answer.
- Give every question a `name` and an answer `text`.
- Mirror the markup with question headings and short answers on the page.
- More on the format in [JSON-LD basics](/learn/json-ld-schema-basics) and [every AEO check](/learn/aeo-checks).
