# State Machines

## Pix transfer

```text
accepted -> settled
accepted -> rejected
```

The sandbox implementation stops at `accepted`. Provider settlement and rejection
states are deferred to the adapter phase.

## Payout

```text
queued -> processing -> paid
queued -> processing -> failed
failed -> queued
```

The sandbox implementation queues payouts and emits `payout.created.v1`.

## Refund

```text
accepted -> settled
accepted -> rejected
```

Refunds are accepted only when the original transaction belongs to the partner
and account.

## Webhook delivery

```text
queued -> delivered
queued -> failed -> queued
failed -> dead_lettered
```

The API queues delivery evidence. Worker retries and dead-letter handling are
documented as the next phase.
