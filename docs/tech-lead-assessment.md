# Tech Lead Assessment

This document evaluates BankPort as if a hiring team were using it to assess
senior or tech-lead readiness. It separates what already demonstrates seniority
from what needed technical improvement and what should come next for the
strongest production-readiness signal.

## Executive assessment

BankPort is now credible as a senior portfolio project because it has a real
product boundary, runnable API, explicit domain language, authorization and
idempotency controls, event contracts, observability, CI, benchmark evidence,
ADRs, and runbooks.

The strongest signal is not the Gin setup. It is the way the repository explains
and tests financial API failure modes: BOLA, duplicate writes, rate limits,
webhook signing, audit evidence, and cumulative refund safety.

## What was already strong

| Area | Evidence | Why it matters |
| --- | --- | --- |
| Product framing | README and `docs/product/` | Shows the author can define users, workflows, non-goals, and roadmap. |
| Middleware architecture | `internal/httpapi/middleware/` | Makes auth, scopes, rate limits, idempotency, logs, metrics, and tracing explicit. |
| Domain tests | `internal/httpapi/router_test.go`, `internal/store/memory_test.go` | Tests business risk instead of only handler happy paths. |
| Security posture | `docs/security/` and auth/scope tests | Shows awareness of BOLA, secrets, rate limits, webhooks, and auditability. |
| Operational evidence | CI, Prometheus, Grafana, runbooks, k6 scripts | Gives reviewers a path to operate and inspect the system. |

## What needed improvement

| Finding | Severity | Why it matters | Change made |
| --- | --- | --- | --- |
| Cumulative refunds could exceed the original transaction if split across multiple requests. | Critical | This is a real money-movement bug. Each refund can look valid while the sum over-refunds. | Added refunded amount tracking, SQL guard documentation, event evidence, and regression test. |
| Idempotency records did not expire. | High | Long-running APIs cannot keep replay cache state forever. Unbounded maps become memory risk. | Added TTL-backed idempotency store with cleanup and tests. |
| Rate-limit windows did not prune old keys. | Medium | Per-partner route maps can grow forever under changing traffic patterns. | Added expired-window pruning and tests. |
| Production-readiness narrative was spread across files. | Medium | A reviewer should not have to infer the author’s judgment. | Added this assessment and linked decisions to tests/docs. |

## Spec-driven acceptance criteria

| Requirement | Evidence |
| --- | --- |
| Financial writes are retry-safe. | Idempotency middleware, OpenAPI `Idempotency-Key`, replay/conflict tests. |
| Tenant isolation is enforced. | Repository ownership checks and BOLA tests. |
| Refunds cannot over-credit an account. | `TestCreateRefundRejectsCumulativeRefundAboveOriginalAmount` and `refunded_amount_cents` migration guard. |
| Concurrent writes preserve financial invariants. | `TestConcurrentPixTransfersDoNotOverspendAccount` and `TestConcurrentRefundsDoNotExceedOriginalAmount`. |
| Middleware state is bounded. | `TestIdempotencyStoreExpiresRecords` and `TestRateLimiterPrunesExpiredWindows`. |
| Public API contract is reviewable. | `openapi.yaml` and `docs/api/`. |
| Operations are observable. | `/metrics`, structured logs, request IDs, correlation IDs, Grafana dashboard, runbooks. |
| CI verifies quality gates. | `.github/workflows/ci.yml`. |
| Performance is measured. | `benchmarks/baseline.md` and `benchmarks/results/2026-05-30-go-benchmark.txt`. |

## What would impress a senior hiring panel next

1. Implement the PostgreSQL repository behind the current store contract.
   The highest-value proof is a transaction that atomically checks idempotency,
   locks or version-checks the account row, mutates balance, appends statement
   evidence, updates cumulative refund state, inserts outbox events, and writes
   audit entries.

2. Move rate limiting and idempotency cache reads to Redis.
   This shows distributed-systems judgment: local memory is acceptable for the
   sandbox, but horizontally scaled APIs need shared short-lived state.

3. Add a durable webhook worker.
   The next production-ready slice should include retry policy, exponential
   backoff, dead-letter state, replay endpoint, signature verification fixtures,
   and metrics for queue age and delivery outcomes.

4. Run k6 against Docker Compose in CI or a reproducible local script.
   The current benchmark evidence is useful; a stronger signal is repeatable
   p50/p95/p99 under smoke, load, stress, and spike profiles with stored output.

## How to discuss this repository in an interview

Lead with the product risk: BankPort is about giving partners a safe financial
API boundary, not about demonstrating Gin syntax. Then point to the controls:
scopes, tenant isolation, idempotency, cumulative refund safety, event contracts,
webhook signatures, observability, runbooks, and ADRs. Finally, be explicit
about deferred complexity: PostgreSQL, Redis, and workers are intentionally the
next production slice because the public contract and invariants needed to be
settled first.
