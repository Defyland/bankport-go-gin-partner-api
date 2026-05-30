# Deployment Readiness

BankPort will eventually run an API process plus optional worker processes for webhook retries, metering, and background delivery.

## Current posture

- Gin API runtime is implemented with health, readiness, metrics, tracing
  middleware, structured logs, auth, scopes, rate limits, idempotency, events,
  webhooks, and audit logs.
- Dockerfile and Docker Compose are present for API, Prometheus, and Grafana.
- OpenAPI and k6 benchmark scripts are present.
- PostgreSQL and Redis remain planned production dependencies, with schema and
  rationale documented before adapter implementation.

## Deferred platform work

- Kubernetes manifests are deferred until the API and worker process model is implemented.
- RabbitMQ or NATS are deferred until webhook retries and outbound event volume justify an additional messaging layer.
- Real internal financial adapters are deferred; fake adapters keep the partner contract stable while other systems evolve.
