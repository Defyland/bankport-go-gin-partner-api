# Glossary

| Term | Meaning |
| --- | --- |
| Partner | External organization allowed to integrate with BankPort APIs. |
| Developer app | A partner-owned API client with specific scopes and rate limits. |
| API key | Secret credential hashed at rest and accepted as a Bearer token or `X-API-Key`. |
| Scope | Permission string such as `accounts:read` or `pix:write`. |
| Account | Partner-owned financial account exposed through the API. |
| Financial write | A command that can mutate balance or create money movement evidence. |
| Idempotency key | Partner-supplied key that makes retries safe for financial writes. |
| Event | CloudEvents-like record emitted after domain changes. |
| Webhook delivery | Queued attempt to send a signed event to a partner endpoint. |
| Audit entry | Partner-visible evidence for accepted and rejected sensitive actions. |
| Correlation ID | Cross-system identifier used to connect logs, errors, events, and support tickets. |
