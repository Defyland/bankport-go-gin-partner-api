# Architecture Overview

BankPort is a modular monolith API platform using Clean/Hexagonal/Ports and
Adapters boundaries. Gin owns transport concerns and explicit middleware
chains, but it is not the architecture. HTTP handlers parse input, call
application use cases, and map responses. `internal/usecase` orchestrates
application behavior through small ports. `internal/domain` owns business
invariants and does not depend on Gin, database, cache, CLI, or observability
SDKs.

## Request path

1. Request ID and correlation ID are attached.
2. Timeout context and tracing span are created.
3. Structured logging and Prometheus middleware observe the request.
4. API key authentication resolves the partner and developer app.
5. Rate limiting runs per partner and route.
6. Scope middleware authorizes the product capability.
7. Financial write routes enforce idempotency.
8. Handlers call use cases and map domain/application errors to the standard
   HTTP envelope.
9. Use cases call ports for account reads, financial commands, webhooks,
   platform reads, audit evidence, and metrics.

See `docs/middleware-chain.md` for route groups, middleware ordering, and test
evidence.

## Architecture boundary

The dependency direction is:

```text
cmd / HTTP / CLI adapters -> internal/usecase -> internal/domain
secondary adapters -> internal/domain
composition roots -> concrete adapters
```

See:

- `docs/architecture/ports-and-adapters.md`
- `docs/architecture/go-architecture.md`
- `docs/architecture/module-boundaries.md`
- `docs/architecture/dependency-rule.md`
- `docs/architecture/testing-strategy.md`

## Observability boundary

Observability is treated as part of the architecture, not as an afterthought.
The API emits structured request logs, Prometheus metrics, OpenTelemetry spans,
health/readiness responses, dashboard panels, and alert rules. See
[observability.md](observability.md).

## Current runtime

The MVP uses an in-memory seeded adapter so the product is runnable without
external services. PostgreSQL and Redis designs are captured in migrations and
docs because they are the correct production backing services for consistency
and TTL state, but introducing them before the API contract is useful would add
debugging and setup cost. Those future adapters must implement ports consumed
by `internal/usecase`; they must not enter handlers or domain code directly.

The canonical API binary is `bankport-api`; `cmd/api` remains as a compatibility
entrypoint. `bankportctl` exposes read-only sandbox inspection commands for
developer-app, rate-limit, and usage-report DX.
