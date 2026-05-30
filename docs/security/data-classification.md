# Data Classification

| Data | Classification | Controls |
| --- | --- | --- |
| API key secret | Secret | Hash at rest, rotate, never log raw value. |
| Webhook signing key | Secret | Environment variable, secret manager in production. |
| Account balance | Confidential financial data | Partner ownership check, audit logs. |
| Statement entries | Confidential financial data | Partner ownership check, scoped reads. |
| Pix key | Personal or business identifier | Minimize logs, mask in future audit exports. |
| Request ID and correlation ID | Operational metadata | Safe for logs and partner support. |
| Event payload | Confidential integration data | Signed webhook delivery, schema versioning. |

The sandbox implementation uses deterministic fake data. Production adapters
must apply the same classification to persisted and downstream data.
