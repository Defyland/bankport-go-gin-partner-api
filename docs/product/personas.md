# Personas

## Partner developer

Integrates BankPort into marketplaces, ERPs, payment operations, or treasury
tools. Needs stable OpenAPI docs, deterministic sandbox data, clear errors,
idempotency, and webhook signatures.

## Partner operations analyst

Investigates failed payouts, duplicate requests, webhook delivery issues, and
rate-limit events. Needs correlation IDs, audit logs, replay evidence, and
runbooks.

## BankPort platform engineer

Owns API safety, rate limits, scope enforcement, observability, and backward
compatibility. Needs tests that catch BOLA, idempotency conflicts, and contract
drift before deploy.

## Security reviewer

Reviews tenant isolation, credential handling, webhook signing, secret
management, and abuse controls.
