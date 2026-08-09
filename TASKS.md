# OpsFlow — Execution Backlog

Status values:

- `[ ]` todo
- `[~]` in progress
- `[x]` done
- `[!]` blocked

## Epic 0 — Foundation

- [x] T001 initialise repository
- [x] T002 create Go workspace/module structure
- [x] T003 create frontend PWA shell
- [x] T004 create Docker Compose development environment
- [x] T005 add Makefile/task commands
- [x] T006 add lint/static analysis
- [x] T007 add baseline Jenkins pipeline

## Epic 1 — Authentication

- [x] T101 create auth service
- [x] T102 implement password hashing
- [x] T103 implement login
- [x] T104 implement JWT access token
- [x] T105 implement refresh token
- [x] T106 implement RBAC
- [x] T107 add auth integration tests

## Epic 2 — API Gateway

- [x] T201 configure gateway
- [x] T202 configure `/api/v1`
- [x] T203 JWT validation
- [x] T204 rate limiting
- [x] T205 CORS
- [x] T206 request/correlation ID
- [x] T207 gateway integration tests

## Epic 3 — Incident

- [x] T301 incident schema
- [x] T302 incident domain model
- [x] T303 create incident API
- [x] T304 incident state machine
- [x] T305 incident timeline
- [x] T306 incident comments
- [x] T307 incident resolution
- [x] T308 incident tests

## Epic 4 — Service Registry

- [x] T401 service schema
- [x] T402 service CRUD
- [x] T403 environments
- [x] T404 dependencies
- [x] T405 health checks
- [x] T406 registry API tests

## Epic 5 — Async

- [x] T501 configure RabbitMQ
- [x] T502 create event contracts
- [x] T503 implement outbox
- [x] T504 notification worker
- [x] T505 retries
- [x] T506 dead-letter queues
- [x] T507 idempotent consumers

## Epic 6 — AI

- [x] T601 define LLMProvider interface
- [x] T602 implement Ollama
- [x] T603 implement cloud provider
- [x] T604 model router
- [x] T605 AI conversation API
- [x] T606 streaming
- [x] T607 tool registry
- [x] T608 read-only operational tools
- [x] T609 tool audit
- [x] T610 approval workflow
- [x] T611 RAG ingestion
- [x] T612 RAG retrieval

## Epic 7 — Platform

- [x] T701 Kubernetes manifests
- [x] T702 probes
- [x] T703 resource limits
- [x] T704 HPA
- [x] T705 Prometheus
- [x] T706 Grafana
- [x] T707 OpenTelemetry
- [x] T708 Elasticsearch / Unified Search
- [x] T709 MinIO
- [x] T710 backup/restore

## Epic 8 — Quality

- [x] T801 integration test suite
- [x] T802 E2E suite
- [x] T803 k6 load test
- [x] T804 security scan
- [x] T805 dependency scan
- [x] T806 failure testing
- [x] T807 documentation review
- [x] T808 portfolio demo scenario

## Execution rule

Implement tasks in dependency order.

Do not mark a task done because code exists. It is done only when its acceptance criteria pass.
