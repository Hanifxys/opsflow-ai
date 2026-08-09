# OpsFlow — Portfolio Demonstration Guide

Welcome to the **OpsFlow** demonstration guide. This document outlines the end-to-end operational scenario showcasing the architecture, microservices topology, outbox event-driven messaging, AI human-in-the-loop guardrails, and observability.

---

## 1. System Topology Overview

```text
                                [ User / Web Frontend ]
                                           │
                                           ▼
                                 [ API Gateway (:8080) ]
       ┌───────────────────┬───────────────┴───────────────┬──────────────────┐
       ▼                   ▼                               ▼                  ▼
[ Auth Service ]  [ Service Registry ]            [ Incident Service ]  [ AI Gateway ]
   (:8081)             (:8083)                          (:8082)            (:8084)
      │                   │                                │                  │
      ▼                   ▼                                ▼                  ▼
[ PostgreSQL ]      [ PostgreSQL ]                   [ PostgreSQL ]    [ Model Router ]
                                                      (Outbox Table)   (Mock/Ollama/Cloud)
                                                           │                  │
                                                           ▼                  ▼
                                                      [ RabbitMQ ] ──► [ Notification ]
                                                      (opsflow.events)     Worker (:8085)
```

---

## 2. Demonstration Scenario: Automated Incident & AI Mitigation

### Step 1: Authentication & Identity (`/api/v1/auth`)
- **Action**: Authenticate as an Operator.
- **Request**: `POST /api/v1/auth/login`
- **Output**: Returns signed JWT Access Token containing `X-User-ID`, `X-User-Email`, and permissions (`incident:write`, `ai:use`, `ai:execute`).

### Step 2: Declare Service Incident (`/api/v1/incidents`)
- **Action**: Declare a critical incident for `payment-service`.
- **Request**: `POST /api/v1/incidents`
  ```json
  {
    "service_id": "c0eebc99-9c0b-4ef8-bb6d-6bb9bd380c11",
    "title": "Payment Database Connection Latency",
    "description": "Connection pool size limit hit, requests failing with HTTP 504",
    "severity": "CRITICAL"
  }
  ```
- **System Event**:
  1. Incident saved in PostgreSQL.
  2. Transactionally inserts `INCIDENT_CREATED` event into `outbox_events` table.
  3. Background publisher picks up outbox event and publishes to RabbitMQ topic exchange `opsflow.events`.
  4. Notification Worker consumes event, checks SHA256 idempotency key in `notifications` table, and dispatches notification.

### Step 3: AI Assistant Mitigation & Human Approval (`/api/v1/ai`)
- **Action**: Operator opens AI Session and asks AI for assistance.
- **Request**: `POST /api/v1/ai/conversations/{id}/messages`
  ```json
  {
    "content": "Payment service is failing. Please restart payment-service in production."
  }
  ```
- **AI Safety Response**:
  1. AI Model Router analyzes request.
  2. Recognizes sensitive operational mutation tool `restart_service`.
  3. Generates a **PENDING Approval Request** (`ai_approvals` table).
  4. **Zero Autonomous Mutation**: Operational mutation is held in check.

### Step 4: Human-in-the-Loop Execution (`/api/v1/ai/approvals`)
- **Action**: Senior Engineer reviews and approves action.
- **Request**: `POST /api/v1/ai/approvals/{id}/approve`
- **System Output**: Action approved, tool mutation executes, and audit trail record is persisted.

### Step 5: Unified Search & Observability
- **Unified Search**: `GET /api/v1/search?q=payment` aggregates incidents and service registry entries.
- **Prometheus Metrics**: `GET /metrics` displays `http_requests_total` and latency histograms.
- **Distributed Tracing**: W3C `traceparent` header propagated across Gateway -> Microservices.

---

## 3. Running Locally

```bash
# 1. Start Infrastructure (PostgreSQL, Redis, RabbitMQ)
docker compose up -d

# 2. Run All Tests
make test
```
