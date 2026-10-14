# Contributing

## Setup

- Go 1.26+, [golangci-lint](https://golangci-lint.run) v2
- [Bun](https://bun.sh) for `mcp/` and `worker/`

```bash
git clone https://github.com/nadimtuhin/findable
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
