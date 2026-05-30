# BankPort Partner API

BankPort is a Go/Gin partner API platform for exposing financial capabilities to
external developers with API-key authentication, scopes, rate limits,
idempotency, event contracts, webhooks, audit logs, and observability.

## 1. What is this product?

BankPort is the public integration boundary between partners and internal
financial systems. The current implementation is a production-shaped sandbox:
the API runs locally, enforces the important controls, emits domain evidence,
and keeps provider adapters fake until the public contract is stable.

## 2. Problem it solves

Financial API partners need safe retries, clear errors, tenant isolation,
webhook evidence, and operational support data. A thin CRUD API is not enough
for money movement because duplicate requests, overbroad credentials, and
contract drift create real financial and support risk.

## 3. Target users

- partner developers integrating account, Pix, payout, refund, and webhook APIs
- partner operations teams investigating failed requests and reconciliation
- platform engineers operating public financial APIs
- security reviewers checking tenant isolation and credential controls

## 4. Main features

- versioned `/v1` API routes
- API-key authentication through Bearer token or `X-API-Key`
- scoped authorization per product capability
- per-partner, per-route fixed-window rate limiting
- TTL-backed idempotency for Pix, payout, and refund writes
- standardized error envelope with request and correlation IDs
- account balance and statement reads
- Pix transfer, payout, and refund commands
- webhook endpoint registration and signed queued delivery evidence
- partner audit logs and deterministic sandbox scenarios

## 5. Architecture overview

The API is a modular monolith. Gin owns routing and middleware composition.
Domain types, repository behavior, metrics, and webhook signing live in focused
internal packages. The sandbox repository is in memory; the production data
model is documented in `db/migrations/001_init.sql`.

## 6. Tech stack

- Go 1.26 compatible module
- Gin HTTP framework
- Prometheus client metrics
- OpenTelemetry tracing API
- slog JSON structured logs
- Docker and Docker Compose
- Grafana dashboard provisioning
- k6 benchmark scripts
- GitHub Actions CI

## 7. Domain model

Core aggregates are Partner, Developer App, Account, Financial Command, Webhook
Endpoint, Event, and Audit Entry. Domain docs live in `docs/domain/`.

## 8. API documentation

The OpenAPI contract is in `openapi.yaml`. Examples and error semantics live in
`docs/api/`.

## 9. Async or event architecture

Accepted financial writes emit CloudEvents-like records such as
`pix.transfer.created.v1`, `payout.created.v1`, and `refund.created.v1`.
Matching webhook endpoints receive queued delivery evidence with timestamped
HMAC signatures. Durable workers and dead-letter queues are documented as next
phase work.

## 10. Database design

`db/migrations/001_init.sql` defines the production PostgreSQL model: partners,
developer apps, API keys, accounts, statements, idempotency keys, financial
commands, webhook endpoints, webhook deliveries, outbox events, and audit
entries. Redis is planned for distributed rate limiting and short-lived
idempotency cache reads.

## 11. Testing strategy

Tests cover:

- domain validation
- API authentication and scopes
- tenant isolation and BOLA prevention
- idempotency replay and conflict handling
- rate-limit failures
- webhook registration and signing
- repository behavior and event queuing
- cumulative refund protection
- idempotency and rate-limit memory cleanup
- native benchmark for the hot read path

```bash
go test ./...
```

## 12. Performance benchmarks

k6 scripts live under `benchmarks/`:

- `k6-smoke.js`
- `k6-load.js`
- `k6-stress.js`
- `k6-spike.js`

Native benchmark:

```bash
go test -bench=. -benchmem ./internal/httpapi
```

Results are stored under `benchmarks/results/`.

## 13. Observability

The API exposes:

- `/health/live`
- `/health/ready`
- `/metrics`
- structured JSON request logs
- request ID and correlation ID headers
- OpenTelemetry spans through middleware
- Prometheus metrics for HTTP, financial commands, rate limits, webhooks, and idempotency
- Grafana dashboard at `deployments/grafana/dashboards/bankport-partner-api.json`

## 14. Security considerations

Security controls include API-key hashing, scoped access, tenant ownership
checks, idempotency request hashes, rate limiting, webhook HMAC signatures,
HTTPS webhook validation, environment-based secrets, and audit entries. See
`docs/security/`.

## 15. Trade-offs and decisions

The project intentionally uses an in-memory sandbox repository so the API can be
reviewed and tested without external services. PostgreSQL, Redis, and webhook
workers are designed but deferred until the contract and invariants are proven.
ADRs live in `docs/adr/`.

For a reviewer-oriented technical assessment, see
`docs/tech-lead-assessment.md`.

For the explicit spec-driven acceptance criteria, implementation mapping, and
verification report, see:

- `docs/spec-driven/senior-readiness-spec.md`
- `docs/spec-driven/implementation-plan.md`
- `docs/spec-driven/verification-report.md`

## 16. How to run locally

With Go:

```bash
go run ./cmd/api
```

With Docker Compose, when the Docker daemon is running:

```bash
docker compose up --build
```

Sandbox keys:

```text
bp_sandbox_full_access_key
bp_sandbox_readonly_key
bp_sandbox_other_partner_key
```

Example:

```bash
curl -H "Authorization: Bearer bp_sandbox_full_access_key" \
  http://localhost:8080/v1/accounts/acct_sandbox_001/balance
```

## 17. How to run tests

```bash
go test ./...
go test -race ./...
go vet ./...
npx @redocly/cli lint openapi.yaml
docker compose config
```

## 18. Failure scenarios

Documented and tested scenarios include invalid credentials, insufficient scope,
foreign account access, missing idempotency key, idempotency conflict, invalid
JSON, insufficient funds, invalid webhook URL, rate-limit spike, and webhook
delivery backlog.

## 19. Roadmap

Next engineering steps:

- implement PostgreSQL repository and Redis-backed distributed limits
- add durable webhook worker with retries and dead-letter handling
- run k6 benchmarks against Docker Compose and store measured p50/p95/p99
- wire OpenTelemetry exporter to a collector
- add partner API-key rotation workflow and replay endpoints

See `docs/engineering-case-study.md` for the full senior engineering narrative.
