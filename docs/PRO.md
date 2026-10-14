# findable Pro

Free scans stay free. Pro sells watching, not scanning.

## Tiers

- **Free**: one-page scan, share card, "watch this score" signup.
- **Pro, $12/mo**: weekly rescans of up to 5 sites, email when a score drops, whole-site crawl (50 pages), PDF report.
- **Agency, $49/mo**: 25 sites, white-label PDF with their logo, client share links.
- **API**: scans over HTTP/MCP, metered, for people building on findable.

## Order of work

1. Watch list (done): `watchers` table in D1, filled from the form under results.
2. Pick an email sender, add double opt-in and an unsubscribe link. Nothing is emailed until then.
3. Weekly cron rescans each watched host, stores scores, emails on change. Free for watchers at first, to learn what people open.
4. Stripe Checkout + a `plans` table keyed by email. Gate site count and crawl size by plan.
5. PDF report and white-label after 10 paying users ask for it.

## Signals to watch

- Signups per 100 scans. Under 2 means the pitch is wrong, not the price.
- Share-link visits that turn into scans.
- Which hosts get watched by more than one email (agencies).
