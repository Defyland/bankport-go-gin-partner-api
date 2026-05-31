# ADR 0002: Use a Modular Monolith Before a Service Split

## Status

Accepted.

## Context

BankPort needs auth, scopes, rate limits, idempotency, financial commands,
events, audit, and observability to behave consistently. Splitting these early
would introduce network failure modes before the product contract is stable.

## Options considered

1. One modular API process
2. Separate auth, ledger, webhook, and audit services
3. Serverless functions per endpoint

## Decision

Use one Go API process with internal module boundaries. Keep domain types and
adapter behavior independent from Gin, but keep the deployable unit simple.

## Consequences

Positive:

- easier local setup and portfolio review
- fewer moving parts during contract design
- request-level tracing and audit evidence stay easy to follow

Negative:

- modules cannot scale independently yet
- a process-level failure affects all public API routes
- strong code boundaries are required to avoid a big ball of mud
