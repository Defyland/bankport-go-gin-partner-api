# ADR 0004: Bound Replay and Rate-limit State

## Status

Accepted.

## Context

BankPort keeps sandbox idempotency records and rate-limit windows in memory.
This is acceptable for local review, but unbounded in-memory state is not
acceptable as a production pattern. A partner can generate many idempotency keys
or many route keys over time, causing memory growth unrelated to active traffic.

## Options considered

1. Keep state forever in process memory
2. Add in-memory TTL and pruning in the sandbox
3. Require Redis before the API can run

## Decision

Use TTL and expired-window pruning in the sandbox implementation. Keep Redis as
the production backing store for distributed rate limits and idempotency cache
reads.

## Consequences

Positive:

- prevents the sandbox from teaching an unsafe operational pattern
- keeps local setup simple
- maps cleanly to Redis TTL semantics later

Negative:

- in-memory state is still not shared across replicas
- process restart still loses replay cache
- production remains dependent on a future Redis adapter
