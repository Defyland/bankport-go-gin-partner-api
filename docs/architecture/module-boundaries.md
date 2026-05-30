# Module Boundaries

| Module | Responsibility | Must not own |
| --- | --- | --- |
| `cmd/api` | process boot, HTTP server, graceful shutdown | domain rules |
| `internal/config` | environment parsing and defaults | request handling |
| `internal/domain` | business types, validation, invariants | Gin context |
| `internal/store` | repository behavior and seeded sandbox data | HTTP errors |
| `internal/httpapi` | route registration and handlers | persistence details |
| `internal/httpapi/middleware` | cross-cutting HTTP policy | money movement rules |
| `internal/observability` | Prometheus metric definitions | route behavior |
| `internal/webhook` | HMAC signing | endpoint storage |

This shape keeps Gin visible for the challenge while preventing framework types
from leaking into core domain rules.
