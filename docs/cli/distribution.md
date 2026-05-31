# CLI and Distribution

`bankportctl` is a local sandbox CLI for inspecting BankPort's seeded platform
state without requiring PostgreSQL, Redis, or a running API server.

## Commands

```bash
go run ./cmd/bankportctl apps list
go run ./cmd/bankportctl apps list --format json
go run ./cmd/bankportctl rate-limits inspect
go run ./cmd/bankportctl usage report --format json
```

Implemented commands are intentionally read-only because the current secondary
adapter uses in-memory sandbox state. The CLI does not pretend to persist
webhook replay or token rotation across processes.

| Command | Purpose | Current backing |
| --- | --- | --- |
| `apps list` | Show seeded developer apps, partner ownership, scopes, and request budgets. | In-memory seeded adapter. |
| `rate-limits inspect` | Show partner API rate-limit partitioning and configured per-minute budgets. | In-memory fixed-window policy metadata. |
| `usage report` | Show sandbox counts for partners, apps, accounts, commands, events, webhooks, and audit entries. | In-memory adapter snapshot. |

## Explicit Gaps

| Command | Gap | Why it is not implemented yet |
| --- | --- | --- |
| `webhooks replay` | Requires durable webhook delivery state and retry history. | Current API queues in-memory delivery evidence only. |
| `tokens rotate` | Requires persistent API key storage, dual-read rotation, revocation, and audit durability. | Current keys are seeded from environment for sandbox review. |

These commands become meaningful after the planned PostgreSQL adapter,
outbox/webhook worker, and API-key rotation store are implemented.

## Release Shape

The repository builds two binaries:

- `bankport-api`: the canonical API process.
- `bankportctl`: the local sandbox operator/developer CLI.

GoReleaser configuration is intentionally small and builds static binaries for
Linux and Darwin on amd64/arm64. It does not publish artifacts from local
validation; release publishing belongs to CI/CD.
