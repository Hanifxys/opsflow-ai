# OpsFlow — Implementation Plan

## Phase 0 — Repository Foundation

Deliver:

- repository structure
- rules
- local development documentation
- Docker Compose skeleton
- Makefile/task runner
- environment template
- linting
- basic CI

Exit criteria:

- project builds
- tests run
- local dependencies start

## Phase 1 — Core API

Build:

- API Gateway
- Auth Service
- User/RBAC
- Incident Service
- PostgreSQL

Exit criteria:

- login works
- JWT validation works
- incident CRUD works
- RBAC works
- integration tests pass

## Phase 2 — Service Registry

Build:

- service CRUD
- environments
- dependencies
- health check configuration

Exit criteria:

- service topology can be represented
- service ownership works
- API documented with OpenAPI

## Phase 3 — Async Platform

Build:

- RabbitMQ
- notification worker
- outbox pattern
- event consumers
- retry/DLQ

Exit criteria:

- incident event reaches worker
- duplicate delivery does not duplicate side effects
- failed messages reach DLQ

## Phase 4 — Redis

Build:

- API rate limiting
- selected read cache
- cache invalidation

Exit criteria:

- rate limit works
- cache failure does not break transactional correctness

## Phase 5 — Search and Knowledge

Build:

- Elasticsearch
- search service
- runbook documents
- RCA indexing
- knowledge ingestion pipeline

Exit criteria:

- incident/runbook search works
- indexing is asynchronous
- transactional data remains in PostgreSQL

## Phase 6 — Observability

Build:

- Prometheus
- Grafana
- structured logs
- OpenTelemetry
- request correlation
- service dashboards

Exit criteria:

- request latency visible
- errors visible
- queue depth visible
- service health visible

## Phase 7 — Kubernetes

Build:

- Deployments
- Services
- ConfigMaps
- Secrets
- probes
- resource requests/limits
- HPA

Exit criteria:

- complete system deploys locally to kind/k3d
- one service scales horizontally
- failed pod is recreated

## Phase 8 — Jenkins CI/CD

Build:

- lint
- unit tests
- integration tests
- security scan
- Docker build
- image scan
- registry push
- Kubernetes deployment
- smoke tests

Exit criteria:

- commit to deployment path is reproducible

## Phase 9 — AI Gateway

Build:

- model provider interface
- Ollama provider
- one cloud provider
- streaming
- conversation persistence
- model switching

Exit criteria:

- user can change provider/model
- application services remain provider agnostic

## Phase 10 — AI Tools

Build read-only tools:

- get_service
- get_service_health
- get_dependencies
- get_recent_deployments
- get_incident
- search_logs
- query_metrics
- search_runbooks
- search_previous_incidents

Exit criteria:

- AI can investigate using real application data
- tool permissions are enforced
- tool calls are audited

## Phase 11 — RAG

Build:

- document ingestion
- chunking
- embeddings
- retrieval
- evidence references

Exit criteria:

- AI can answer questions using approved runbooks/RCA/architecture documents
- retrieved evidence is visible

## Phase 12 — Human Approval

Build:

- AI action proposals
- approval requests
- expiry
- execution audit

Initial mutation tools:

- create incident
- update incident
- request scale
- request rollback

No automatic production mutation.

## Phase 13 — PWA

Build:

- installable PWA
- dashboard
- incident workspace
- service workspace
- deployment workspace
- AI chat
- notifications

## Phase 14 — Reliability and Portfolio Hardening

Add:

- load testing with k6
- failure scenarios
- backup/restore
- rate-limit tests
- security tests
- dependency failure tests
- architecture diagrams
- ADRs
- runbooks
- demo environment

## Final acceptance

The complete project must demonstrate:

```text
PWA
 ↓
API Gateway
 ↓
Microservices
 ↓
PostgreSQL / Redis / RabbitMQ / MinIO
 ↓
Search / Observability
 ↓
Kubernetes
 ↓
Jenkins CI/CD
 ↓
AI Gateway
 ↓
RAG + Tools + Human Approval
```
