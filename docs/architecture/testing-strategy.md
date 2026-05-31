# Testing Strategy

BankPort tests each architectural layer at the boundary where defects would
matter.

## Domain Tests

Package: `internal/domain`

Purpose:

- validate financial input invariants without Gin, database, Redis, or provider
  SDKs;
- preserve `errors.Is` behavior for domain errors;
- keep domain rules deterministic and cheap to run.

Examples:

- `TestPixTransferRequestValidate`
- `TestPayoutRequestValidatesBankAccountShape`
- `TestWebhookEndpointRejectsUnsupportedEventType`

## Use Case Tests

Package: `internal/usecase`

Purpose:

- verify application orchestration with fake adapters;
- ensure audit and metrics policy does not live in Gin handlers;
- test success and rejection paths without HTTP or in-memory store coupling.

Examples:

- `TestCreatePixTransferAuditsAndRecordsMetrics`
- `TestCreatePixTransferAuditsRejectedDomainError`
- `TestRegisterWebhookEndpointAuditsResult`

## Adapter Tests

Packages: `internal/store`, `internal/webhook`, `internal/httpapi`,
`internal/httpapi/middleware`, `internal/app/bankportctl`

Purpose:

- verify in-memory adapter behavior, tenant isolation, concurrency guards, and
  cumulative refund protection;
- verify HMAC webhook signing;
- verify HTTP route/middleware behavior and error envelopes;
- verify CLI output and usage errors.

Examples:

- `TestConcurrentPixTransfersDoNotOverspendAccount`
- `TestConcurrentRefundsDoNotExceedOriginalAmount`
- `TestSignerDerivesEndpointSpecificSignatures`
- `TestIdempotencyConcurrentSameKeyRunsHandlerOnce`
- `TestAppsListJSON`

## Contract and Runtime Tests

Purpose:

- Redocly validates OpenAPI contract drift.
- `go test -race -coverpkg=./...` verifies package interactions and data races.
- Docker build runs the test stage inside the image build.
- Compose runtime smoke checks validate liveness, readiness, authenticated
  balance read, and Prometheus readiness.

## What Is Not Faked

- Domain tests do not fake Gin because Gin is not a domain dependency.
- Use case tests fake ports because they exercise application orchestration, not
  adapter correctness.
- Adapter tests use real in-memory adapter state because that is the current
  secondary adapter.

## Future Tests

When PostgreSQL and Redis adapters are added:

- keep domain tests unchanged;
- keep use-case tests on fakes;
- add adapter integration tests around SQL transactions, idempotency unique
  constraints, Redis TTLs, webhook outbox leasing, and key rotation lifecycle.
