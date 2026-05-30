# Request Flow Diagram

```mermaid
flowchart LR
  A["Partner request"] --> B["Request identity"]
  B --> C["Trace span"]
  C --> D["Auth"]
  D --> E["Rate limit"]
  E --> F["Scope check"]
  F --> G{"Financial write?"}
  G -- yes --> H["Idempotency"]
  G -- no --> I["Handler"]
  H --> I
  I --> J["Repository"]
  J --> K["Events and audit"]
  I --> L["Standard response"]
```
