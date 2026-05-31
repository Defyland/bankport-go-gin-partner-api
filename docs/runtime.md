# Runtime

BankPort runs as a single Go HTTP process for the sandbox API. The preferred
binary name is `bankport-api`; `cmd/api` remains as a compatibility entrypoint
for older challenge runners.

## Startup

Startup validates configuration before binding the HTTP socket. Invalid
production secrets, malformed environment values, invalid timeouts, invalid
ports, and invalid request-size limits fail closed.

The startup log includes:

- `addr`
- `env`
- `go_version`
- `gomaxprocs`
- `num_cpu`
- `pprof_enabled`

These fields make local and CI evidence easier to compare with production
runtime behavior.

## HTTP Server Timeouts

The API server configures explicit HTTP timeouts:

| Timeout | Source | Purpose |
| --- | --- | --- |
| `ReadHeaderTimeout` | fixed 5 seconds | Reduce slowloris-style header abuse. |
| `ReadTimeout` | `REQUEST_TIMEOUT + 2s` | Bound request body and handler reads. |
| `WriteTimeout` | `REQUEST_TIMEOUT + 2s` | Bound response writes. |
| `IdleTimeout` | fixed 60 seconds | Reclaim idle keep-alive connections. |

Handlers also receive a request-scoped context from middleware based on
`REQUEST_TIMEOUT`.

## Graceful Shutdown

`bankport-api` listens for `SIGINT` and `SIGTERM`, stops accepting new
requests, and waits up to `SHUTDOWN_TIMEOUT` for in-flight work to finish. The
sandbox adapter checks request contexts before financial mutations so
canceled requests do not mutate balances.

## Optional pprof

Runtime profiling is disabled by default. Enable it only in trusted local or
private diagnostic environments:

```bash
PPROF_ENABLED=true go run ./cmd/bankport-api
```

When enabled, pprof routes are mounted under `/debug/pprof/` on the API server.
These routes are intentionally omitted from the partner OpenAPI contract because
they are operational diagnostics, not public partner API surface.
