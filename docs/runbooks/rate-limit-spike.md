# Runbook: Rate Limit Spike

## Symptoms

- partners report HTTP 429 responses
- Prometheus shows `rate_limit_exceeded_total` increasing
- support tickets share a small set of correlation IDs or routes

## Triage

1. Check the affected `partner_id` and route in logs.
2. Confirm the configured `RATE_LIMIT_PER_MINUTE`.
3. Compare 429 volume with successful request throughput.
4. Verify whether traffic is a legitimate batch, retry loop, or credential leak.

## Recovery

- For legitimate traffic, raise the partner rate limit after reviewing downstream capacity.
- For retry loops, point the partner to `Retry-After` and idempotency behavior.
- For suspected abuse, rotate the API key and preserve audit evidence.

## Follow-up

Add partner-specific dashboards when a customer repeatedly reaches limits.
