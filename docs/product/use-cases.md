# Use Cases

## Read account balances

A partner reads available and pending balance before sending a payout or
internal reconciliation job.

## Submit Pix transfers

A partner submits a Pix transfer with an idempotency key. BankPort validates the
account, balance, scope, and request body before accepting the command.

## Submit payouts and refunds

Partners queue payout and refund commands while BankPort preserves replay-safe
responses and emits domain events.

## Register webhook receivers

Partners register HTTPS webhook receivers and choose event types. BankPort queues
signed delivery attempts when financial events are accepted.

## Investigate partner incidents

Support engineers search audit logs with request IDs and correlation IDs to
explain whether a failure came from auth, scopes, idempotency, rate limits, or
domain validation.
