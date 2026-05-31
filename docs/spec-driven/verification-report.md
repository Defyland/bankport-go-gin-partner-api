# Verification Report

## Summary

Verification was rerun on 2026-05-31 after the tech-lead hardening pass and the
Clean/Hexagonal/Ports and Adapters refactor. The latest pass covers
idempotency concurrency, timeout non-cacheability, strict environment parsing,
production config validation, request body limits, context cancellation,
bounded identity/idempotency headers, low-cardinality route labels, unique
fallback request IDs, non-panicking sandbox ID generation, endpoint-specific
webhook signatures, webhook event allowlisting, financial input shape checks,
pinned validation tooling, canonical `bankport-api` runtime, optional pprof,
`bankportctl` sandbox CLI, GoReleaser release shape, middleware-chain docs,
`internal/usecase` orchestration behind small ports, and static dependency-rule
evidence.

The current state passes Go tests, race/coverage, vet, vulnerability scan,
OpenAPI lint, Docker Compose config validation, Docker image build, Compose API
and Prometheus runtime health checks, binary build, CLI smoke checks,
GoReleaser YAML syntax validation, and the architecture boundary grep.

## Commands Run

| Command | Result |
| --- | --- |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto go env GOVERSION GOTOOLCHAIN` | Passed, `go1.26.3`, `auto` |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto gofmt -l cmd internal` | Passed, no files listed |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto go mod tidy -diff` | Passed, no diff |
| `ruby -e 'require "yaml"; YAML.load_file(".goreleaser.yaml")'` | Passed, GoReleaser YAML syntax is valid |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto go test ./...` | Passed |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto go test -race -coverpkg=./... -coverprofile=/private/tmp/bankport-coverage.out ./...` | Passed, coverage profile generated |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto go tool cover -func=/private/tmp/bankport-coverage.out` | Passed, total coverage `78.3%` |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto go vet ./...` | Passed |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...` | Passed, 0 called vulnerabilities |
| `npx --yes @redocly/cli@2.31.5 lint openapi.yaml` | Passed |
| `docker compose config` | Passed |
| `docker build -t bankport-go-gin-partner-api:validation .` | Passed, including Dockerfile test stage |
| `docker compose up -d --build api prometheus` | Passed, started API and Prometheus |
| `curl -fsS http://127.0.0.1:8080/health/live` | Passed, API live |
| `curl -fsS http://127.0.0.1:8080/health/ready` | Passed, API ready with dependency checks |
| `curl -fsS http://127.0.0.1:8080/v1/accounts/acct_sandbox_001/balance -H 'Authorization: Bearer bp_sandbox_full_access_key'` | Passed, authenticated sandbox balance returned |
| `curl -fsS http://127.0.0.1:9090/-/ready` | Passed, Prometheus ready |
| `docker compose down` | Passed, removed local validation containers and network |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto go test -bench=. -benchmem ./internal/httpapi` | Passed, `BenchmarkGetBalanceRequest-10 102674 10853 ns/op 11607 B/op 96 allocs/op` |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto go build -trimpath -o /private/tmp/bankport-api ./cmd/bankport-api` | Passed, produced 21 MB binary |
| `ASDF_GOLANG_VERSION=1.25.10 GOTOOLCHAIN=auto go build -trimpath -o /private/tmp/bankportctl ./cmd/bankportctl` | Passed, produced 3.3 MB binary |
| `/private/tmp/bankportctl apps list` | Passed, listed three sandbox developer apps |
| `/private/tmp/bankportctl rate-limits inspect` | Passed, listed fixed-window sandbox policies |
| `/private/tmp/bankportctl usage report --format json` | Passed, returned sandbox usage counts |
| `ASDF_RIPGREP_VERSION=14.1.1 rg -n "github.com/gin-gonic/gin|gin\\.Context|gorm|redis|sql|httpapi|middleware" internal/domain internal/usecase internal/store` | Passed, no matches |

## Passing Criteria

- Required docs structure exists, including `docs/spec-driven/`.
- README points to the case study, assessment, and spec-driven evidence docs.
- API contract is valid OpenAPI.
- Clean/Hexagonal/Ports and Adapters docs are explicit:
  `docs/architecture/ports-and-adapters.md`,
  `docs/architecture/go-architecture.md`,
  `docs/architecture/module-boundaries.md`,
  `docs/architecture/dependency-rule.md`, and
  `docs/architecture/testing-strategy.md`.
- `internal/httpapi` owns Gin routing and response mapping; application
  orchestration sits in `internal/usecase`.
- `internal/usecase` defines consumed ports and tests orchestration with fake
  adapters.
- Canonical API binary is `cmd/bankport-api`; `cmd/api` remains compatibility
  entrypoint.
- `bankportctl` provides read-only sandbox app, rate-limit, and usage-report
  inspection without external services.
- Auth, scopes, tenant isolation, idempotency replay/conflict/concurrency,
  idempotency key validation, timeout non-cacheability, rate limiting,
  webhooks, webhook event allowlisting, audit, cumulative refund protection,
  financial input shape checks, request body limits, strict config parsing,
  production config validation, context cancellation, non-panicking sandbox ID
  generation, unique fallback request IDs, and cleanup behavior have automated
  tests.
- Observability includes logs, metrics, traces, low-cardinality route labels,
  dashboard provisioning, alert rules, optional pprof, and runbooks.
- CI covers format, tests, race/coverage, pinned security scan, pinned OpenAPI
  lint, binary build, Compose validation, and Docker build validation.
- Local Docker runtime evidence covers API readiness, authenticated business
  endpoint behavior, Prometheus scraping, and metrics exposure.
- Static dependency evidence shows no Gin, HTTP adapter, store, SQL, or Redis
  dependency leaking into `internal/domain` or `internal/usecase`.

## Partial Criteria

- Production persistence is planned and specified, but runtime still uses an
  in-memory sandbox adapter.
- Redis-backed distributed rate limiting is planned; current limiter is
  process-local.
- Shared multi-instance idempotency reservation is planned; current
  wait-and-replay guarantee is in-process.
- Durable webhook worker and DLQ are planned; current API queues delivery
  evidence in memory.
- `bankportctl webhooks replay` and `bankportctl tokens rotate` are intentionally
  deferred until durable webhook delivery and API-key storage exist.
- Trace exporter and Alertmanager routing are planned; local Compose includes
  instrumentation and alert rules but not those external integrations.
- GoReleaser binary was not installed locally; `.goreleaser.yaml` syntax was
  validated, but schema validation should run in CI once GoReleaser is added to
  the release workflow.

## Failed or Blocked Criteria

- None in this verification pass.

## Remaining Risk

The repository is senior-ready as a production-shaped sandbox. The largest
remaining production-readiness risk is replacing in-memory state with durable
PostgreSQL/Redis adapters behind the current ports while preserving transaction
semantics, shared idempotency reservation, cumulative refund protection,
webhook delivery durability, and the public API contract.
