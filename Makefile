DIR    := $(shell pwd)
BIN    := $(DIR)/dist/server
MAIN   := $(DIR)/cmd/server/main.go
WIZARD := $(DIR)/scripts/env-wizard.py

.PHONY: run build-bin run-bin deps-tidy lint test-api \
        docker-image docker-stack docker-start docker-stop docker-tail docker-services \
        config config-update config-reset \
        dev build start tidy vet test-integration \
        docker-build docker-dev docker-up docker-down docker-logs docker-infra \
        env env-update env-reset help

# ── Test configuration (override with environment variables) ──────────────────
TEST_DB_HOST     ?= localhost
TEST_DB_PORT     ?= 5432
TEST_DB_USER     ?= postgres
TEST_DB_PASSWORD ?= postgres
TEST_DB_NAME     ?= dx_auth_test
TEST_REDIS_HOST  ?= localhost
TEST_REDIS_PORT  ?= 6379

help:
	@echo ""
	@echo "  Usage: make [target]"
	@echo ""
	@echo "  Environment"
	@echo "    config        Create or update .env with the guided setup wizard"
	@echo "    config-update Update an existing .env with the same wizard"
	@echo "    config-reset  Recreate .env from scratch with the wizard"
	@echo ""
	@echo "  Development"
	@echo "    run           Run the server with go run (development mode)"
	@echo "    build-bin     Compile the server binary to ./dist/server"
	@echo "    run-bin       Run the existing compiled binary without rebuilding"
	@echo "    deps-tidy     Run go mod tidy"
	@echo "    lint          Run go vet"
	@echo ""
	@echo "  Testing"
	@echo "    test-api      Run API integration tests (requires Postgres + Redis)"
	@echo ""
	@echo "  Docker"
	@echo "    docker-image    Build the Docker image and save to ./dist/"
	@echo "    docker-stack    Build image and start all containers"
	@echo "    docker-start    Start all containers using existing image (no build)"
	@echo "    docker-stop     Stop and remove containers"
	@echo "    docker-tail     Follow container logs"
	@echo "    docker-services Start only postgres and redis containers"
	@echo ""

# ── Environment ──────────────────────────────────────────────────────────────

config:
	@if [ -f $(DIR)/.env ]; then \
		echo ""; \
		echo "  .env already exists — opening update mode."; \
		echo ""; \
		python3 $(WIZARD) update; \
	else \
		python3 $(WIZARD) create; \
	fi

config-update:
	@if [ ! -f $(DIR)/.env ]; then \
		echo ""; \
		echo "  No .env found."; \
		echo "    Run 'make config' to create one."; \
		echo ""; \
	else \
		python3 $(WIZARD) update; \
	fi

config-reset:
	@python3 $(WIZARD) create

# ── Development ───────────────────────────────────────────────────────────────

run:
	go run $(MAIN)

build-bin:
	mkdir -p $(DIR)/dist
	go build -o $(BIN) $(MAIN)

run-bin:
	$(BIN)

deps-tidy:
	go mod tidy

lint:
	go vet ./...

test-api:
	TEST_DB_HOST=$(TEST_DB_HOST) \
	TEST_DB_PORT=$(TEST_DB_PORT) \
	TEST_DB_USER=$(TEST_DB_USER) \
	TEST_DB_PASSWORD=$(TEST_DB_PASSWORD) \
	TEST_DB_NAME=$(TEST_DB_NAME) \
	TEST_REDIS_HOST=$(TEST_REDIS_HOST) \
	TEST_REDIS_PORT=$(TEST_REDIS_PORT) \
	go test -v -race -count=1 -timeout=120s ./tests/integration/...

# ── Docker ────────────────────────────────────────────────────────────────────

docker-image:
	mkdir -p $(DIR)/dist
	docker build --no-cache -t laodvx-server-auth .
	docker save laodvx-server-auth -o $(DIR)/dist/laodvx-server-auth.tar

docker-stack: docker-image
	docker load -i $(DIR)/dist/laodvx-server-auth.tar
	docker compose up -d

docker-start:
	docker load -i $(DIR)/dist/laodvx-server-auth.tar
	docker compose up -d

docker-stop:
	docker compose down

docker-tail:
	docker compose logs -f

docker-services:
	docker compose up -d postgres redis

# ── Backward-compatible aliases ───────────────────────────────────────────────

env: config

env-update: config-update

env-reset: config-reset

dev: run

build: build-bin

start: run-bin

tidy: deps-tidy

vet: lint

test-integration: test-api

docker-build: docker-image

docker-dev: docker-stack

docker-up: docker-start

docker-down: docker-stop

docker-logs: docker-tail

docker-infra: docker-services
