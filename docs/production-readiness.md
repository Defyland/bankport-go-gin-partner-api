# Production Readiness

BankPort is a production-shaped sandbox. The repository includes the controls a
senior reviewer should expect locally, while naming the pieces that require
real infrastructure before a production claim would be honest.

## Implemented Locally

| Area | Evidence |
| --- | --- |
| Process lifecycle | Graceful shutdown on `SIGINT`/`SIGTERM`; startup validation before listen; startup log with Go version, `GOMAXPROCS`, and CPU count. |
| HTTP hardening | Explicit read-header, read, write, idle, request-body, and request-context timeouts. |
| Health/readiness | `/health/live` and `/health/ready` with dependency checks. |
| Observability | Structured logs, Prometheus metrics, OpenTelemetry spans, Grafana dashboard, Prometheus alerts, runbooks, optional pprof. |
| API platform controls | API-key auth, scopes, per-partner route rate limits, request IDs, correlation IDs, bounded headers, standardized errors. |
| Financial safety | Tenant isolation, idempotency reservation/replay/conflict, timeout non-cacheability, available-balance checks, cumulative refund guard. |
| Contract | OpenAPI 3.1 with examples, auth, idempotency, rate-limit responses, error envelope schema, bounded request shapes. |
| DX | `bankport-api`, `bankportctl`, Makefile targets, Dockerfile, Docker Compose, pinned validation tools, GoReleaser config. |

## Production Gaps

| Gap | Why it is not local-only | Required production move |
| --- | --- | --- |
| Durable persistence adapter | In-memory state cannot survive restart or coordinate multiple API replicas. | Implement PostgreSQL adapter with transaction tests and guarded updates. |
| Distributed rate limits | Fixed-window memory state is process-local. | Use Redis or another shared low-latency store while keeping policy documented. |
| Shared idempotency reservation | Current wait-and-replay is in-process. | Persist idempotency records in PostgreSQL and use Redis for fast shared reservation/cache. |
| Durable webhook replay | Current API queues delivery evidence in memory. | Add outbox table, worker, retry budget, DLQ, replay endpoint, and queue metrics. |
| Token rotation | Seeded API keys come from environment. | Add persistent key records, dual-read rotation, revocation, and audit trail. |
| Trace export | Spans are created locally but not exported to a collector in Compose. | Add OpenTelemetry Collector and exporter configuration. |
| Alert routing | Alert rules exist; Alertmanager routing is not wired locally. | Add Alertmanager configuration and escalation policy. |

## Review Position

It is fair to call the repo senior-ready as an API-platform sandbox. It is not
fair to call it production-ready for money movement until the durable adapters,
distributed coordination, webhook worker, and credential lifecycle are
implemented and tested.
