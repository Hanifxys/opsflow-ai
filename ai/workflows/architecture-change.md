# AI Architecture Change Workflow

Architecture changes require explicit review.

```text
Problem
 ↓
Current limitation
 ↓
Options
 ↓
Trade-offs
 ↓
Recommendation
 ↓
ADR
 ↓
Architecture update
 ↓
Design update
 ↓
Implementation plan
 ↓
Code
```

The AI must not silently replace:

- RabbitMQ
- Redis
- PostgreSQL
- Elasticsearch
- Kubernetes
- API Gateway
- model provider abstraction

with alternatives.

Any replacement requires an ADR.
