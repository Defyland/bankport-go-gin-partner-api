# Deployment Readiness

BankPort will eventually run an API process plus optional worker processes for webhook retries, metering, and background delivery.

## Current posture

- Phase 0 documents the external partner contract and middleware-oriented architecture.
- PostgreSQL and Redis are planned core dependencies.
- OpenAPI, health, readiness, metrics, traces, and k6 coverage are expected before the first runtime slice is considered complete.

## Deferred platform work

- Kubernetes manifests are deferred until the API and worker process model is implemented.
- RabbitMQ or NATS are deferred until webhook retries and outbound event volume justify an additional messaging layer.
- Real internal financial adapters are deferred; fake adapters keep the partner contract stable while other systems evolve.
