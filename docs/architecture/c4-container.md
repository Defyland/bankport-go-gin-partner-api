# C4 Container

```mermaid
flowchart TB
  Partner["Partner client"] --> API["Gin API process"]
  API --> Auth["Auth and scope middleware"]
  API --> RateLimit["Rate limit middleware"]
  API --> Idempotency["Idempotency middleware"]
  API --> Domain["Domain repository"]
  Domain --> Store["In-memory sandbox store"]
  Domain --> Events["Outbox-style events"]
  Events --> Webhooks["Webhook delivery queue"]
  API --> Metrics["Prometheus /metrics"]
  API --> Logs["Structured JSON logs"]
```

Production adapters will replace the in-memory store with PostgreSQL and Redis
without changing the public route handlers or OpenAPI contract.
