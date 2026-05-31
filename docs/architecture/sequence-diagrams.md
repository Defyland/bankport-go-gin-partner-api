# Sequence Diagrams

## Idempotent Pix transfer

```mermaid
sequenceDiagram
  participant P as Partner
  participant G as Gin middleware
  participant H as HTTP handler
  participant U as Use case
  participant R as Store adapter
  participant W as Webhook queue
  P->>G: POST /v1/pix/transfers + Idempotency-Key
  G->>G: request/correlation ID, auth, rate limit, scope
  G->>G: hash method + route + body
  G->>H: bind JSON and partner context
  H->>U: CreatePixTransfer
  U->>R: CreatePixTransfer port
  R->>R: verify account ownership and balance
  R->>R: debit account and append statement
  R->>W: queue signed delivery for matching endpoints
  R-->>U: transfer + queued delivery count
  U->>R: append audit evidence
  U-->>H: transfer + queued delivery count
  H-->>G: 202 response body
  G->>G: cache response by idempotency key
  G-->>P: 202 accepted
```

## Idempotency replay

```mermaid
sequenceDiagram
  participant P as Partner
  participant I as Idempotency middleware
  P->>I: retry same route/key/body
  I->>I: hash matches stored request hash
  I-->>P: cached status/body + Idempotency-Replayed=true
```
