# The shell reads .env, so a value with quotes or # works as in a shell script.
# Add LOAD_ENV before a command that needs the values of .env.
LOAD_ENV := set -a; [ ! -f .env ] || . ./.env; set +a;

GOLANGCI_LINT_VERSION := v2.15.0
GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: dev test test-e2e lint fmt

dev: ## Start the API on the host
	$(LOAD_ENV) cd backend && go run ./cmd/api

test: ## Run the unit tests
	cd backend && go test ./...

test-e2e: ## Run the E2E tests, one after the other
	cd backend && go test -tags e2e -p 1 ./api-test/e2e/...

lint: ## Check the code
	cd backend && $(GOLANGCI_LINT) run --build-tags e2e ./...

fmt: ## Format the code
	cd backend && $(GOLANGCI_LINT) fmt ./...
