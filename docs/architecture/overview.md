# Architecture Overview

BankPort is a modular monolith API. Gin owns transport concerns and explicit
middleware chains. Domain types and repository behavior live under `internal/`
without depending on Gin.

## Request path

1. Request ID and correlation ID are attached.
2. Timeout context and tracing span are created.
3. Structured logging and Prometheus middleware observe the request.
4. API key authentication resolves the partner and developer app.
5. Rate limiting runs per partner and route.
6. Scope middleware authorizes the product capability.
7. Financial write routes enforce idempotency.
8. Handlers call domain repository methods and emit events/audit evidence.

See `docs/middleware-chain.md` for route groups, middleware ordering, and test
evidence.

## Observability boundary

Observability is treated as part of the architecture, not as an afterthought.
The API emits structured request logs, Prometheus metrics, OpenTelemetry spans,
health/readiness responses, dashboard panels, and alert rules. See
[observability.md](observability.md).

## Current runtime

The MVP uses an in-memory seeded repository so the product is runnable without
external services. PostgreSQL and Redis designs are captured in migrations and
docs because they are the correct production backing services for consistency
and TTL state, but introducing them before the API contract is useful would add
debugging and setup cost.

The canonical API binary is `bankport-api`; `cmd/api` remains as a compatibility
entrypoint. `bankportctl` exposes read-only sandbox inspection commands for
developer-app, rate-limit, and usage-report DX.
