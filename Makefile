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

.PHONY: install gen up down logs migrate-up migrate-down dev dev-api dev-web test test-e2e lint fmt

install: ## Install the frontend dependencies
	pnpm install

gen: ## Generate the mocks of the ports and the frontend types from the OpenAPI spec
	cd backend && $(MOCKERY)
	pnpm --filter frontend run gen

up: ## Start Postgres and Redis for development
	$(COMPOSE_DEV) up -d --wait postgres redis

# Compose needs a value for POSTGRES_PASSWORD with all commands, also when .env is missing.
# The two commands below do not use the value.
down: ## Stop the development services. The data stays in the volumes
	POSTGRES_PASSWORD=unused $(COMPOSE_DEV) down

logs: ## Show the logs of the development services
	POSTGRES_PASSWORD=unused $(COMPOSE_DEV) logs -f

migrate-up: ## Apply the new database migrations
	$(LOAD_ENV) cd backend && go run ./cmd/migrate up

migrate-down: ## Revert the last database migration
	$(LOAD_ENV) cd backend && go run ./cmd/migrate down

dev: ## Start the API and the frontend on the host. Press Ctrl-C to stop both
	$(MAKE) -j2 dev-api dev-web

dev-api: ## Start the API on the host
	$(LOAD_ENV) cd backend && go run ./cmd/api

dev-web: ## Start the frontend dev server on the host
	pnpm --filter frontend run dev

test: ## Run the unit tests of the backend and the frontend
	cd backend && go test ./...
	pnpm --filter frontend run test

test-e2e: ## Run the E2E tests, one after the other
	cd backend && go test -tags e2e -p 1 ./api-test/e2e/...

lint: ## Check the code of the backend and the frontend
	cd backend && $(GOLANGCI_LINT) run --build-tags e2e ./...
	pnpm --filter frontend run lint

fmt: ## Format the code of the backend and the frontend
	cd backend && $(GOLANGCI_LINT) fmt ./...
	pnpm --filter frontend run fmt
