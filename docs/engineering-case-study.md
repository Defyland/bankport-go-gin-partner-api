# Engineering Case Study

## 1. Product Context

BankPort is a partner API for financial integrations. It targets external
developers and operations teams that need stable account reads, safe financial
writes, webhook evidence, and auditability without depending on internal banking
schemas.

## 2. Domain Model

The core aggregates are Partner, Account, Financial Command, Webhook Endpoint,
Event, and Audit Entry. The most important invariant is tenant isolation:
authenticated partners can only read and mutate their own accounts.

## 3. Architecture

The system is a Go modular monolith using Gin for versioned routes and explicit
middleware chains. Domain types stay framework-independent under `internal/`.
The current repository is an in-memory sandbox; production persistence is
designed around PostgreSQL and Redis.

## 4. Key Trade-offs

The MVP chooses a runnable, testable sandbox over a half-integrated database
stack. This lowers local setup cost while preserving production decisions in
migrations, ADRs, tests, and docs.

## 5. Data Model

The migration defines partners, developer apps, API keys, accounts, statement
entries, idempotency keys, financial commands, webhook endpoints, webhook
deliveries, outbox events, and audit entries. Hot tables have partner and time
indexes.

## 6. Consistency Model

Financial writes need strong consistency for account ownership, available
balance, idempotency conflict detection, event append, and audit evidence. In
production these belong in one PostgreSQL transaction plus Redis-assisted
rate-limit checks. Refund acceptance must guard the cumulative refunded amount
on the original transaction, not just validate each refund request in isolation.

## 7. Failure Scenarios

Covered failures include missing auth, insufficient scope, tenant isolation,
idempotency conflict, rate-limit rejection, invalid webhook URL, insufficient
funds, original transaction lookup failure, and cumulative refund attempts above
the original transfer amount.

## 8. Performance Strategy

Gin routing is not expected to be the bottleneck. The real bottleneck is
financial write serialization around account balance and idempotency records.
k6 scripts cover smoke, load, stress, and spike profiles.

## 9. Scalability Strategy

Read-heavy endpoints scale horizontally. Write-heavy endpoints need partitioning
by partner or account once a small number of accounts become hot. Webhook
workers should scale independently after durable queues are introduced.

## 10. Security Model

Security controls include API-key hashing, scoped authorization, tenant
ownership checks, idempotency keys, HTTPS webhook validation, HMAC webhook
signatures, structured audit logs, and secret management through environment
variables.

## 11. Observability

The API emits structured JSON logs, request IDs, correlation IDs, Prometheus
HTTP/domain metrics, and OpenTelemetry spans. Grafana panels focus on throughput,
p95 latency, financial commands, rate limits, and idempotency conflicts.

## 12. Operational Cost

The current runtime has one process. Production adds PostgreSQL, Redis, worker
queues, dashboards, backups, and secret rotation. Those costs are deferred until
the partner API contract is proven.

## 13. Maintainability

Module boundaries are explicit: config, domain, store, middleware, HTTP API,
observability, and webhook signing. Tests read like business controls instead
of implementation details.

## 14. Product Decisions

The API optimizes for partner trust: predictable errors, retries that do not
double-charge, correlation IDs for support, and deterministic sandbox scenarios.

## 15. What I Would Do Next

Implement PostgreSQL and Redis adapters behind the current repository and
middleware interfaces, then add a durable webhook worker with retry and
dead-letter semantics.
