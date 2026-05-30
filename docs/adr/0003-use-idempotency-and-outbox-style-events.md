# ADR 0003: Use Idempotency and Outbox-style Events for Financial Writes

## Status

Accepted.

## Context

Partners retry financial write requests when networks fail. Without idempotency,
retries can double-debit an account. Without event evidence, partners cannot
reconcile accepted commands or webhook delivery issues.

## Options considered

1. Accept duplicate commands and rely on partner behavior
2. Use idempotency keys without event evidence
3. Use idempotency keys and outbox-style event records

## Decision

Financial writes require `Idempotency-Key`. The API hashes method, route, and
body. Same key and body replays the cached response; same key and different body
returns 409. Accepted commands emit versioned events and queue webhook delivery
evidence.

## Consequences

Positive:

- partner retries are safe
- support can connect request, event, webhook, and audit evidence
- production migration maps cleanly to PostgreSQL unique constraints

Negative:

- every write carries state-management overhead
- idempotency records need TTL and retention rules
- response caching must avoid leaking one partner response to another
