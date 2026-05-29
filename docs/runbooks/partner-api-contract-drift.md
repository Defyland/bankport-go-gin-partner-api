# Partner API Contract Drift

Use this runbook when a partner starts failing after a BankPort request or webhook payload change.

## Triage

- Identify the affected route, API product, and scope.
- Compare the payload or error envelope with the versioned schema under `docs/events/` or the OpenAPI contract.
- Check whether the same `correlation_id` appears in request logs and usage events.
- Verify idempotency behavior if the failing route is a financial write operation.

## Recovery

- Restore backward-compatible fields or introduce a new version.
- Replay failed webhook deliveries only after confirming the contract is accepted.
- Keep request and audit evidence intact; do not rewrite prior partner responses.
