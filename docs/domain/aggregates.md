# Aggregates

## Partner

Root: `Partner`

Entities:

- developer app
- API key
- scope set

Invariants:

- API keys authenticate a single developer app.
- scope checks happen before route handlers execute.
- rate limits are partitioned by partner and route.

## Account

Root: `Account`

Entities:

- statement entry
- Pix transfer
- payout
- refund

Invariants:

- financial writes can only target accounts owned by the authenticated partner.
- available balance cannot become negative.
- all accepted writes append account statement evidence.

## Webhook endpoint

Root: `WebhookEndpoint`

Entities:

- webhook delivery
- event contract

Invariants:

- non-localhost endpoints must use HTTPS.
- event delivery signatures use the event payload and a timestamp.
- duplicate delivery is allowed and consumers must deduplicate by `event_id`.
