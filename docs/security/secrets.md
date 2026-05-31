# Secrets

## Current environment variables

| Variable | Purpose |
| --- | --- |
| `BANKPORT_FULL_ACCESS_API_KEY` | Sandbox full-access key override. |
| `BANKPORT_READONLY_API_KEY` | Sandbox read-only key override. |
| `BANKPORT_OTHER_PARTNER_API_KEY` | Sandbox second-tenant key override. |
| `API_KEY_HASH_PEPPER` | HMAC pepper used when hashing API keys at rest. |
| `WEBHOOK_SIGNING_KEY` | HMAC secret for webhook delivery signatures. |
| `MAX_REQUEST_BODY_BYTES` | Maximum accepted request body size before parsing. |

## Rules

- Never commit production API keys or webhook signing keys.
- `BANKPORT_ENV=production` must fail startup unless the API-key pepper,
  webhook signing key, and seeded API keys are replaced with non-default values.
- Store only peppered API-key HMAC hashes in production persistence.
- Rotate `API_KEY_HASH_PEPPER` through a dual-read migration: write new hashes
  with the new pepper, accept old hashes during migration, then remove the old
  pepper after all active keys are reissued or rehashed.
- Rotate API keys by creating a new active key, validating partner traffic, then
  revoking the previous key.
- Use a cloud secret manager or sealed environment injection for production.
- Do not include raw secrets in structured logs, traces, audit entries, or error
  responses.
- Webhook delivery signatures derive endpoint-specific signing material from the
  root signing key and endpoint secret ID; production persistence should store
  endpoint secret metadata with a rotation history.
