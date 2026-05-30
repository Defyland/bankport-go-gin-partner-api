# Secrets

## Current environment variables

| Variable | Purpose |
| --- | --- |
| `BANKPORT_FULL_ACCESS_API_KEY` | Sandbox full-access key override. |
| `BANKPORT_READONLY_API_KEY` | Sandbox read-only key override. |
| `BANKPORT_OTHER_PARTNER_API_KEY` | Sandbox second-tenant key override. |
| `WEBHOOK_SIGNING_KEY` | HMAC secret for webhook delivery signatures. |

## Rules

- Never commit production API keys or webhook signing keys.
- Store only API-key hashes in production persistence.
- Rotate API keys by creating a new active key, validating partner traffic, then
  revoking the previous key.
- Use a cloud secret manager or sealed environment injection for production.
- Do not include raw secrets in structured logs, traces, audit entries, or error
  responses.
