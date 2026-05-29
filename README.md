# BankPort

Partner API platform built in Go with Gin to showcase enterprise-grade financial API exposure, middleware-driven policy enforcement, and a developer-facing BaaS integration surface.

## Status

Phase 0 bootstrap only. This repository currently establishes naming, scope, documentation structure, and engineering expectations. It does not yet contain Gin handlers, middlewares, Redis-backed rate limiting, or internal financial adapters.

## Product intent

BankPort is planned as a partner gateway for financial APIs. It exposes controlled access to balances, statements, Pix transfers, payouts, refunds, webhooks, logs, and sandbox scenarios through authenticated, scoped, rate-limited, idempotent APIs for external partners.

## Planned stack

- Go
- Gin
- PostgreSQL
- Redis
- OpenAPI
- OpenTelemetry
- Prometheus and Grafana
- Docker Compose
- k6
- RabbitMQ or NATS as a later-phase webhook delivery backend

## Engineering focus

This project is meant to demonstrate:

- explicit use of Gin route groups and middleware chains in a serious API-first product
- API-key and client-credential style partner authentication with scopes and policies
- rate limiting, idempotency, and standardized error envelopes for financial write APIs
- request logging, usage metering, auditability, and partner-oriented developer experience
- fake adapters that protect internal systems while exposing a stable partner API boundary
- sandbox scenarios and webhook delivery workflows as first-class integration features

## Bootstrap contents

- repository initialized and synchronized with GitHub
- mandatory documentation folders created, including `docs/events/` and `docs/security/`
- baseline engineering spec captured in `docs/engineering-baseline.md`
- partner API event contracts documented in `docs/events/README.md`
- threat model documented in `docs/security/threat-model.md`
- deployment readiness documented in `docs/architecture/deployment-readiness.md`

## Next phase

The first implementation slice should prioritize Gin route groups, core middleware, partner auth, scopes, rate limits, idempotency, fake partner APIs, webhook endpoint registration, and OpenAPI-backed contract coverage.
