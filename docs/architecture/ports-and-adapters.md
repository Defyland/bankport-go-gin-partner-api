# Ports and Adapters

BankPort is not MVC with renamed folders and it is not Gin-driven
architecture. Gin is a delivery adapter at the edge. The application core is
organized around use cases, domain invariants, and ports consumed by those use
cases.

## Architecture Direction

```text
cmd/*, HTTP, CLI
  -> internal/app
  -> internal/httpapi or internal/app/bankportctl
  -> internal/usecase
  -> internal/domain

internal/usecase
  -> ports defined in internal/usecase
  <- adapters: internal/store, internal/webhook, internal/observability
```

Dependencies point inward. The domain does not import Gin, SQL, Redis,
Prometheus, OpenTelemetry, or CLI packages.

## Primary Adapters

| Adapter | Package | Responsibility |
| --- | --- | --- |
| HTTP API | `internal/httpapi` | Route registration, JSON binding, partner context extraction, response mapping, OpenAPI-facing behavior. |
| HTTP middleware pipeline | `internal/httpapi/middleware` | API-platform edge policies: request identity, body limits, timeout context, tracing, logs, metrics, auth, scopes, rate limits, idempotency. |
| API process | `cmd/bankport-api`, `internal/app/bankportapi` | Config validation, dependency composition, HTTP server timeouts, graceful shutdown. |
| Compatibility process | `cmd/api` | Backward-compatible API entrypoint that delegates to the same app runner. |
| CLI | `cmd/bankportctl`, `internal/app/bankportctl` | Local sandbox inspection for apps, rate-limit policy, and usage reports. |

Primary adapters may know delivery protocols and framework details. They should
not decide financial behavior, audit policy, or domain invariants.

## Application Use Cases

`internal/usecase` owns application orchestration:

- account reads
- financial command orchestration
- webhook endpoint registration orchestration
- audit entry creation policy
- financial-command metric outcomes
- webhook queue metric outcomes
- application ports for accounts, commands, webhooks, platform reads, metrics,
  and event signing

The use-case package defines small interfaces where it consumes them. That is
intentional: adapters implement what the use case needs instead of forcing the
application layer to depend on a broad repository abstraction.

## Domain

`internal/domain` owns business language and invariants:

- partners and developer apps
- scopes
- account and statement types
- Pix transfer, payout, refund, webhook endpoint, event, delivery, audit, and
  sandbox scenario types
- financial input validation
- webhook URL and event subscription validation
- domain error taxonomy and `errors.Is` compatibility

Domain types have JSON tags because the sandbox API returns domain-shaped
responses directly. That is acceptable for this challenge because there is no
database/cache model leaking into domain. If the public contract diverges from
domain language, introduce HTTP DTOs in `internal/httpapi`.

## Secondary Adapters

| Adapter | Package | Ports implemented |
| --- | --- | --- |
| In-memory sandbox store | `internal/store` | account reads, financial commands, webhook registration, platform reads, audit writes, CLI read models |
| Webhook signing | `internal/webhook` | event signing function injected into use cases and consumed by store delivery queueing |
| Prometheus metrics | `internal/httpapi.metricsRecorder` + `internal/observability` | use-case metrics port plus HTTP middleware metrics |

Planned adapters:

- PostgreSQL adapter for durable accounts, idempotency records, audit,
  outbox, webhook delivery state, and refund guards.
- Redis adapter for distributed rate-limit windows and fast idempotency
  reservation/cache.
- Webhook HTTP delivery worker with retry/DLQ ports.
- Provider adapters for Pix, payout, and refund rails.

## Why Rate Limit and Idempotency Are Middleware

Rate limiting and idempotency are API-platform edge policies in the current
runtime:

- rate limiting needs route-pattern labels and partner request identity before
  handler execution;
- idempotency must capture/replay HTTP status, headers, and response body;
- unauthorized or insufficient-scope requests must not create idempotency
  records.

Those policies are documented and tested as middleware pipeline behavior. If
BankPort later needs non-HTTP delivery adapters for financial writes, the shared
policy can be extracted behind application ports while keeping the HTTP replay
adapter at the edge.

## Evidence

- `internal/httpapi/router.go` handlers parse, call use cases, and map
  responses/errors.
- `internal/httpapi/router.go` imports `internal/usecase` and does not import
  `internal/store`.
- `internal/usecase/service.go` orchestrates audit, metrics, command results,
  and webhook queue metrics through ports.
- `internal/usecase/service_test.go` tests use cases with fake adapters.
- `internal/domain/models_test.go` tests invariants without Gin, database, or
  cache.
- `internal/store/memory_test.go` tests the in-memory adapter behavior.
