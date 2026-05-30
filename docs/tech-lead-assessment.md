# Tech Lead Assessment

This document is written for a reviewer evaluating whether BankPort demonstrates
senior/tech-lead judgment. It separates evidence already implemented from the
next moves that would make the repository closer to production-ready.

## What Already Signals Seniority

- The project is framed as a product, not a framework demo: partner developers,
  partner operations, financial write safety, webhook evidence, and auditability
  are explicit.
- The API has real edge controls: API-key authentication, scopes, rate limits,
  idempotency, standardized errors, request IDs, correlation IDs, and audit logs.
- The domain model includes financial invariants: tenant ownership, available
  balance checks, idempotency conflicts, webhook HTTPS validation, and cumulative
  refund limits.
- The architecture is intentionally constrained: modular monolith before
  microservices, fake adapters before real banking providers, PostgreSQL/Redis
  planned behind stable interfaces.
- Observability is operationally useful: structured logs, domain metrics,
  traces, dashboard, alert rules, and runbooks.
- The repository has CI, OpenAPI, Docker assets, benchmark scripts, measured
  benchmark output, ADRs, security docs, scalability docs, and cost analysis.

## What Needed Technical Improvement

| Finding | Why it matters | Change applied |
| --- | --- | --- |
| Refund validation only checked each refund against the original amount. | Multiple partial refunds could exceed the original Pix transfer, a real financial correctness bug. | Added cumulative refund tracking, production SQL guard documentation, and regression test. |
| Idempotency cache had no expiry behavior. | A long-lived API process could grow memory without bound and replay stale financial responses longer than intended. | Added configurable `IDEMPOTENCY_TTL`, expiry cleanup, and test coverage. |
| Rate-limit windows had no cleanup path. | Many partners/routes over time could leave stale limiter keys in memory. | Added expired-window pruning and test coverage. |
| Observability had dashboards but no alert rules. | A reviewer should see how operators know when to act, not only where charts live. | Added Prometheus alert rules and an observability subsystem doc. |
| Spec-driven evidence was missing. | The new standard requires explicit acceptance criteria, plan, and verification report before claiming readiness. | Added `docs/spec-driven/` with readiness spec, plan, and report. |

## What I Would Prioritize Next

1. Implement the PostgreSQL repository with transaction tests around account
   debits, idempotency unique constraints, outbox insert, and cumulative refund
   guarded updates.
2. Add Redis-backed distributed rate limiting and idempotency cache reads while
   keeping PostgreSQL as the source of truth.
3. Add a durable webhook worker with retry queue, dead-letter queue, replay
   endpoint, endpoint-level retry budget, and queue-depth metrics.
4. Wire OpenTelemetry exporter and Prometheus Alertmanager in Compose.
5. Add provider adapter interfaces with contract tests for Pix, payout, and
   refund rails before integrating any real provider.

## Reviewer Narrative

BankPort is strongest when described as a production-shaped partner API sandbox:
it proves product thinking, domain safety, operational controls, and explicit
trade-offs without pretending that fake adapters are real banking integrations.
That honesty is a senior signal. The next strongest signal would be replacing
the in-memory repository with PostgreSQL/Redis adapters while preserving the
existing tests and public API contract.
