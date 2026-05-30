# Operational Cost

## Current MVP cost

The local runtime has one API process plus optional Prometheus and Grafana in
Docker Compose. This keeps setup cost low while still proving observability and
API behavior.

## Production components

| Component | Cost introduced | Reason |
| --- | --- | --- |
| API deployment | deploy pipeline, scaling policy | public partner traffic |
| PostgreSQL | backups, migrations, locks, query tuning | source of truth and idempotency constraints |
| Redis | memory sizing, eviction monitoring | low-latency distributed rate limits |
| Webhook worker | retry operations, DLQ handling | reliable partner event delivery |
| Prometheus/Grafana | dashboard upkeep, storage | incident response and SLO evidence |
| Secret manager | rotation workflow | API and webhook key safety |

## Debugging cost

More services mean more failure modes. The current modular monolith avoids early
network boundaries so a maintainer can trace auth, idempotency, domain mutation,
events, and audit logs in one process.

## Deploy cost

The Dockerfile runs tests before producing the image. CI also runs format, vet,
race-enabled tests, coverage, OpenAPI lint, security scan, compose validation,
and Docker build.

## Accepted trade-offs

- In-memory sandbox state is not durable. This is acceptable for the first
  contract slice and keeps local setup fast.
- PostgreSQL/Redis adapters are deferred. This avoids adding infrastructure
  before the contract and invariants are proven.
- Webhook delivery is queued but not sent by a worker yet. The public contract
  and signing rules are still testable.
