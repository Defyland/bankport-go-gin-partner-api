# BankPort Benchmark Baseline

## Scope

The benchmark package covers four public traffic profiles:

- smoke: one virtual user exercising balance reads and Pix writes
- load: mixed reads and low-rate writes at a steady partner workload
- stress: read-heavy growth until rate limits become the expected control
- spike: sudden sandbox reads to prove graceful throttling

## Commands

```bash
k6 run benchmarks/k6-smoke.js
k6 run benchmarks/k6-load.js
k6 run benchmarks/k6-stress.js
k6 run benchmarks/k6-spike.js
go test -bench=. -benchmem ./internal/httpapi
```

## Current measured baseline

The Go native benchmark was measured locally on 2026-05-30 using Go 1.26.3
darwin/arm64 against the in-process Gin router.

| Scenario | p50 | p95 | p99 | Throughput | Error rate | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| HTTP balance read loop | 0.341 ms | 0.534 ms | 0.740 ms | 2501.52 rps | 0.0% | Local loopback API on port 18080. |
| `BenchmarkGetBalanceRequest` | n/a | n/a | n/a | 10566 ns/op | 0.0% | In-process benchmark: 10885 B/op, 85 allocs/op. |

## Bottleneck hypothesis

The first real bottleneck is not Gin routing. It is the shared consistency path
for financial writes: account balance mutation, idempotency lookup, outbox event
write, and webhook delivery enqueue. In production this path must run inside a
single PostgreSQL transaction with a unique idempotency constraint and optimistic
account version checks.

## Next optimization

Add a PostgreSQL-backed repository benchmark once the in-memory sandbox is
replaced by the production persistence adapter. Measure account-row contention
under repeated writes to the same partner account and compare it with sharded
accounts or per-account queueing.
