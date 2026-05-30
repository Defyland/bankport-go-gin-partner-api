# Authorization Matrix

| Route | Scope | Notes |
| --- | --- | --- |
| `GET /v1/accounts/{id}/balance` | `accounts:read` | Account must belong to authenticated partner. |
| `GET /v1/accounts/{id}/statements` | `accounts:read` | Foreign accounts return 404. |
| `POST /v1/pix/transfers` | `pix:write` | Requires `Idempotency-Key`. |
| `POST /v1/payouts` | `payouts:write` | Requires `Idempotency-Key`. |
| `POST /v1/refunds` | `refunds:write` | Requires `Idempotency-Key`. |
| `POST /v1/webhooks/endpoints` | `webhooks:write` | Requires HTTPS outside localhost. |
| `GET /v1/audit-logs` | `audit:read` | Returns entries for current partner only. |
| `GET /v1/sandbox/scenarios` | `sandbox:read` | Safe sandbox metadata. |

Seeded sandbox keys:

| Key | Scopes |
| --- | --- |
| `bp_sandbox_full_access_key` | all MVP scopes |
| `bp_sandbox_readonly_key` | `accounts:read`, `sandbox:read` |
| `bp_sandbox_other_partner_key` | second tenant for isolation tests |
