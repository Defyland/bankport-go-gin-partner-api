# Scalability

## Hot path

Financial writes are the hot path: authenticate partner, rate-limit route,
check idempotency, mutate account balance, append statement evidence, emit an
outbox event, and queue webhook delivery.

## Read-heavy operations

- balance reads
- statement reads
- sandbox scenario reads

These scale horizontally behind a load balancer once persistence moves to
PostgreSQL read replicas or cached read models.

## Write-heavy operations

- Pix transfers
- payouts
- refunds
- webhook endpoint changes

Writes require account-level consistency. The first production adapter should
use PostgreSQL transactions with optimistic account versions and idempotency
unique constraints.

Refunds also require a guarded update on the original transaction. The
production path should update `refunded_amount_cents` only when the cumulative
refund remains less than or equal to the original transfer amount. This avoids
the common partial-refund bug where each individual refund is valid but the sum
over-refunds the transaction.

## Fastest-growing tables

- `audit_entries`
- `outbox_events`
- `webhook_deliveries`
- `statement_entries`

Retention and partitioning should start with monthly partitions by
`occurred_at` or `created_at`.

## Hot partitions

Large partners can create hot keys on:

- partner route rate limits
- a single high-volume account
- webhook endpoint retry queues

Mitigations include per-account queues, partner-specific limits, outbox workers
partitioned by partner ID, and endpoint-level retry budgets.

## What can be eventual

- webhook delivery
- audit export
- partner usage metering
- statement read model refresh after command acceptance

## What must not be eventual

- idempotency conflict detection
- account ownership checks
- available balance mutation for accepted financial writes
- cumulative refund limit checks
- API-key and scope authorization
