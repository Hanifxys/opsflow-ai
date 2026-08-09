# OpsFlow — AI-Driven Development Workflow

OpsFlow is developed through a specification-driven, AI-assisted workflow.

The goal is not to let AI freely generate the whole system. The goal is to give AI a controlled engineering process where every implementation decision is traceable to requirements, architecture, design and tests.

## Source of truth hierarchy

```text
User requirement
      ↓
rules.md
      ↓
prd.md
      ↓
architecture.md
      ↓
design.md
      ↓
schema.md
      ↓
ADR
      ↓
implementation-plan.md
      ↓
TASKS.md
      ↓
Code
      ↓
Tests
      ↓
CI/CD
```

## AI workflow

```text
IDEA
  ↓
PRODUCT ANALYST
  ↓
PRD
  ↓
SYSTEM ARCHITECT
  ↓
ARCHITECTURE
  ↓
TECHNICAL DESIGNER
  ↓
DATABASE/API DESIGN
  ↓
SECURITY REVIEW
  ↓
DEVOPS DESIGN
  ↓
AI ARCHITECTURE
  ↓
IMPLEMENTATION PLANNER
  ↓
CODING AGENT
  ↓
TEST AGENT
  ↓
REVIEW AGENT
  ↓
DOCUMENTATION AGENT
  ↓
CI/CD
```

## Rule

AI must not skip directly from an idea to large-scale implementation.

Each stage produces an artifact that becomes input to the next stage.

## Repository layout

```text
docs/
  product/
  architecture/
  design/
  security/
  api/
  database/
  devops/
  ai/
  testing/
  adr/

ai/
  agents/
  prompts/
  workflows/

TASKS.md
IMPLEMENTATION-PLAN.md
AGENTS.md
rules.md
prd.md
architecture.md
design.md
schema.md
```

## Getting Started

### Prerequisites

- **Go** ≥ 1.22
- **Node.js** ≥ 20 LTS
- **Docker** and **Docker Compose**
- **golangci-lint** (optional, for `make lint`)
- **Make** (optional, for task runner)

### Setup

```bash
# 1. Copy environment template
cp .env.example .env

# 2. Start infrastructure
docker compose up -d

# 3. Install frontend dependencies
cd frontend && npm ci && cd ..

# 4. Build all Go services
# Option A: Using make
make build

# Option B: Manually per service
cd services/gateway && go build ./cmd/server && cd ../..

# 5. Run a service (e.g. API Gateway)
make run-gateway

# 6. Start frontend dev server
make fe-dev
# → http://localhost:3000
```

### Useful Commands

| Command           | Description                            |
|-------------------|----------------------------------------|
| `make up`         | Start infrastructure containers        |
| `make down`       | Stop infrastructure containers         |
| `make reset`      | Reset infrastructure (remove volumes)  |
| `make build`      | Build all Go services                  |
| `make test`       | Run all Go tests                       |
| `make lint`       | Run golangci-lint                      |
| `make check`      | Pre-commit gate (vet + lint + test)    |
| `make fe-dev`     | Start frontend dev server              |
| `make fe-build`   | Build frontend for production          |
