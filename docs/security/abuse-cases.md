# Abuse Cases

| Abuse case | Control | Test or evidence |
| --- | --- | --- |
| Partner reuses a Pix idempotency key with a larger amount. | Request hash conflict check. | `TestIdempotencyConflict` |
| Read-only API key attempts a financial write. | Scope middleware. | `TestRejectsInsufficientScope` |
| Partner tries to read another tenant account. | Repository partner ownership check and 404 response. | `TestTenantIsolationHidesForeignAccount` |
| Attacker floods one route. | Fixed-window rate limiting and 429 response. | `TestRateLimitExceeded` |
| Partner registers plain HTTP webhook receiver. | HTTPS validation except localhost. | `TestWebhookEndpointRequiresHTTPSOutsideLocalhost` |
| Webhook payload is tampered with. | Timestamped HMAC signature. | `TestSignerCreatesVersionedHMACSignature` |
| Missing credentials. | Authentication middleware. | `TestRequiresAuthentication` |

Future controls include Redis-backed distributed limits, OAuth client
credentials, mTLS at ingress, and signed webhook replay verification.
