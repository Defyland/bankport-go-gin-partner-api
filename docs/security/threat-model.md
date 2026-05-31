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
| Stolen API credential | peppered HMAC hashes, rotation workflows, and scoped access |
| Overbroad partner access | API products, scopes, and partner resolution middleware |
| Request replay | idempotency keys with request-hash comparison, in-flight reservation, and wait-and-replay behavior |
| Oversized request body | configured max body size before JSON parsing |
| Rate-limit abuse | process-local limits in the sandbox, Redis-backed limits planned for multi-instance production, and explicit error envelope |
| Webhook tampering | endpoint-specific HMAC signatures, supported event allowlist, URL user-info rejection, and delivery-attempt auditing |
| Internal adapter timeout | timeout middleware, circuit-breaker-ready adapter boundaries, and correlation IDs |

## Residual risks

- mTLS and full OAuth provider semantics are deferred; the first slice focuses on partner auth and policy enforcement.
- Service mesh is deferred; middleware, idempotency, and observability are the current edge controls.
- Real downstream integrations remain fake adapters until the public contract is stable.
- In-flight idempotency reservation is process-local until the PostgreSQL/Redis
  adapter phase introduces shared reservation semantics.

## Trust boundary decisions

- Public traffic terminates at the Gin API process.
- Auth, scope, rate limit, and idempotency middleware run before domain mutation.
- The store adapter hides cross-tenant resources with 404 responses.
- Webhook receivers are treated as untrusted external systems.
- Correlation IDs are partner-controlled metadata and are never used for authorization.

## Monitoring controls

- `bankport_partner_api_rate_limit_exceeded_total`
- `bankport_partner_api_idempotency_conflicts_total`
- `bankport_partner_api_financial_commands_total`
- structured log fields: `request_id`, `correlation_id`, `partner_id`, `developer_app_id`, route, status
