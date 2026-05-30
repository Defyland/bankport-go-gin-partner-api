# Invariants

The implemented tests cover the highest-risk invariants:

| Invariant | Enforcement | Test evidence |
| --- | --- | --- |
| Missing or invalid API key cannot access `/v1` routes. | Auth middleware | `TestRequiresAuthentication` |
| Read-only keys cannot execute financial writes. | Scope middleware | `TestRejectsInsufficientScope` |
| Partners cannot read or mutate another partner account. | Repository ownership checks | `TestTenantIsolationHidesForeignAccount`, `TestCreatePixTransferDebitsOnlyPartnerOwnedAccount` |
| Financial writes require idempotency keys. | Idempotency middleware | API tests and OpenAPI contract |
| Same idempotency key and body returns the cached response. | Response capture cache | `TestIdempotentFinancialWriteReplaysCachedResponse` |
| Same idempotency key with a different body is rejected. | Request hash comparison | `TestIdempotencyConflict` |
| Account balance cannot go below zero. | Domain repository check | domain error path and API 422 behavior |
| Cumulative refunds cannot exceed the original transaction amount. | Refunded amount tracked per original transfer; production SQL guard documented in migration. | `TestCreateRefundRejectsCumulativeRefundAboveOriginalAmount` |
| Concurrent writes cannot overspend or over-refund. | Repository lock in sandbox; production requires row-level guarded updates. | `TestConcurrentPixTransfersDoNotOverspendAccount`, `TestConcurrentRefundsDoNotExceedOriginalAmount` |
| Webhook receivers outside localhost must use HTTPS. | Domain validation | `TestWebhookEndpointRequiresHTTPSOutsideLocalhost` |
| Rate limits produce 429 and retry metadata. | Rate-limit middleware | `TestRateLimitExceeded` |
| Idempotency records expire instead of growing without bound. | TTL-backed store cleanup. | `TestIdempotencyStoreExpiresRecords` |
| Rate-limit windows expire instead of growing without bound. | Window pruning in limiter. | `TestRateLimiterPrunesExpiredWindows` |
