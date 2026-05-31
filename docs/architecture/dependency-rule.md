# Dependency Rule

BankPort follows a one-way dependency rule:

```text
delivery adapters -> use cases -> domain
secondary adapters -> domain
composition root -> all concrete packages
```

The application layer defines the ports it consumes. Secondary adapters satisfy
those ports. The domain does not depend on adapters.

## Allowed Imports

| From | May import |
| --- | --- |
| `internal/domain` | Go standard library only. |
| `internal/usecase` | `internal/domain`, Go standard library. |
| `internal/httpapi` | `config`, `domain`, `middleware`, `observability`, `usecase`, `webhook`, Gin, Prometheus HTTP handler. |
| `internal/httpapi/middleware` | `domain`, `observability`, Gin, OpenTelemetry, standard library. |
| `internal/store` | `config`, `domain`, standard library. |
| `internal/webhook` | `domain`, standard library. |
| `internal/app/*` | concrete adapters needed for composition. |

## Forbidden Direction

- `internal/domain` must not import `gin`, `sql`, Redis clients, Prometheus,
  OpenTelemetry, Docker, or CLI packages.
- `internal/usecase` must not import `gin`, `internal/httpapi`, `internal/store`,
  `internal/observability`, `database/sql`, Redis clients, or provider SDKs.
- `internal/store` must not import `internal/httpapi` or HTTP error envelopes.
- `internal/httpapi` should not implement financial mutation rules; it calls
  use cases and maps responses.

## Current Static Evidence

The current code respects the critical boundary:

- `internal/domain` has no imports from BankPort internal packages and no Gin
  imports.
- `internal/usecase` imports only `internal/domain` plus the standard library.
- `internal/httpapi/router.go` depends on `internal/usecase` and not on
  `internal/store`.
- `internal/store` implements use-case ports without importing HTTP packages.

## How To Check Manually

```bash
rg -n "github.com/gin-gonic/gin|gin\\.Context|gorm|redis|sql|httpapi|middleware" internal/domain internal/usecase internal/store
```

Expected result: no matches.

## Future Adapter Rule

PostgreSQL, Redis, webhook HTTP clients, and provider SDKs must enter under
adapter packages that implement ports consumed by `internal/usecase`. They
should not be imported directly by handlers or domain code.
