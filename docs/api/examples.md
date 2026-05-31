# API Examples

## Authentication

```bash
curl -H "Authorization: Bearer bp_sandbox_full_access_key" \
  http://localhost:8080/v1/accounts/acct_sandbox_001/balance
```

`X-API-Key` is accepted for partner compatibility, but Bearer tokens are the
documented default.

## Pix transfer

```bash
curl -X POST http://localhost:8080/v1/pix/transfers \
  -H "Authorization: Bearer bp_sandbox_full_access_key" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-demo-001" \
  -d '{
    "source_account_id": "acct_sandbox_001",
    "amount_cents": 1500,
    "currency": "BRL",
    "pix_key": "merchant@example.com",
    "description": "Marketplace settlement"
  }'
```

Successful response:

```json
{
  "data": {
    "transfer": {
      "id": "pix_...",
      "partner_id": "partner_sandbox_bank",
      "source_account_id": "acct_sandbox_001",
      "amount_cents": 1500,
      "currency": "BRL",
      "pix_key": "merchant@example.com",
      "status": "accepted"
    },
    "queued_webhook_deliveries": 0
  },
  "request_id": "req_...",
  "correlation_id": "req_..."
}
```

## Idempotency replay

Retrying the same method, route, idempotency key, and body returns the cached
status and body with:

```text
Idempotency-Replayed: true
```

Changing the body with the same key returns `409 idempotency_conflict`.
If a duplicate request arrives while the first request is still executing in the
same API process, it waits for the original response and then replays it instead
of executing the financial handler twice.

Sandbox idempotency records expire after `IDEMPOTENCY_TTL`, which defaults to
24 hours. The production adapter should persist the same expiry in PostgreSQL
and use Redis only as a fast cache plus shared reservation layer, not as the
source of truth.
