# First start: run make install, copy .env.example to .env and set its empty values, then run make up.
# Then run make migrate-up and make dev.

# The shell reads .env, so a value with quotes or # works as in a shell script.
# Add LOAD_ENV before a command that needs the values of .env.
LOAD_ENV := set -a; [ ! -f .env ] || . ./.env; set +a;

MOCKERY_VERSION := v3.8.1
MOCKERY := go run github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
GOLANGCI_LINT_VERSION := v2.15.0
GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

COMPOSE_DEV := docker compose -f compose.yaml -f compose.dev.yaml

.PHONY: install gen stack-up stack-down up down logs migrate-up migrate-down dev dev-api dev-worker dev-web test test-e2e test-web-e2e lint fmt

install: ## Install the frontend dependencies
	pnpm install

gen: ## Generate the mocks of the ports and the frontend types from the OpenAPI spec
	cd backend && $(MOCKERY)
	pnpm --filter frontend run gen

stack-up: ## Build and start the full stack, as it runs in deployment. Stop make dev first, because both use the ports 5651 and 5653
	docker compose up -d --build --wait

# Compose needs a value for POSTGRES_PASSWORD with all commands, also when .env is missing.
# --remove-orphans also removes the services that only compose.dev.yaml defines, for example Asynqmon.
stack-down: ## Stop the full stack. The data stays in the volumes
	POSTGRES_PASSWORD=unused docker compose down --remove-orphans

up: ## Start Postgres, Redis and Asynqmon for development
	$(COMPOSE_DEV) up -d --wait postgres redis asynqmon

# The next 2 commands do not use the value of POSTGRES_PASSWORD, but Compose needs a value.
down: ## Stop the development services. The data stays in the volumes
	POSTGRES_PASSWORD=unused $(COMPOSE_DEV) down

logs: ## Show the logs of the development services
	POSTGRES_PASSWORD=unused $(COMPOSE_DEV) logs -f

migrate-up: ## Apply the new database migrations
	$(LOAD_ENV) cd backend && go run ./cmd/migrate up

migrate-down: ## Revert the last database migration
	$(LOAD_ENV) cd backend && go run ./cmd/migrate down

dev: ## Start the API, the worker and the frontend on the host. Press Ctrl-C to stop the 3 processes
	$(MAKE) -j3 dev-api dev-worker dev-web

dev-api: ## Start the API on the host
	$(LOAD_ENV) cd backend && go run ./cmd/api

dev-worker: ## Start the worker on the host
	$(LOAD_ENV) cd backend && go run ./cmd/worker

dev-web: ## Start the frontend dev server on the host
	pnpm --filter frontend run dev

test: ## Run the unit tests of the backend and the frontend
	cd backend && go test ./...
	pnpm --filter frontend run test

test-e2e: ## Run the E2E tests, one after the other
	cd backend && go test -tags e2e -p 1 ./api-test/e2e/...

test-web-e2e: ## Run the browser tests against the full stack in a Compose project of its own, on the port 5655
	./scripts/test-web-e2e.sh

lint: ## Check the code of the backend and the frontend
	cd backend && $(GOLANGCI_LINT) run --build-tags e2e ./...
	pnpm --filter frontend run lint

fmt: ## Format the code of the backend and the frontend
	cd backend && $(GOLANGCI_LINT) fmt ./...
	pnpm --filter frontend run fmt
