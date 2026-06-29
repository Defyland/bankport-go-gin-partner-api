GO ?= go
GOFMT ?= gofmt
GOVULNCHECK_VERSION ?= v1.3.0
GOVULNCHECK ?= $(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
REDOCLY ?= npx --yes @redocly/cli@2.31.5

.PHONY: fmt fmt-check mod-check vet build test test-race test-race-cover coverage bench run security compose-config docker-build openapi-lint verify ci

fmt:
	$(GOFMT) -w cmd internal

fmt-check:
	test -z "$$($(GOFMT) -l cmd internal)"

mod-check:
	@tmpdir=$$(mktemp -d); \
	cp go.mod "$$tmpdir/go.mod"; \
	cp go.sum "$$tmpdir/go.sum"; \
	$(GO) mod tidy; \
	cmp -s go.mod "$$tmpdir/go.mod" && cmp -s go.sum "$$tmpdir/go.sum" || { \
		diff -u "$$tmpdir/go.mod" go.mod || true; \
		diff -u "$$tmpdir/go.sum" go.sum || true; \
		rm -rf "$$tmpdir"; \
		exit 1; \
	}; \
	rm -rf "$$tmpdir"

vet:
	$(GO) vet ./...

build:
	$(GO) build ./cmd/bankport-api ./cmd/bankportctl

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

test-race-cover:
	$(GO) test -race -coverpkg=./... -coverprofile=coverage.out ./...

coverage:
	$(GO) tool cover -func=coverage.out

bench:
	$(GO) test -bench=. -benchmem ./internal/httpapi

run:
	$(GO) run ./cmd/bankport-api

security:
	$(GOVULNCHECK) ./...

compose-config:
	docker compose config

docker-build:
	docker build -t bankport-go-gin-partner-api:local .

openapi-lint:
	$(REDOCLY) lint openapi.yaml

verify: fmt-check mod-check vet test-race-cover build coverage security openapi-lint

ci: verify compose-config docker-build
