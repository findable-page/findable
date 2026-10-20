---
title: "robots.txt and AI Crawlers: GPTBot, ClaudeBot, PerplexityBot"
description: "How robots.txt rules decide whether GPTBot, ClaudeBot, PerplexityBot and Google-Extended can fetch your page, and how findable tests each one."
slug: robots-txt-ai-crawlers
date: 2026-10-15
updated: 2026-10-15
author: "findable team"
faq:
  - q: "Do I need to name AI bots in robots.txt?"
    a: "No. A bot with no group of its own follows the User-agent: * rules. Name a bot only when you want it to behave differently from everyone else."
  - q: "Does blocking GPTBot remove me from ChatGPT?"
    a: "It stops that crawler from fetching your pages. Vendors can run separate bots for training and for search, so read each vendor's bot documentation before you decide."
  - q: "Is robots.txt a security tool?"
    a: "No. It is a request that well-behaved crawlers follow. Anything private needs real access control."
  - q: "Why are these results not in my score?"
    a: "They are Site checks. They need the -site flag or the site option in the API, and they are reported apart from the SEO, AEO and GEO score."
---

# robots.txt and AI crawlers: GPTBot, ClaudeBot, PerplexityBot

Whether an AI tool can read your page starts with robots.txt. findable tests this with its Site checks. They run only when you ask for them, and they do not change your score.

## What does findable test?

| Check id | What it tests |
| --- | --- |
| `robots-txt` | `/robots.txt` exists and returns 200 |
| `robots-page` | The `*` group lets crawlers fetch the page you scanned |
| `robots-gptbot` | GPTBot may fetch the page |
| `robots-claudebot` | ClaudeBot may fetch the page |
| `robots-perplexitybot` | PerplexityBot may fetch the page |
| `robots-google-extended` | Google-Extended may fetch the page |

Run them with `findable -site https://example.com`, or send `"site": true` to `/api/scan`. The check tests the exact path you scanned, so a rule that blocks `/blog/` fails for a blog URL and passes for the home page.

## How does findable read the rules?

It follows the usual matching rules:

1. A bot uses the group that names it. If no group names it, it uses the `*` group.
2. A named group replaces the `*` group completely. Rules are not merged.
3. The longest matching pattern wins. On a tie, Allow wins.
4. `*` matches any characters and a trailing `$` anchors the end of the path.
5. An empty `Disallow:` blocks nothing.

Agent names are matched without case.

## What does a clear policy look like?

This allows everyone, and keeps GPTBot out of one folder:

```text
User-agent: *
Disallow: /admin/

User-agent: GPTBot
Disallow: /private/
```

Note the trap: GPTBot now follows only its own group, so `/admin/` is no longer blocked for it. If you want both, repeat the line inside the GPTBot group.

## How do you block an AI bot completely?

```text
User-agent: ClaudeBot
Disallow: /
```

Do this on purpose. If you want to be cited by AI answers, a block works against you. See [why AI answer engines skip your site](/learn/why-ai-answer-engines-skip-your-site).

## What are the common mistakes?

- A blanket `Disallow: /` under `User-agent: *` left over from a staging site.
- Adding a named group and forgetting it drops your `*` rules.
- Serving a 404, or a redirect that lands on an HTML page. Crawlers follow redirects, so the file counts as present, but an HTML page holds no rules and nothing is restricted.
- Treating robots.txt as a way to hide pages. A blocked URL can still be listed if other sites link to it.

## Key takeaways

- Run the checks with `-site`. They are unscored.
- A named bot group overrides `*` entirely.
- Longest pattern wins, and Allow wins ties.
- Decide each AI bot on purpose, then confirm with a scan.
- Next, see [llms.txt](/learn/llms-txt).
