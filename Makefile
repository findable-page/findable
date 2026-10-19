BIN := findable

.PHONY: lint test build

lint:
	go vet ./...
	golangci-lint run ./...
	cd mcp && bun run lint && bun run typecheck

test:
	go test -race -timeout 10s ./...

build:
	go build -o $(BIN) ./cmd/findable
