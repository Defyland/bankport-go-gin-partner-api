# Senior Readiness Spec

This spec applies `specs/general-project-spec.md`,
`specs/senior-engineering-rubric.md`, and
`specs/spec-driven-senior-quality.md` to BankPort. It is intentionally honest:
the project is portfolio senior-ready as a production-shaped sandbox, while
durable PostgreSQL, Redis, webhook workers, and real financial adapters remain
planned production hardening.

## Product Bar

BankPort must read like a real partner API product: named users, a financial
integration problem, a concrete workflow, non-goals, business value, and a
roadmap that distinguishes sandbox capability from production adapters.

## Domain Bar

The repository must model partner access, account exposure, financial commands,
webhooks, events, and audit evidence with shared language across docs, code, and
tests. High-risk rules must be explicit: tenant isolation, scoped access,
available balance, idempotency, webhook HTTPS, and cumulative refund limits.

## Architecture Bar

The architecture must justify a modular monolith in Go/Gin, document boundaries,
sequence flows, deployment view, observability view, rejected alternatives, and
why microservices, Kubernetes, and durable queues are deferred.

## API Bar

The API must expose versioned routes, OpenAPI docs, documented authentication,
scope behavior, idempotency rules, request/response examples, and standardized
errors for validation, authorization, rate limits, idempotency conflicts, and
domain failures.

## Data and Consistency Bar

The docs and migration must explain transaction boundaries, unique constraints,
indexes, foreign keys, optimistic locking assumptions, idempotency retention,
rollback strategy, and the guarded cumulative refund update required for
production correctness.

## Security Bar

The system must document and test API-key strategy, scope authorization, BOLA
prevention, tenant isolation, input validation, idempotency abuse, rate-limit
abuse, webhook signing, secret management, audit logging, and residual risks.

## Observability Bar

The runtime must expose structured logs, request IDs, correlation IDs, metrics,
traces, health, readiness, Prometheus metrics, Grafana dashboard, alert rules,
and runbooks that connect symptoms to partner impact.

## Performance Bar

The repository must include native Go and k6 benchmarks with dataset, p50/p95/p99
or honest measurement scope, throughput, error rate, resource notes, bottleneck
hypothesis, and next optimization.

## Scalability Bar

The project must name hot paths, read-heavy routes, write-heavy flows,
fastest-growing tables, queue buildup risk, hot partitions, horizontal scale
boundaries, async candidates, and flows that must not be eventual.

## Operational Cost Bar

The docs must name infrastructure components, non-financial operational cost,
debugging complexity, deployment complexity, backup/retention needs, monitoring
burden, vendor lock-in, and simpler alternatives rejected.

## Maintainability Bar

Module boundaries, scripts, seed data, error codes, test strategy, extension
points, and planned production adapters must be easy to find and consistent with
the implementation.

## Readability Bar

Code, tests, and docs must use BankPort domain language: partner, developer app,
account, Pix transfer, payout, refund, webhook delivery, idempotency key,
correlation ID, and audit entry.

## Test and CI Bar

The project must include unit tests, API/request tests, authorization tests,
security/failure tests, messaging or webhook tests where async behavior exists,
native benchmark, formatting, vet, race/coverage, govulncheck, OpenAPI lint,
Compose validation, and Docker build validation in CI.

## Evidence Matrix

| Criterion | Evidence | Status | Notes |
| --- | --- | --- | --- |
| Product problem, users, workflow, non-goals, and business value are explicit. | `README.md`, `docs/product/problem.md`, `docs/product/personas.md`, `docs/product/use-cases.md`, `docs/product/non-goals.md` | Done | README reads as product and engineering entrypoint. |
| Central case study explains product, domain, architecture, failure, security, scale, cost, and maintainability. | `docs/engineering-case-study.md` | Done | Uses required rubric table of contents. |
| Spec-driven acceptance criteria exist. | `docs/spec-driven/senior-readiness-spec.md` | Done | This file maps gates to evidence. |
| Implementation plan maps changes to acceptance criteria. | `docs/spec-driven/implementation-plan.md` | Done | Plan records scope and verification commands. |
| Verification report records actual commands and results. | `docs/spec-driven/verification-report.md` | Done | Updated after local validation. |
| Domain model is explicit. | `docs/domain/glossary.md`, `docs/domain/bounded-contexts.md`, `docs/domain/aggregates.md`, `docs/domain/invariants.md`, `docs/domain/state-machines.md` | Done | Invariants link to tests. |
| Cumulative refund rule is enforced. | `internal/store/memory.go`, `internal/store/memory_test.go`, `db/migrations/001_init.sql` | Done | Prevents partial refunds from exceeding original Pix transfer amount. |
| API is versioned, authenticated, scoped, and documented. | `openapi.yaml`, `docs/api/examples.md`, `docs/api/error-format.md`, `internal/httpapi/router.go` | Done | Financial writes require idempotency keys. |
| Data model names constraints, indexes, and transaction assumptions. | `db/migrations/001_init.sql`, `docs/scalability.md`, `docs/engineering-case-study.md` | Done | Production adapters still planned, but schema is concrete. |
| Tenant isolation and BOLA prevention are tested. | `TestTenantIsolationHidesForeignAccount`, `TestCreatePixTransferDebitsOnlyPartnerOwnedAccount` | Done | Cross-partner account access returns not found. |
| API-key scope enforcement is tested. | `TestRejectsInsufficientScope` | Done | Read-only key cannot write Pix transfer. |
| API-key hashes are protected by a secret pepper. | `internal/store/memory.go`, `docs/security/secrets.md` | Done | Sandbox uses HMAC-SHA256 with `API_KEY_HASH_PEPPER`; production rotation path is documented. |
| Idempotency replay, conflict, TTL cleanup, and same-key concurrency are tested. | `TestIdempotentFinancialWriteReplaysCachedResponse`, `TestIdempotencyConflict`, `TestIdempotencyStoreExpiresRecords`, `TestIdempotencyConcurrentSameKeyRunsHandlerOnce` | Done | In-process financial writes reserve the idempotency key before handler execution, so duplicate concurrent requests wait and replay instead of double-executing. |
| Rate limiting and cleanup are tested. | `TestRateLimitExceeded`, `TestRateLimiterPrunesExpiredWindows` | Done | Current implementation is in-memory fixed window; Redis is planned. |
| Production configuration fails closed for unsafe defaults. | `Config.Validate`, `TestValidateRejectsProductionDefaults`, `TestValidateAcceptsProductionSecretsAndKeys`, `cmd/api/main.go` | Done | Production mode rejects default peppers, signing keys, and sandbox API keys before the server starts. |
| Request body size is bounded. | `RequestBodyLimit`, `TestRejectsOversizedFinancialBody`, `MAX_REQUEST_BODY_BYTES` | Done | Oversized financial writes return 413 before JSON binding or domain mutation. |
| Request cancellation is honored before mutation. | `internal/store/memory.go`, `TestCreatePixTransferHonorsCanceledContextBeforeMutation` | Done | Repository methods check context before reads/writes and before financial state changes. |
| Webhook signing and delivery evidence are tested. | `TestWebhookRegistrationAndDeliveryQueue`, `TestSignerCreatesVersionedHMACSignature`, `TestSignerDerivesEndpointSpecificSignatures` | Done | Signatures derive endpoint-specific material from the root signing key; durable worker and DLQ are planned. |
| Observability has metrics, logs, traces, dashboard, alerts, and runbooks. | `internal/observability/metrics.go`, `docs/architecture/observability.md`, `deployments/grafana/dashboards/bankport-partner-api.json`, `deployments/prometheus/alerts.yml`, `docs/runbooks/`, `TestMetricsUseRoutePatternForAccountIDs` | Done | Metrics and traces use route patterns to avoid account-id cardinality; trace exporter and Alertmanager routing are planned. |
| Benchmarks include method and measured result. | `benchmarks/baseline.md`, `benchmarks/results/2026-05-30-go-benchmark.txt`, `benchmarks/k6-*.js`, `docs/spec-driven/verification-report.md` | Done | Native benchmark and Docker runtime validation are recorded; k6 scripts remain available for load profiles. |
| CI covers format, tests, coverage, security, OpenAPI, Compose, and Docker build. | `.github/workflows/ci.yml` | Done | Docker build validation is in CI. |
| Production persistence is implemented. | `db/migrations/001_init.sql`, `docs/product/roadmap.md` | Planned | PostgreSQL repository is intentionally deferred; current runtime is sandbox in memory. |
| Distributed rate limiting is implemented. | `docs/architecture/deployment-view.md`, `docs/scalability.md` | Planned | Redis adapter is planned after API contract stabilization. |
| Durable webhook worker with retries and DLQ is implemented. | `docs/runbooks/webhook-delivery-backlog.md`, `docs/product/roadmap.md` | Planned | API queues delivery evidence; worker is next production phase. |

## Out of Scope

- Real Pix, payout, or banking provider integration.
- PostgreSQL repository implementation.
- Redis-backed distributed rate limiter and shared idempotency reservation/cache.
- Durable webhook worker, retry queues, and dead-letter queues.
- OAuth client credentials and mTLS enforcement.
- Kubernetes manifests and multi-region active-active writes.
