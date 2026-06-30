# ADR 0007: Publish the Repository Under the MIT License

## Status

Accepted.

## Context

BankPort is already a public partner-API asset with rate-limit behavior,
idempotency guarantees, docs, benchmarks, and a demo deployment path. Without
an explicit license, the repository can be studied but its reuse boundary stays
legally ambiguous.

## Options considered

1. Keep the default all-rights-reserved posture
2. Publish under the MIT License
3. Delay licensing until PostgreSQL, Redis, and durable webhook delivery land

## Decision

Publish the repository under the MIT License and document that in the README.

## Consequences

Positive:

- Reviewers and learners can adapt the partner API patterns with a standard
  permissive license.
- The public service demo and the public legal contract now match.

Negative:

- Downstream forks may reuse only the API surface and skip the runbook caveats.
- Dependency and third-party terms still require independent review.
