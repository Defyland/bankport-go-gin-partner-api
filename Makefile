GO ?= go
GOFMT ?= gofmt
GOVULNCHECK_VERSION ?= v1.3.0
GOVULNCHECK ?= $(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
REDOCLY ?= npx --yes @redocly/cli@2.31.5

.PHONY: fmt test test-race bench run security docker-build openapi-lint

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

security:
	$(GOVULNCHECK) ./...

docker-build:
	docker build -t bankport-go-gin-partner-api:local .

openapi-lint:
	$(REDOCLY) lint openapi.yaml
