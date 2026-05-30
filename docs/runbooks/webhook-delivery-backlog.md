# Runbook: Webhook Delivery Backlog

## Symptoms

- queued webhook deliveries grow faster than they drain
- partners report missing Pix, payout, or refund events
- event creation succeeds but partner operations cannot reconcile

## Triage

1. Identify event type, endpoint ID, and partner ID.
2. Check whether endpoint URL was recently changed.
3. Verify webhook signature secret rotation history.
4. Compare event payload with the schema under `docs/events/`.
5. Preserve original `event_id` and `correlation_id` for replay.

## Recovery

- Pause retries for endpoints returning permanent errors.
- Replay only after the partner confirms contract compatibility.
- Move poison deliveries to a dead-letter queue when worker support is enabled.

## Current MVP note

The API queues delivery evidence in memory. A durable worker, retry queue, and
dead-letter queue are next-phase production work.
