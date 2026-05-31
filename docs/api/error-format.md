# Error Format

All application errors use the same envelope:

```json
{
  "error": {
    "code": "insufficient_scope",
    "message": "The API key does not include the required scope.",
    "details": {
      "required_scope": "pix:write"
    }
  },
  "request_id": "req_...",
  "correlation_id": "corr_..."
}
```

## Important codes

| Code | HTTP status | Meaning |
| --- | --- | --- |
| `authentication_required` | 401 | No API key was provided. |
| `invalid_api_key` | 401 | API key was not recognized. |
| `insufficient_scope` | 403 | API key lacks the route scope. |
| `idempotency_key_required` | 400 | Financial write omitted `Idempotency-Key`. |
| `idempotency_conflict` | 409 | Same key was reused with a different body. |
| `idempotency_original_failed` | 409 | A duplicate request waited for an in-flight request that did not produce a replayable response. |
| `idempotency_wait_timeout` | 408 | A duplicate request timed out while waiting for the original request with the same key. |
| `request_body_too_large` | 413 | Request body exceeds `MAX_REQUEST_BODY_BYTES`. |
| `validation_failed` | 400 | Request body failed domain validation. |
| `account_not_found` | 404 | Account does not exist for this partner. |
| `insufficient_funds` | 422 | Account balance cannot cover the command. |
| `refund_exceeds_original` | 422 | Refund would exceed the original transaction amount, including previous partial refunds. |
| `request_canceled` | 408 | Request context was canceled before completion. |
| `request_timeout` | 504 | Request exceeded the configured timeout. |
| `rate_limit_exceeded` | 429 | Partner exceeded the route rate limit. |
