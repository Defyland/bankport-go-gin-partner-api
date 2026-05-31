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
| Concurrent requests could reuse the same idempotency key before the first response was cached. | Two identical in-flight financial writes could execute the handler twice in a single API process. | Added in-flight idempotency reservation with wait-and-replay behavior plus a concurrent middleware regression test. |
| Request-timeout responses could be cached as idempotent outcomes. | A canceled write that did not produce a durable financial result should not permanently replay cancellation for the same key. | Idempotency now skips caching HTTP 408 responses and has regression coverage for retry execution after timeout. |
| Rate-limit windows had no cleanup path. | Many partners/routes over time could leave stale limiter keys in memory. | Added expired-window pruning and test coverage. |
| Production config accepted sandbox defaults. | A production-shaped API must fail closed when secrets, peppers, or sandbox API keys are not replaced. | Added `Config.Validate`, startup validation, and tests that reject production defaults. |
| Malformed env values fell back to defaults. | Silent fallback hides deployment mistakes and makes production behavior depend on typo-prone config. | Config loading now records parse errors and startup validation rejects malformed port, duration, bool, integer, and log-level values. |
| Sandbox ID generation panicked when the entropy reader failed. | A rare OS entropy failure should not crash a reference API for non-secret IDs. | Added a hashed time/counter fallback and regression test while keeping real secrets environment-backed and validated. |
| Request bodies were read without a configured size cap. | An attacker could force memory pressure before JSON validation. | Added `MAX_REQUEST_BODY_BYTES`, `http.MaxBytesReader`, 413 error mapping, and API coverage. |
| Repository methods ignored canceled request contexts. | Timeouts should stop work before financial state changes, not only decorate the request. | Added context checks before reads/writes and a canceled-write regression test. |
| Request ID entropy fallback returned a constant value. | Duplicate request IDs make incident correlation and audit reconstruction unreliable exactly when the runtime is already degraded. | Added hashed time/counter fallback and regression coverage for distinct fallback request IDs. |
| Metrics and traces could use raw unmatched paths. | High-cardinality path labels make dashboards and tracing noisier under malformed traffic. | Route labels now fall back to `unmatched`, and account routes are tested for route-pattern labels. |
| Webhook signatures used only the root signing key. | Endpoint-specific signing material is closer to production rotation and blast-radius expectations. | Signatures now derive per-endpoint material from the root signing key and endpoint secret ID. |
| Observability had dashboards but no alert rules. | A reviewer should see how operators know when to act, not only where charts live. | Added Prometheus alert rules and an observability subsystem doc. |
| Validation tooling was not fully pinned. | Unpinned local/CI lint tooling can drift and make evidence non-reproducible. | Added Go `toolchain go1.26.3`, CI version envs, and pinned Redocly CLI `2.31.5` in CI, Makefile, and docs. |
| Spec-driven evidence was missing. | The new standard requires explicit acceptance criteria, plan, and verification report before claiming readiness. | Added `docs/spec-driven/` with readiness spec, plan, and report. |

## What I Would Prioritize Next

1. Implement the PostgreSQL repository with transaction tests around account
   debits, idempotency unique constraints, outbox insert, and cumulative refund
   guarded updates.
2. Add Redis-backed distributed rate limiting and shared idempotency
   reservation/cache reads while keeping PostgreSQL as the source of truth.
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
