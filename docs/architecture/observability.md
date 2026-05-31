# Observability

BankPort needs observability that explains partner impact, not only process
health. The implemented runtime exposes request identity, correlation identity,
domain metrics, traces, and runbook-linked alerts.

## Signals

| Signal | Evidence | Purpose |
| --- | --- | --- |
| Structured logs | `middleware.StructuredLogger` | Tie route, status, partner, developer app, request ID, and correlation ID. |
| Metrics | `internal/observability/metrics.go` | Track HTTP, financial commands, webhooks, rate limits, idempotency replays, and conflicts with route-pattern labels. |
| Traces | `middleware.Tracing` | Attach route pattern, status, request ID, and correlation ID to OpenTelemetry spans. |
| Health | `/health/live` | Detect process availability. |
| Readiness | `/health/ready` | Detect runtime dependency readiness boundary. |
| Dashboard | `deployments/grafana/dashboards/bankport-partner-api.json` | Review throughput, p95 latency, and domain controls. |
| Alerts | `deployments/prometheus/alerts.yml` | Page or ticket on 5xx rate, p95 latency, rate-limit spike, and idempotency conflict spike. |

## Operational Questions

- Which partner and developer app were affected?
- Was the request rejected by authentication, scope, rate limit, idempotency, or domain validation?
- Did a financial command emit an event and queue webhook delivery evidence?
- Is partner retry behavior causing idempotency conflicts?
- Is the API protecting itself with 429 responses or failing with 5xx responses?
- Are unmatched paths grouped as `unmatched` instead of creating high-cardinality
  metrics or trace names?
- Are caller-supplied request and correlation IDs bounded and normalized before
  they are copied into logs, response headers, metrics context, or traces?

## Known Gaps

- Traces use the OpenTelemetry API but no collector exporter is wired in local
  Compose yet.
- Webhook worker depth is modeled as queued delivery evidence; a durable worker
  and queue-depth metric are planned with the production adapter.
- Prometheus alerts are configured locally, but no Alertmanager route is shipped
  in this repository.
