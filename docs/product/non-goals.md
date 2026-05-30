# Non-goals

The first production-shaped slice intentionally does not include:

- real core-banking or Pix provider integration
- OAuth authorization-code flows
- mTLS enforcement
- asynchronous worker deployment
- multi-region active-active writes
- partner self-service UI
- Kubernetes manifests
- event sourcing as the system of record

These are deferred because the public partner contract, middleware controls,
domain invariants, and observability need to be correct before adding more
moving parts.
