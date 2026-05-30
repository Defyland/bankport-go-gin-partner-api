# C4 Context

```mermaid
flowchart LR
  PartnerDeveloper["Partner developer"] --> BankPort["BankPort Partner API"]
  PartnerOps["Partner operations analyst"] --> BankPort
  BankPort --> PartnerWebhook["Partner webhook receiver"]
  BankPort --> CoreBanking["Internal banking adapters (deferred)"]
  BankPort --> Observability["Prometheus, Grafana, traces, logs"]
```

BankPort is the public boundary between external partners and internal financial
systems. The current implementation keeps provider adapters fake while making
the public API controls concrete and testable.
