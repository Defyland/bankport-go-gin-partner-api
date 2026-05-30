# Secrets

## Current environment variables

| Variable | Purpose |
| --- | --- |
| `BANKPORT_FULL_ACCESS_API_KEY` | Sandbox full-access key override. |
| `BANKPORT_READONLY_API_KEY` | Sandbox read-only key override. |
| `BANKPORT_OTHER_PARTNER_API_KEY` | Sandbox second-tenant key override. |
| `API_KEY_HASH_PEPPER` | HMAC pepper used when hashing API keys at rest. |
| `WEBHOOK_SIGNING_KEY` | HMAC secret for webhook delivery signatures. |

## Rules

- Never commit production API keys or webhook signing keys.
- Store only peppered API-key HMAC hashes in production persistence.
- Rotate `API_KEY_HASH_PEPPER` through a dual-read migration: write new hashes
  with the new pepper, accept old hashes during migration, then remove the old
  pepper after all active keys are reissued or rehashed.
- Rotate API keys by creating a new active key, validating partner traffic, then
  revoking the previous key.
- Use a cloud secret manager or sealed environment injection for production.
- Do not include raw secrets in structured logs, traces, audit entries, or error
  responses.
