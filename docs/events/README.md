# BankPort Event Contracts

BankPort emits partner-facing and operational events around API usage, financial write commands, and webhook delivery. These contracts must stay stable even when internal adapters evolve.

## Envelope

Every event must include:

- `event_id`
- `event_type`
- `schema_version`
- `occurred_at`
- `producer`
- `partner_id`
- `developer_app_id`
- `correlation_id`
- `payload`

## Compatibility policy

- Consumers deduplicate by `event_id`.
- New fields must be optional until all known consumers accept them.
- Error and rate-limit events must preserve the same `correlation_id` returned to the partner.
- Webhook replay must preserve the original event payload and append delivery-attempt evidence separately.

## Versioned schemas

- [pix.transfer.created.v1.json](pix.transfer.created.v1.json)
- [payout.created.v1.json](payout.created.v1.json)
- [refund.created.v1.json](refund.created.v1.json)
- [webhook.delivery.requested.v1.json](webhook.delivery.requested.v1.json)
- [api.rate_limit_exceeded.v1.json](api.rate_limit_exceeded.v1.json)
