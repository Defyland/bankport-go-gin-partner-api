# Railway Deployment

BankPort includes `railway.json` for a single-service Railway deployment that
keeps the partner API publicly runnable without pretending the rest of the
production topology already exists.

## Runtime shape

- builder: `Dockerfile`
- process: `/bankport`
- health check: `/health/ready`
- runtime state: in-memory seeded sandbox adapter
- readiness/liveness endpoints: `/health/ready` and `/health/live`

The Railway path is a reviewer/demo surface. It proves the HTTP contract,
authn/authz boundaries, rate limits, idempotency semantics, audit evidence, and
observability endpoints from one deployable process.

## Required variables

Set these in Railway so the public demo does not reuse the repo defaults:

```bash
BANKPORT_ENV=production
API_KEY_HASH_PEPPER=<32+-character-secret>
WEBHOOK_SIGNING_KEY=<32+-character-secret>
BANKPORT_FULL_ACCESS_API_KEY=<demo-full-access-key>
BANKPORT_READONLY_API_KEY=<demo-read-only-key>
BANKPORT_OTHER_PARTNER_API_KEY=<demo-other-partner-key>
```

Optional variables:

```bash
LOG_LEVEL=info
RATE_LIMIT_PER_MINUTE=120
REQUEST_TIMEOUT=3s
SHUTDOWN_TIMEOUT=8s
OTEL_SERVICE_NAME=bankport-partner-api
```

Railway injects `PORT`; BankPort already binds to that variable directly.

## Five-minute verification

After Railway deploys:

```bash
curl -fsS "$RAILWAY_PUBLIC_DOMAIN/health/live"
curl -fsS "$RAILWAY_PUBLIC_DOMAIN/health/ready"
curl -fsS "$RAILWAY_PUBLIC_DOMAIN/metrics" | head
curl -fsS "$RAILWAY_PUBLIC_DOMAIN/v1/accounts/acct_sandbox_001/balance" \
  -H "Authorization: Bearer $BANKPORT_READONLY_API_KEY"
```

Expected API result: HTTP `200` with the seeded account balance payload.

## Limits

- The Railway demo is intentionally single-process and in-memory.
- Idempotency, rate limiting, audit, and webhook evidence are truthful for one
  instance, not for a replicated fleet.
- PostgreSQL persistence, Redis-backed coordination, durable webhook workers,
  and key rotation remain required before real money-movement production claims.
