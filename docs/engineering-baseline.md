# BankPort Engineering Baseline

This repository follows the initiative-wide standards below.

## Mandatory outcomes

- product-grade `README.md` with product and engineering sections
- `openapi.yaml` once the HTTP surface exists
- `docs/adr/`, `docs/architecture/`, `docs/events/`, `docs/benchmarks/`, `docs/api/`, `docs/diagrams/`, `docs/runbooks/`, and `docs/security/`
- atomic Conventional Commit history
- GitHub Actions for lint, tests, security, build, coverage, and OpenAPI validation
- observability with structured logs, metrics, traces, request IDs, and readiness endpoints
- documented k6 performance baselines

## BankPort-specific emphasis

- Gin as an explicit framework choice for route groups, middleware chains, JSON binding, and request validation
- partner auth and access control through API products, scopes, and policy checks
- Redis-backed rate limiting and idempotency for financial write operations
- consistent error envelopes, audit logs, request logs, and usage metering
- webhook delivery with retry, replay, signatures, and operator visibility
- fake internal adapters that keep partner APIs decoupled from financial source systems

## Current implementation boundary

This repository now includes the first runnable Gin API slice. It implements
middleware controls, sandbox adapters, events, webhooks, audit logs,
observability, tests, OpenAPI, Docker, CI, and benchmark scripts. PostgreSQL,
Redis, real financial adapters, and durable webhook workers are intentionally
deferred until the contract and invariants are validated.
