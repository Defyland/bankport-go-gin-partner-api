# ADR 0001: Use Gin as the HTTP Framework

## Status

Accepted.

## Context

BankPort is an API-first partner platform where the framework itself is part of the engineering story. The project needs explicit route groups, middleware chains, JSON binding, validation, recovery, versioned APIs, and a clear developer experience for enterprise-style integrations.

## Decision

BankPort uses Gin as the HTTP framework. Domain logic remains framework-independent behind internal packages, but request routing, middleware composition, and transport concerns are implemented through Gin.

## Consequences

- The repository can demonstrate professional Go API construction with a recognizable framework.
- Middleware concerns such as request IDs, correlation IDs, auth, scopes, rate limits, idempotency, and metrics stay visible and testable.
- Domain packages remain decoupled from Gin-specific types.
