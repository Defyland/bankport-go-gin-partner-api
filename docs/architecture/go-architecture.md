# Go Architecture

BankPort follows Go package boundaries instead of framework folders. There are
no controllers, services, repositories, and models copied from MVC terminology.
Packages are named by architectural role and dependency direction.

## Package Map

| Package | Layer | Notes |
| --- | --- | --- |
| `cmd/bankport-api` | executable adapter | Canonical API process. |
| `cmd/api` | executable adapter | Compatibility API process. |
| `cmd/bankportctl` | executable adapter | Local sandbox CLI. |
| `internal/app/bankportapi` | composition | Loads config, wires API dependencies, starts HTTP server. |
| `internal/app/bankportctl` | CLI application adapter | Parses CLI commands and calls sandbox read models. |
| `internal/httpapi` | primary adapter | Gin routes and HTTP response mapping. |
| `internal/httpapi/middleware` | primary adapter policy pipeline | Request identity, body limit, timeout, tracing, logging, metrics, auth, scopes, rate limit, idempotency. |
| `internal/usecase` | application core | Use cases and ports consumed by use cases. |
| `internal/domain` | domain core | Business types, invariants, and domain errors. |
| `internal/store` | secondary adapter | In-memory sandbox implementation of application ports. |
| `internal/webhook` | secondary adapter | HMAC signer. |
| `internal/observability` | secondary adapter support | Prometheus collector definitions. |
| `internal/config` | cross-cutting runtime config | Environment parsing and validation. |

## What Belongs Where

HTTP handlers:

- extract route params and partner context;
- bind request JSON;
- call a use case;
- map domain/application errors to the standard error envelope;
- return HTTP status and response body.

Use cases:

- orchestrate application behavior;
- call ports;
- record audit and application metrics;
- coordinate command results with webhook queue outcomes;
- avoid Gin, Prometheus concrete types, SQL, Redis, and Docker knowledge.

Domain:

- validates business inputs;
- names invariants and domain errors;
- stays independent from delivery and infrastructure.

Adapters:

- implement ports;
- translate framework, persistence, network, or observability APIs into
  application needs;
- are replaceable without changing domain invariants.

## Anti-Patterns This Repo Avoids

- Handler methods containing financial mutation policy.
- Domain types importing Gin or database/cache clients.
- A broad global repository interface owned by the infrastructure package.
- Persistence or cache models leaking into domain tests.
- Middleware order being implicit tribal knowledge.
- OpenAPI drifting away from runtime validation.

## Accepted Local Trade-Offs

- Domain response structs have JSON tags because the sandbox public contract is
  intentionally close to domain language. Introduce HTTP DTOs only when the
  public contract diverges.
- The in-memory store currently performs some command-state mutation and event
  queueing because it is the only secondary adapter. The use-case layer now owns
  application-level audit and metrics orchestration; a PostgreSQL adapter should
  move durable transaction boundaries behind the same ports.
