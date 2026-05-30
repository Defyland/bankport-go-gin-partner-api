# Roadmap

## Now

- Gin API with versioned routes
- API-key authentication and scoped authorization
- fixed-window rate limiting
- idempotent financial writes
- in-memory sandbox repository
- webhook registration and queued delivery evidence
- Prometheus metrics, structured logs, request IDs, and tracing spans

## Next

- PostgreSQL repository using `db/migrations/001_init.sql`
- Redis-backed rate limiting and idempotency TTLs
- webhook worker with retries, dead-letter queue, and replay endpoint
- OpenTelemetry exporter wiring for a collector
- k6 benchmark results from Docker Compose

## Later

- OAuth client credentials for production partners
- mTLS at the ingress boundary
- partner portal for API keys, webhook endpoints, and audit search
- provider adapters for real Pix, payout, and refund rails
- multi-region read replicas and account-level write partitioning
