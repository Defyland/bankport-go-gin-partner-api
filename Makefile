GO ?= go
GOFMT ?= gofmt

.PHONY: fmt test test-race bench run docker-build openapi-lint

fmt:
	$(GOFMT) -w cmd internal

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

bench:
	$(GO) test -bench=. -benchmem ./internal/httpapi

run:
	$(GO) run ./cmd/api

docker-build:
	docker build -t bankport-go-gin-partner-api:local .

openapi-lint:
	npx @redocly/cli lint openapi.yaml
