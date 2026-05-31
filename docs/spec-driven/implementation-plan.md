# Implementation Plan

## Scope

Apply the spec-driven senior quality standard to BankPort without editing files
outside `bankport-go-gin-partner-api/` except for reading shared specs. The work
focuses on making current behavior, documentation, tests, and verification
match each other.

## Files to Create or Update

| Area | Files |
| --- | --- |
| Spec-driven docs | `docs/spec-driven/senior-readiness-spec.md`, `docs/spec-driven/implementation-plan.md`, `docs/spec-driven/verification-report.md` |
| Technical assessment | `docs/tech-lead-assessment.md`, `README.md` |
| Domain correctness | `internal/store/memory.go`, `internal/store/memory_test.go`, `db/migrations/001_init.sql`, `docs/domain/invariants.md`, `docs/scalability.md`, `docs/engineering-case-study.md` |
| Operational safeguards | `cmd/api/main.go`, `internal/config/config.go`, `internal/config/config_test.go`, `internal/httpapi/router.go`, `internal/httpapi/router_test.go`, `internal/httpapi/middleware/idempotency.go`, `internal/httpapi/middleware/idempotency_test.go`, `internal/httpapi/middleware/rate_limit.go`, `internal/httpapi/middleware/rate_limit_test.go`, `internal/httpapi/middleware/request.go`, `internal/httpapi/middleware/context.go` |
| Observability | `docs/architecture/observability.md`, `deployments/prometheus/alerts.yml`, `deployments/prometheus/prometheus.yml`, `docker-compose.yml` |
| Webhook signing | `internal/webhook/signer.go`, `internal/webhook/signer_test.go`, `internal/store/memory.go`, `internal/store/memory_test.go` |
| API contract | `openapi.yaml`, `docs/api/examples.md`, `docs/api/error-format.md` |

## Acceptance Criteria Mapping

| Acceptance criterion | Change |
| --- | --- |
| Spec-driven docs exist and identify Done, Partial, Planned, and risks. | Add `docs/spec-driven/*`. |
| Reviewer can see what is senior-level and what remains production hardening. | Add `docs/tech-lead-assessment.md` and README links. |
| Money movement has a real senior-level invariant beyond CRUD. | Enforce cumulative refund ceiling and document guarded SQL update. |
| Idempotency state does not grow forever. | Add configurable TTL and cleanup test. |
| Rate-limit state does not grow forever. | Add expired-window pruning and cleanup test. |
| Observability includes alerting evidence, not only dashboard screenshots. | Add Prometheus alert rules and observability doc. |
| Concurrent requests with the same idempotency key cannot execute the financial handler twice. | Add in-flight idempotency reservation, wait-and-replay behavior, and a concurrent middleware regression test. |
| Production mode fails closed when secrets or sandbox API keys are still defaults. | Add `Config.Validate`, startup validation, and production config tests. |
| Oversized JSON bodies cannot exhaust memory before validation. | Add `MAX_REQUEST_BODY_BYTES`, `http.MaxBytesReader`, 413 error mapping, and API test coverage. |
| Request cancellation is honored before financial mutations. | Check request context before repository reads/writes and test canceled Pix transfer behavior. |
| Observability avoids path-parameter cardinality. | Use route patterns for metrics/traces and test account IDs are not emitted as route labels. |
| Webhook signatures are endpoint-specific. | Derive endpoint signing material from the root signing key and endpoint secret ID; add regression coverage. |
| Verification is auditable. | Record exact commands and results in `verification-report.md`. |

## Verification Commands

Run locally with the Go runtime available in this environment:

```bash
/tmp/codex-go1.26.3/bin/gofmt -w cmd internal
/tmp/codex-go1.26.3/bin/go test ./...
/tmp/codex-go1.26.3/bin/go test -race -coverpkg=./... -coverprofile=coverage.out ./...
/tmp/codex-go1.26.3/bin/go vet ./...
/tmp/codex-go1.26.3/bin/go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
npx --yes @redocly/cli@2.31.5 lint openapi.yaml
docker compose config
/tmp/codex-go1.26.3/bin/go test -bench=. -benchmem ./internal/httpapi
/tmp/codex-go1.26.3/bin/go build -trimpath -o /tmp/bankport-partner-api ./cmd/api
```

## Risks

- Docker daemon may be unavailable locally; CI still validates Docker build.
- The in-memory repository proves contract and invariants but is not durable.
- In-process idempotency single-flight prevents duplicate execution in this
  sandbox; multi-instance production deployments must move the reservation to a
  shared transactional store.
- Prometheus alerts are present, but Alertmanager routing is not included.
- Webhook delivery evidence is queued in memory; durable retry/DLQ worker is
  planned.

## Deferred Work

- PostgreSQL repository adapter and migrations runner.
- Redis-backed distributed rate limits and shared idempotency reservation/cache.
- Durable webhook worker with retry queue, DLQ, replay endpoint, and queue-depth
  metric.
- OpenTelemetry collector/exporter wiring in Compose.
- OAuth client credentials and mTLS ingress hardening.
