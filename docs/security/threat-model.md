# Threat Model

## Assets

- partner API keys, client credentials, and webhook signing secrets
- scope assignments and API-product permissions
- idempotency records and cached responses
- request logs, usage events, and audit logs
- webhook payloads and delivery attempts

## Trust boundaries

- partners call public APIs over authenticated HTTP
- middleware enforces auth, scopes, rate limits, and idempotency before adapter calls
- fake internal adapters represent downstream financial systems
- webhook endpoints are untrusted external receivers

## Primary threats

| Threat | Control |
| --- | --- |
| Stolen API credential | hashed secrets, rotation workflows, and scoped access |
| Overbroad partner access | API products, scopes, and partner resolution middleware |
| Request replay | idempotency keys with request-hash comparison |
| Rate-limit abuse | Redis-backed limits and explicit error envelope |
| Webhook tampering | HMAC signatures and delivery-attempt auditing |
| Internal adapter timeout | timeout middleware, circuit-breaker-ready adapter boundaries, and correlation IDs |

## Residual risks

- mTLS and full OAuth provider semantics are deferred; the first slice focuses on partner auth and policy enforcement.
- Service mesh is deferred; middleware, idempotency, and observability are the current edge controls.
- Real downstream integrations remain fake adapters until the public contract is stable.
