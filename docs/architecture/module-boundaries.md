# Module Boundaries

| Module | Responsibility | Must not own |
| --- | --- | --- |
| `cmd/bankport-api` | canonical API binary entrypoint | domain rules |
| `cmd/api` | compatibility API entrypoint | domain rules |
| `cmd/bankportctl` | local sandbox CLI entrypoint | persistence internals |
| `internal/app` | executable orchestration for API and CLI | domain invariants |
| `internal/config` | environment parsing and defaults | request handling |
| `internal/domain` | business types, validation, invariants | Gin context |
| `internal/usecase` | application orchestration and ports consumed by use cases | Gin, persistence clients, Prometheus collectors |
| `internal/store` | secondary adapter behavior and seeded sandbox data | HTTP errors or route decisions |
| `internal/httpapi` | route registration, JSON binding, partner context extraction, use-case calls, response/error mapping | financial mutation policy |
| `internal/httpapi/middleware` | API-platform edge policy pipeline | money movement invariants |
| `internal/observability` | Prometheus metric definitions | route behavior |
| `internal/webhook` | HMAC signing | endpoint storage |

This shape keeps Gin visible for the challenge while preventing framework types
from leaking into application orchestration or core domain rules.

## Boundary Checks

- Handlers should stay thin: parse/auth context, call a use case, map response.
- Use cases may coordinate audit entries, metrics outcomes, command results,
  and webhook queue outcomes through ports.
- Domain code may expose business types and errors, but must not import
  delivery or infrastructure packages.
- Store and future PostgreSQL/Redis/webhook HTTP/provider packages are
  adapters. They implement ports; they do not define the application flow.
