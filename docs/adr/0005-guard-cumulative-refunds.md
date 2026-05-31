# ADR 0005: Guard Cumulative Refunds

## Status

Accepted.

## Context

A refund request can be individually valid while the total refunded amount for
the original transaction becomes invalid. For example, a `2000` cent Pix
transfer followed by refunds of `1500` and `600` cents would over-credit the
account unless the system tracks cumulative refund state.

## Options considered

1. Validate only each refund amount against the original amount
2. Rely on partner idempotency keys to prevent duplicates
3. Track cumulative refunded amount and reject over-refunds

## Decision

Track refunded amount per original transfer and reject any refund that would
make the cumulative amount exceed the original transaction amount. The
production schema includes `pix_transfers.refunded_amount_cents` so PostgreSQL
can enforce the guard inside the financial write transaction.

## Consequences

Positive:

- prevents over-crediting accounts through partial refund sequences
- provides a clear SQL transaction pattern for production
- creates interview-grade evidence of domain risk analysis

Negative:

- refund acceptance now depends on original transaction state
- production writes must coordinate on the original transaction row
- concurrency tests become mandatory before replacing the sandbox adapter
