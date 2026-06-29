# ADR 0006: Use Railway for a Single-service Partner API Demo

## Status

Accepted.

## Context

BankPort is a runnable HTTP service and benefits from a public demo deployment.
The repository already has a Dockerfile, liveness/readiness endpoints, and a
seeded in-memory sandbox, but it lacked a lightweight deployment surface that a
reviewer could understand without reproducing the full local Docker Compose
stack.

## Options considered

1. Do not add Railway support.
2. Add Railway for a single API process using the in-memory sandbox.
3. Block any demo deployment until PostgreSQL, Redis, and durable webhook
   workers exist.

## Decision

Add Railway config as code for a single `bankport-api` process. Keep the demo
honest by documenting it as a public sandbox surface, not as a production
topology. Require deploy-time key and secret overrides so the public demo does
not boot with the repository defaults.

## Consequences

Positive:

- the repo gains a truthful public-service deployment path;
- Railway can use `/health/ready` directly for activation checks;
- reviewers can exercise the API contract without recreating Prometheus and
  Grafana locally first.

Negative:

- the Railway path is still single-instance and in-memory;
- demo secrets must be configured out of band in Railway variables;
- the deployment does not prove distributed idempotency, Redis rate limiting,
  or durable webhook delivery.

## Verification evidence

- `railway.json`
- `docs/deployment/railway.md`
- `Dockerfile`
- `README.md`
