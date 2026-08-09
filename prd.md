# OpsFlow — Product Requirements Document

## 1. Product Overview

**OpsFlow** is a self-hostable, cloud-native operations platform for engineering and DevOps teams.

It provides a single workspace to manage:

- services and ownership
- incidents
- deployments and release evidence
- operational health
- runbooks and knowledge
- asynchronous operational jobs
- AI-assisted investigation and analysis

The platform is designed as a portfolio-grade production-style system, not as a demo CRUD application.

## 2. Problem Statement

Engineering teams often operate across disconnected systems:

- source control
- CI/CD
- monitoring
- logs
- incident records
- runbooks
- deployment history
- chat
- infrastructure

This creates fragmented operational context. During an incident, engineers spend time collecting evidence before they can investigate.

OpsFlow centralises the operational context and provides an AI Copilot that can retrieve approved operational data and recommend next actions.

## 3. Goals

### Primary goals

1. Provide one operational dashboard.
2. Track service ownership and dependencies.
3. Manage incidents from detection through RCA.
4. Record deployment and release evidence.
5. Provide health and synthetic checks.
6. Support asynchronous workflows through RabbitMQ.
7. Provide search through Elasticsearch.
8. Provide caching and rate limiting through Redis.
9. Run as containers and deploy to Kubernetes.
10. Provide CI/CD through Jenkins.
11. Provide model-agnostic AI through an AI Gateway.
12. Support RAG, tool calling and human approval for sensitive actions.
13. Be runnable locally using free/open-source software.

### Non-goals for MVP

- full Jira replacement
- full observability platform replacement
- automatic production remediation
- financial/accounting features
- native mobile application
- multi-cloud orchestration

## 4. Target Users

### Engineer

Needs service, deployment, incident and diagnostic context.

### Operations / SRE

Needs health, incidents, dependencies, alerts and operational evidence.

### Engineering Lead

Needs service ownership, release readiness, reliability trends and incident summaries.

### Administrator

Manages users, roles, permissions, model providers and system configuration.

## 5. Core User Journeys

### 5.1 Service registration

1. User creates a service.
2. Adds owner/team.
3. Adds repository and environments.
4. Defines health checks.
5. Defines dependencies.
6. Service becomes visible on the operations dashboard.

### 5.2 Incident management

1. Incident is created manually or through an integration.
2. Severity and affected service are recorded.
3. Incident enters an explicit state machine.
4. Timeline events are recorded.
5. Evidence is attached.
6. Incident is resolved.
7. RCA is created.
8. Knowledge can be linked to future investigations.

### 5.3 Deployment validation

1. CI pipeline creates deployment evidence.
2. OpsFlow records commit, image, environment and timestamp.
3. Post-deployment checks run.
4. Results are stored.
5. Dashboard shows deployment health.

### 5.4 AI investigation

User asks:

> Why is payment-service unhealthy?

AI:

1. identifies the service
2. retrieves service metadata
3. queries health
4. checks recent deployments
5. queries approved logs/metrics
6. searches runbooks and previous incidents
7. produces an evidence-backed answer
8. provides recommended next steps
9. asks for approval before sensitive actions

## 6. Functional Requirements

### Authentication

- login
- logout
- refresh token
- password hashing
- JWT access token
- role-based access control
- permission checks

### Service Management

- CRUD services
- ownership
- environment mapping
- dependency mapping
- health checks
- service criticality
- repository metadata

### Incident Management

- create/update incidents
- severity
- status
- assignment
- timeline
- comments
- evidence
- RCA
- audit trail

### Deployment

- deployment registration
- commit SHA
- image tag
- environment
- deployment status
- validation results
- rollback metadata

### Health

- HTTP health checks
- synthetic transactions
- response time
- availability history
- check status

### Search

- incident search
- runbook search
- RCA search
- log search metadata
- semantic knowledge retrieval

### Notifications

- asynchronous notifications
- incident notifications
- deployment notifications
- configurable channels

### AI

- model selection
- provider abstraction
- conversation history
- tool calling
- RAG
- evidence references
- token/usage tracking
- permission-aware tools
- human approval for mutations

## 7. Non-Functional Requirements

### Availability

Internal target: 99.9% for the platform when deployed with redundant components.

### Performance

- normal API p95 < 500 ms excluding long-running AI requests
- health checks should not block user requests
- asynchronous jobs must be queued

### Security

- HTTPS in deployed environments
- JWT validation
- RBAC
- least privilege
- secret management
- input validation
- rate limiting
- audit logging
- dependency and image scanning

### Scalability

Services must be horizontally scalable where stateless.

### Observability

All services should expose:

- structured logs
- health endpoint
- metrics
- trace/request correlation

## 8. MVP Scope

### Must have

- PWA
- API Gateway
- Auth Service
- User/RBAC
- Service Registry
- Incident Service
- PostgreSQL
- Redis
- RabbitMQ
- Docker
- Kubernetes
- Jenkins
- Prometheus/Grafana
- basic AI Gateway
- one cloud model provider
- Ollama local provider

### Later

- Elasticsearch/Kibana
- advanced RAG
- MCP server
- synthetic transaction engine
- deployment risk scoring
- automatic anomaly correlation
- advanced integrations

## 9. Success Criteria

A complete MVP must demonstrate:

1. User can log in.
2. User can register a service.
3. User can create and resolve an incident.
4. Incident events are persisted.
5. Notification work is processed asynchronously.
6. Redis caching/rate limiting works.
7. Application runs in Docker.
8. Application deploys to Kubernetes.
9. Jenkins can build, test and deploy.
10. HPA can scale at least one stateless service.
11. AI can answer operational questions using tools.
12. AI can switch between local and remote models.
13. Sensitive actions require human approval.
14. Monitoring shows application health.
