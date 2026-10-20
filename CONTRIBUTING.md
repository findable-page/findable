# Contributing

## Setup

- Go 1.26+, [golangci-lint](https://golangci-lint.run) v2
- [Bun](https://bun.sh) for `mcp/` and `worker/`

```bash
git clone https://github.com/findable-page/findable
cd findable
(cd mcp && bun install) && (cd worker && bun install)
```

## The gate

Every change must pass:

```bash
make lint test
```

That runs `go vet`, `golangci-lint run`, `go test -race`, and `biome check` plus `tsc --noEmit` in `mcp/` and `worker/`.

## Rules

- Behaviour changes need a failing test first.
- Scores, CSV bytes and API JSON are a contract: a change to any of them needs a reason in the PR.
- One concern per pull request.
- Changes to the fetch guard (`internal/fetch/guard.go`, `worker/src/guard.ts`) must keep the policy in SECURITY.md true and add a test for each new blocked case.

## Deploying static assets

`worker/public/` holds `favicon.svg`, `apple-touch-icon.png` and `og.png`, and nothing else; `worker/test/public.test.ts` fails when the set drifts from what the UI references.
`wrangler deploy` uploads the directory through the `assets` setting. A deploy through the Cloudflare API does not, so upload the files in `worker/public/` as assets yourself.
Keep the copies in `internal/server/web/` in step by hand.
