# OpsFlow — Detailed Technical Design

## 1. Design Goals

The design must be:

- modular
- testable
- observable
- secure by default
- locally reproducible
- Kubernetes-ready
- model-provider agnostic
- understandable by a single developer

## 2. Recommended Backend Structure

Use Go for backend services.

```text
service/
├── cmd/
│   └── server/
├── internal/
│   ├── domain/
│   ├── application/
│   ├── ports/
│   ├── adapters/
│   │   ├── http/
│   │   ├── postgres/
│   │   ├── redis/
│   │   └── rabbitmq/
│   └── config/
├── migrations/
├── tests/
└── Dockerfile
```

Use dependency inversion. Domain logic must not depend directly on infrastructure libraries.

## 3. API Design

Use REST with JSON.

Example:

```http
POST /api/v1/incidents
GET /api/v1/incidents
GET /api/v1/incidents/{id}
PATCH /api/v1/incidents/{id}
POST /api/v1/incidents/{id}/events
POST /api/v1/incidents/{id}/resolve
```

Responses should use a consistent envelope:

```json
{
  "data": {},
  "meta": {
    "request_id": "req-123"
  }
}
```

Errors:

```json
{
  "error": {
    "code": "INCIDENT_NOT_FOUND",
    "message": "Incident was not found",
    "request_id": "req-123"
  }
}
```

Never expose internal stack traces to clients.

## 4. Incident State Machine

```text
OPEN
 |
 v
INVESTIGATING
 |
 v
MITIGATING
 |
 v
RESOLVED
 |
 v
CLOSED
```

Allowed transitions must be explicit.

Invalid transitions return a domain error.

## 5. Service Health Design

A health check contains:

- service ID
- endpoint
- method
- expected status
- timeout
- interval
- active flag

Worker executes checks asynchronously.

Result:

- status
- latency
- timestamp
- error classification

Consecutive failures can emit `ServiceHealthChanged`.

## 6. RabbitMQ Design

Use topic exchanges where event routing is required.

Example:

```text
opsflow.events
```

Routing keys:

```text
incident.created
incident.resolved
deployment.created
deployment.validated
service.health_changed
```

Queues:

```text
notification.worker
search.indexer
audit.worker
```

Use dead-letter queues.

Consumers must be idempotent.

## 7. Redis Design

Use Redis for:

- API rate limiting
- response cache
- short-lived refresh/session metadata
- distributed locks only when strictly required

Do not use Redis as the primary source of truth.

## 8. Elasticsearch Design

Indexes:

```text
opsflow-incidents
opsflow-runbooks
opsflow-rcas
opsflow-deployments
opsflow-logs
```

Indexing should be asynchronous.

Transactional write:

```text
PostgreSQL
   |
   v
Outbox/Event
   |
   v
RabbitMQ
   |
   v
Search Indexer
   |
   v
Elasticsearch
```

## 9. Outbox Pattern

For important events:

```text
BEGIN
  update business data
  insert outbox event
COMMIT
```

A publisher worker reads unpublished events and publishes them to RabbitMQ.

This prevents a database update succeeding while event publication fails.

## 10. AI Model Abstraction

Interface:

```go
type LLMProvider interface {
    Chat(ctx context.Context, request ChatRequest) (ChatResponse, error)
    Stream(ctx context.Context, request ChatRequest) (<-chan Token, error)
}
```

Providers:

```text
providers/
├── openai/
├── gemini/
└── ollama/
```

The AI Gateway owns provider-specific implementation.

Application services must not call provider SDKs directly.

## 11. AI Model Routing

Model configuration:

```yaml
models:
  default: local
  providers:
    local:
      type: ollama
      model: qwen
    cloud:
      type: openai
      model: configured-by-env
```

Provider secrets must come from environment/secret management.

Never commit API keys.

## 12. AI Tool Execution

Tool flow:

```text
User
 |
 v
AI Gateway
 |
 v
Model
 |
 | tool call
 v
Tool Executor
 |
 +-- authorization
 +-- validation
 +-- execution
 +-- audit
 |
 v
Tool result
 |
 v
Model
 |
 v
Final response
```

The model never directly receives database credentials or Kubernetes credentials.

## 13. AI Approval

Sensitive operation:

```text
AI recommendation
      |
      v
Approval Request
      |
      v
Human approves
      |
      v
Execution Service
      |
      v
Audit event
```

No autonomous production mutation in MVP.

## 14. PWA Design

Frontend should support:

- installable PWA
- responsive layout
- dashboard
- incident workspace
- service registry
- AI chat
- notifications

Offline support is limited to UI shell and safe cached read data. Do not promise full offline transactional operations in MVP.

## 15. Security Design

### Authentication

- Argon2id or bcrypt
- short-lived access token
- refresh token rotation where feasible

### Authorisation

RBAC:

```text
ADMIN
OPS
ENGINEER
VIEWER
```

Permission examples:

```text
service:read
service:write
incident:read
incident:create
incident:update
incident:resolve
release:read
release:write
ai:use
ai:execute
```

### Security controls

- TLS
- CORS allowlist
- rate limiting
- input validation
- SQL parameterisation
- secure headers
- secret isolation
- audit logging
- image scanning
- dependency scanning

## 16. Testing Design

### Unit

Domain/application logic.

### Integration

PostgreSQL, Redis and RabbitMQ using containers.

### Contract

OpenAPI validation and service contract tests.

### E2E

PWA -> Gateway -> service -> database.

### Load

k6 scenarios:

- normal traffic
- burst traffic
- AI request load
- incident creation load

## 17. Deployment Design

Each service has:

- Dockerfile
- Kubernetes Deployment
- Service
- ConfigMap/Secret references
- probes
- resources

Use Helm only after plain Kubernetes manifests are understood.

## 18. Recovery

Baseline:

- PostgreSQL backup
- MinIO backup
- configuration stored in Git
- reproducible Kubernetes manifests
- RabbitMQ durable queues

Document:

- backup frequency
- restore procedure
- RPO
- RTO

## 19. Architectural Decision Records

Initial ADRs should cover:

- why microservices
- why REST
- why RabbitMQ
- why Redis
- why PostgreSQL
- why Elasticsearch
- why Kubernetes
- why model abstraction
- why human approval for AI mutations
- why modular monolith is acceptable for early development

## 20. Engineering Rule

Complexity must be justified by a requirement.

If a component does not solve a defined problem, do not add it merely for portfolio value.
