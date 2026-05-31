# Kubernetes

BankPort does not include Kubernetes manifests because the current scope is a
local, production-shaped sandbox without external infrastructure. This document
captures the deployment shape expected once PostgreSQL, Redis, webhook workers,
and secret management exist.

## Workloads

| Workload | Replicas | Notes |
| --- | --- | --- |
| `bankport-api` | 2+ | Stateless API pods after PostgreSQL/Redis adapters exist. |
| `webhook-worker` | 1+ | Planned worker for durable delivery retries and DLQ movement. |
| `otel-collector` | 1+ | Planned collector for trace export. |

## Probes

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: http
readinessProbe:
  httpGet:
    path: /health/ready
    port: http
```

Readiness must include persistence, idempotency/rate-limit backing stores, and
worker dependency checks once those adapters are implemented.

## Runtime Settings

Recommended pod settings for the API:

- Set `BANKPORT_ENV=production`.
- Inject `API_KEY_HASH_PEPPER` and `WEBHOOK_SIGNING_KEY` from a secret manager.
- Keep `PPROF_ENABLED=false` by default.
- Configure CPU limits deliberately because Go derives `GOMAXPROCS` from
  runtime CPU availability in modern releases.
- Route `/metrics` to Prometheus scraping only.

## Why Manifests Are Deferred

Adding full manifests now would be misleading because:

- The runtime state adapter is still in-memory.
- Rate limiting and idempotency are process-local.
- Webhook delivery is not yet a durable worker.
- Secrets are environment-backed for sandbox review, not backed by a cluster
  secret lifecycle.

The correct next step is to add manifests together with the PostgreSQL/Redis
adapters and worker so readiness, scaling, storage, and rollback semantics are
real.
