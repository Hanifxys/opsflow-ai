# =============================================================================
# OpsFlow — Makefile
# =============================================================================
# Single task runner for all development operations.

.PHONY: help up down reset \
        build test test-coverage lint fmt vet tidy \
        fe-install fe-dev fe-build fe-lint \
        check run-gateway run-auth run-incident run-registry run-ai-gateway

SERVICES := auth incident registry ai-gateway gateway
SERVICE_DIRS := $(addprefix services/,$(SERVICES))
ALL_MODULES := pkg/common $(SERVICE_DIRS)

# ─── Help ────────────────────────────────────────────────────────────────────

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── Infrastructure ─────────────────────────────────────────────────────────

up: ## Start infrastructure (PostgreSQL, Redis, RabbitMQ, MinIO)
	docker compose up -d

down: ## Stop infrastructure
	docker compose down

reset: ## Reset infrastructure (remove volumes)
	docker compose down -v
	docker compose up -d

# ─── Go ──────────────────────────────────────────────────────────────────────

build: ## Build all Go services
	@for dir in $(ALL_MODULES); do \
		echo "Building $$dir..."; \
		cd $$dir && go build ./... && cd $(CURDIR); \
	done

test: ## Run all Go tests
	@for dir in $(ALL_MODULES); do \
		echo "Testing $$dir..."; \
		cd $$dir && go test ./... && cd $(CURDIR); \
	done

test-coverage: ## Run Go tests with coverage
	@for dir in $(ALL_MODULES); do \
		echo "Testing $$dir with coverage..."; \
		cd $$dir && go test -coverprofile=coverage.out ./... && cd $(CURDIR); \
	done

lint: ## Run golangci-lint on all modules
	@for dir in $(ALL_MODULES); do \
		echo "Linting $$dir..."; \
		cd $$dir && golangci-lint run ./... && cd $(CURDIR); \
	done

fmt: ## Format Go code
	@for dir in $(ALL_MODULES); do \
		cd $$dir && gofmt -w . && cd $(CURDIR); \
	done

vet: ## Run go vet on all modules
	@for dir in $(ALL_MODULES); do \
		echo "Vetting $$dir..."; \
		cd $$dir && go vet ./... && cd $(CURDIR); \
	done

tidy: ## Tidy all Go modules
	go work sync
	@for dir in $(ALL_MODULES); do \
		echo "Tidying $$dir..."; \
		cd $$dir && go mod tidy && cd $(CURDIR); \
	done

# ─── Frontend ────────────────────────────────────────────────────────────────

fe-install: ## Install frontend dependencies
	cd frontend && npm ci

fe-dev: ## Start frontend dev server
	cd frontend && npm run dev

fe-build: ## Build frontend for production
	cd frontend && npm run build

fe-lint: ## Lint frontend code
	cd frontend && npx eslint src/

# ─── Run Services (local dev) ────────────────────────────────────────────────

run-gateway: ## Run API Gateway
	cd services/gateway && go run ./cmd/server

run-auth: ## Run Auth Service
	cd services/auth && go run ./cmd/server

run-incident: ## Run Incident Service
	cd services/incident && go run ./cmd/server

run-registry: ## Run Registry Service
	cd services/registry && go run ./cmd/server

run-ai-gateway: ## Run AI Gateway
	cd services/ai-gateway && go run ./cmd/server

# ─── Aggregate ───────────────────────────────────────────────────────────────

check: vet lint test ## Pre-commit gate: vet + lint + test
