# Pricing or Plans

BankPort is modeled as an API product platform, not a direct banking app.

## Sandbox

- deterministic seed accounts
- low default rate limit
- fake financial adapters
- webhook delivery evidence without external retries

## Partner launch

- scoped API keys per developer app
- higher negotiated rate limits
- audit log retention
- webhook signing and replay support

## Enterprise

- dedicated rate-limit budget
- mTLS and custom trust boundary controls
- extended audit retention
- custom webhook retry policy
- support SLO tied to correlation ID evidence
