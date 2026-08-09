# OpsFlow — System Architecture

## 1. Architecture Principles

1. API Gateway is the public entry point.
2. Internal services communicate through Kubernetes DNS and explicit APIs.
3. RabbitMQ is used for asynchronous workflows and domain events.
4. Redis is used for cache, rate limiting and short-lived state.
5. PostgreSQL is the transactional source of truth.
6. Elasticsearch is for operational search and indexed documents/log metadata.
7. MinIO is used for object storage.
8. AI access is isolated behind an AI Gateway.
9. AI tools are permission-aware.
10. Production-changing actions require human approval.
11. Services should be stateless unless state is explicitly required.
12. Every important operation has a correlation/request ID.
13. Infrastructure must be reproducible.
14. Local development must work without paid infrastructure.
15. Any component that issues outbound requests to a stored or user-influenced
    target (health checkers, AI tool executors) is an SSRF boundary and must be
    designed as one — see § 8a.

## 2. High-Level Architecture

```text
                           Internet / LAN
                                |
                                v
                       +-------------------+
                       | Ingress / LB      |
                       | Traefik or Kong   |
                       +---------+---------+
                                 |
                                 v
                       +-------------------+
                       |    API Gateway    |
                       | Auth / Routing    |
                       | Rate Limit / CORS|
                       +---------+---------+
                                 |
       +-------------------------+---------------------------+
       |                         |                           |
       v                         v                           v
+--------------+         +---------------+          +----------------+
| Auth Service |         | Ops Services  |          |  AI Gateway    |
+--------------+         +-------+-------+          +-------+--------+
                                |                           |
              +-----------------+-------------+             |
              |               |               |             |
              v               v               v             v
        Incident Service  Service Registry  Release    Model Router
                                                        /    |    \
                                                       /     |     \
                                                    OpenAI  Gemini Ollama

                         Data / Messaging
              +-------------+---------+-------------+
              |             |         |             |
              v             v         v             v
          PostgreSQL      Redis    RabbitMQ        MinIO
                                      |
                              +-------+-------+
                              |       |       |
                              v       v       v
                           Workers  Audit  Notify

                         Search / Observability
                              |
                    +---------+----------+
                    |                    |
                    v                    v
              Elasticsearch           Metrics
                    |                    |
                  Kibana          Prometheus/Grafana

                         Secrets
                              |
                      HashiCorp Vault
                              |
                    Dynamic DB creds / KV / PKI

                         Platform
                              |
                         Kubernetes
                              |
                    HPA / Deployment / Service
                              |
                    Terraform (provisioning)
                              |
                         Jenkins CI/CD
```

## 3. Service Boundaries

### API Gateway

Responsibilities:

- external routing
- JWT verification
- rate limiting
- CORS
- request ID
- API version routing
- basic policy enforcement

It must not contain business logic.

### Auth Service

Responsibilities:

- authentication
- credential verification
- token issuance
- refresh token lifecycle
- user identity

### User Service

Responsibilities:

- users
- roles
- permissions
- team membership

### Service Registry

Responsibilities:

- service metadata
- ownership
- environment
- dependencies
- health configuration

### Incident Service

Responsibilities:

- incident lifecycle
- severity
- assignment
- timeline
- RCA metadata

### Release Service

Responsibilities:

- deployment records
- release evidence
- validation results
- commit/image metadata

### Notification Service

Consumes events and sends notifications.

### Search Service

Owns indexing workflows and search abstraction.

### AI Gateway

Responsibilities:

- model abstraction
- model routing
- prompt/context handling
- tool execution orchestration
- usage tracking
- cost/rate governance (per-conversation and per-user budgets, see `rules.md` § 12)
- safety policies, including prompt-injection defense (`rules.md` § 12)

## 4. Communication

### Synchronous

REST/HTTP for:

- user-facing queries
- CRUD
- authentication
- immediate validation

### Asynchronous

RabbitMQ for:

- notification jobs
- search indexing
- evidence processing
- incident events
- deployment events

## 5. Event Examples

```text
IncidentCreated
IncidentResolved
DeploymentCreated
DeploymentValidated
ServiceHealthChanged
EvidenceUploaded
RCACompleted
NotificationRequested
```

Consumers must be idempotent.

## 6. API Gateway Routing

```text
/api/v1/auth/*          -> auth-service
/api/v1/users/*         -> user-service
/api/v1/services/*      -> service-registry
/api/v1/incidents/*     -> incident-service
/api/v1/releases/*      -> release-service
/api/v1/search/*        -> search-service
/api/v1/ai/*            -> ai-gateway
```

## 7. Authentication Flow

```text
PWA
 |
 | credentials
 v
Auth Service
 |
 | access + refresh token
 v
PWA
 |
 | Bearer JWT
 v
API Gateway
 |
 | validated identity
 v
Internal Service
```

JWT validation must include:

- signature
- issuer
- audience
- expiry
- not-before when used

## 8. AI Architecture

```text
PWA
 |
 v
API Gateway
 |
 v
AI Gateway
 |
 +-- Conversation Manager
 |
 +-- Model Router
 |
 +-- Tool Registry
 |
 +-- RAG Retriever
 |
 +-- Guardrails (injection defense, schema validation, cost caps)
 |
 +-- Usage Tracker
 |
 +----> OpenAI
 +----> Gemini
 +----> Ollama
```

### AI tools

Read-only:

- get_service
- get_service_health
- get_dependencies
- get_recent_deployments
- get_incident
- search_logs
- query_metrics
- search_runbooks
- search_previous_incidents

Mutation:

- create_incident
- update_incident
- request_rollback
- request_scale

Mutation tools require explicit approval.

### 8a. Outbound request boundary (SSRF)

Any tool or worker that resolves a URL/host from stored or user-influenced data —
`get_service_health`, `query_metrics`, and the health check worker in particular —
must not fetch that target directly on trust. Required controls:

- destination allowlist scoped to registered service environments; reject anything
  not in the service registry
- block requests to link-local, loopback, and internal metadata ranges
  (`169.254.0.0/16`, `127.0.0.0/8`, cloud metadata IPs) regardless of allowlist state
- DNS resolution pinned/re-validated at request time to prevent TOCTOU rebinding
- outbound calls made from a network-isolated egress path, not the same identity
  used for internal service-to-service calls

This applies whether the target was entered by an admin registering a service or
surfaced indirectly through AI tool arguments — the trust boundary is the same
either way.

## 9. RAG

Knowledge sources:

- runbooks
- ADRs
- RCA documents
- architecture documents
- troubleshooting guides

Pipeline:

```text
Document
 -> parse
 -> chunk
 -> embed
 -> index
 -> retrieve
 -> rerank if needed
 -> provide evidence to model
```

Do not use raw production secrets or unrestricted logs as RAG documents.

Retrieved content is data, not instruction — see `rules.md` § 12 (Injection defense).

## 10. Kubernetes

Each stateless service should use:

- Deployment
- Service
- ConfigMap where appropriate
- Secret references
- readiness probe
- liveness probe
- resource requests/limits

At least one service should demonstrate HPA.

## 11. CI/CD

```text
GitHub
  |
  v
Jenkins
  |
  +-- lint
  +-- unit tests
  +-- integration tests
  +-- security scan
  +-- Docker build
  +-- image scan
  +-- push registry
  +-- deploy
  +-- smoke test
```

## 12. Observability

Every service should provide:

- `/health`
- `/ready`
- structured JSON logs
- Prometheus metrics
- correlation ID

Recommended telemetry:

- request count
- request latency
- error count
- queue depth
- cache hit ratio
- DB latency
- AI request count
- AI latency
- token usage
- AI cost per conversation/user (see `rules.md` § 12 Cost governance)

## 13. Failure Behaviour

### Redis unavailable

Application should degrade to database access where safe. Cache failure must not destroy transactional correctness.

### RabbitMQ unavailable

API should fail fast for operations that require asynchronous processing or persist an outbox event for later delivery.

### AI provider unavailable

AI Gateway should return a controlled provider error and optionally route to another configured provider.

### Elasticsearch unavailable

Core transactional workflows must continue. Search can be temporarily degraded.

### Vault unavailable

Services with already-issued short-lived credentials continue until expiry.
New credential requests fail closed — do not fall back to a cached or default
secret. This is a deliberate availability/security trade-off.

## 14. Deployment Modes

### Local

Docker Compose.

### Kubernetes local

kind or k3d.

### Homelab

Kubernetes + local registry/object storage.

### Provisioning

Terraform provisions the Kubernetes cluster resources, namespaces, and supporting
infra (registry, object storage buckets) declaratively, even in local/homelab mode
— e.g. via the `kind` or `helm` Terraform providers. This keeps "reproducible
infrastructure" (§1.13) enforced by tooling rather than by discipline alone, and
is the same practice that scales to a real cloud target later without a rewrite.

No paid cloud service is required for the baseline project.

## 15. Secrets Management

Static environment variables are acceptable for local development only.

For homelab/Kubernetes deployment, secrets are issued by HashiCorp Vault:

- KV engine for static application secrets
- database secrets engine for short-lived, per-service PostgreSQL credentials
- PKI engine for internal service certificates where mTLS is introduced

The AI Gateway's provider API keys are stored in Vault and injected at runtime —
never baked into images or committed to Git (see `rules.md` § 11).

This is optional for the MVP milestone but is the natural next step once
Kubernetes Secrets stop being sufficient, and keeps the project's infra stack
internally consistent with HashiCorp tooling rather than mixing secret stores
ad hoc.