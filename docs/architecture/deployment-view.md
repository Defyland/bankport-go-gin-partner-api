# Deployment View

## Local

```text
docker compose up --build
```

Containers:

- `api`: BankPort Gin API
- `prometheus`: scrapes `/metrics`
- `grafana`: provisions the BankPort dashboard

## Production target

The first production target is a single horizontally scaled API deployment
behind an ingress or load balancer. PostgreSQL becomes the source of truth for
accounts, idempotency records, outbox events, webhooks, and audit entries. Redis
backs low-latency rate limits and short-lived idempotency cache reads.

## Deferred platform work

Kubernetes, service mesh, and multi-region active-active writes are deferred
until provider adapters and partner traffic justify the operational cost.
