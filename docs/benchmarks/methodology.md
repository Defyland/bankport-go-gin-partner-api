# Benchmark Methodology

Benchmarks must prove behavior under load, not just produce a large number.

## Dataset

The sandbox seed contains two accounts for `partner_sandbox_bank` and one account
for a second partner used for BOLA tests. Financial write benchmarks use
`acct_sandbox_001` with one-cent commands to avoid exhausting the balance before
the scenario completes.

## Profiles

- Smoke: validates the deployed API, auth, idempotency, and JSON contract.
- Load: checks steady read-heavy usage with occasional financial writes.
- Stress: identifies when rate limits become the active protection.
- Spike: validates sudden partner traffic without process crashes.

## Metrics

Record p50, p95, p99, throughput, error rate, CPU, memory, rate-limit count,
idempotency conflicts, and queued webhook deliveries. Prometheus queries live in
the Grafana dashboard under `deployments/grafana/dashboards/`.

## Acceptance threshold

For the sandbox MVP, p95 below 250 ms during the load profile and error rate
below 2% are acceptable. Rate-limited responses are not counted as product
failures during stress and spike profiles when they are expected by the test.
