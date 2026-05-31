# Middleware Chain

BankPort keeps the Gin chain explicit because the framework and API-platform
behavior are part of the challenge evidence.

## Global Chain

Order matters:

| Order | Middleware | Responsibility | Failure behavior |
| --- | --- | --- | --- |
| 1 | `gin.Recovery` | Convert panics to HTTP 500 and keep the process alive. | Gin recovery response. |
| 2 | `RequestIdentity` | Normalize or generate request/correlation IDs and echo response headers. | Invalid caller IDs are replaced, not trusted. |
| 3 | `RequestBodyLimit` | Apply `http.MaxBytesReader` before JSON binding. | `413 request_body_too_large`. |
| 4 | `Timeout` | Attach request-scoped context with `REQUEST_TIMEOUT`. | Downstream repository methods observe cancellation. |
| 5 | `Tracing` | Create OpenTelemetry span with route, status, request ID, and correlation ID. | Records 5xx as span errors. |
| 6 | `StructuredLogger` | Emit JSON request log with route, status, partner, app, and identity fields. | Always logs after handler execution. |
| 7 | `Metrics` | Observe Prometheus HTTP counters/histograms using route patterns. | Avoids raw-account-ID label cardinality. |

## Public Operational Routes

These routes intentionally bypass partner authentication:

- `GET /health/live`
- `GET /health/ready`
- `GET /metrics`

`/debug/pprof/*` is also unauthenticated, but only mounted when
`PPROF_ENABLED=true`. It is for trusted local/private diagnostics, not public
partner traffic.

## Partner API Chain

All `/v1` routes use:

| Order | Middleware | Responsibility |
| --- | --- | --- |
| 1 | `Authenticate` | Resolve API key from Bearer token or `X-API-Key`. |
| 2 | `RateLimit` | Enforce partner + route-pattern fixed-window budget. |
| 3 | `RequireScopes` | Enforce route-specific product capability. |
| 4 | `Idempotency` | Financial writes only: reserve key, hash method/route/body, replay or reject conflicts. |

Financial write routes use idempotency after scope checks so unauthorized
requests do not create idempotency records.

## Route Groups

| Group | Routes | Chain |
| --- | --- | --- |
| root | health, readiness, metrics, optional pprof | global only |
| `/v1/accounts` | balance, statements | global + auth + rate limit + `accounts:read` |
| `/v1/pix` | transfers | global + auth + rate limit + `pix:write` + idempotency |
| `/v1/payouts` | payout creation | global + auth + rate limit + `payouts:write` + idempotency |
| `/v1/refunds` | refund creation | global + auth + rate limit + `refunds:write` + idempotency |
| `/v1/webhooks` | endpoint registration | global + auth + rate limit + `webhooks:write` |
| `/v1/audit-logs` | audit evidence | global + auth + rate limit + `audit:read` |
| `/v1/sandbox` | deterministic scenarios | global + auth + rate limit + `sandbox:read` |

## Test Evidence

- `TestRequiresAuthentication`
- `TestRejectsInsufficientScope`
- `TestRateLimitExceeded`
- `TestIdempotentFinancialWriteReplaysCachedResponse`
- `TestIdempotencyConflict`
- `TestRejectsOversizedFinancialBody`
- `TestRequestIdentityRejectsUnsafeCallerSuppliedIDs`
- `TestMetricsUseRoutePatternForAccountIDs`
- `TestPprofDisabledByDefault`
- `TestPprofEnabledByConfiguration`
