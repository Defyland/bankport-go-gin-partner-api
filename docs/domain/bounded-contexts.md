# Bounded Contexts

## Partner access

Owns API keys, developer apps, scopes, and rate limits. It decides whether a
request can reach product functionality.

## Account exposure

Owns read models for balances and statements. It hides foreign partner accounts
by returning 404.

## Financial commands

Owns Pix transfer, payout, and refund commands. It validates money invariants,
applies account mutations, and emits events.

## Webhooks

Owns endpoint registration, event matching, delivery signatures, and queued
delivery evidence.

## Audit and observability

Owns request IDs, correlation IDs, structured logs, domain metrics, and partner
audit entries.
