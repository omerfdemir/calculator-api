SHELL := /bin/bash
.DEFAULT_GOAL := help

BACKEND  := backend
FRONTEND := frontend
E2E      := e2e
COVERAGE := coverage

# Minimum statement coverage of the calculator and API modules.
BACKEND_COVERAGE_MIN := 90

.PHONY: help install run-backend run-frontend test test-backend test-frontend \
	lint lint-backend lint-frontend lint-e2e coverage coverage-backend coverage-frontend \
	up down logs e2e check

help: ## Show the targets.
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-18s %s\n", $$1, $$2}'

$(FRONTEND)/node_modules: $(FRONTEND)/package-lock.json
	cd $(FRONTEND) && npm ci
	@touch $@

$(E2E)/node_modules: $(E2E)/package-lock.json
	cd $(E2E) && npm ci
	@touch $@

install: $(FRONTEND)/node_modules $(E2E)/node_modules ## Install the Node packages.

run-backend: ## Run the API on port 8080.
	cd $(BACKEND) && go run ./cmd/server

run-frontend: $(FRONTEND)/node_modules ## Run the Vite dev server (proxies /api to port 8080).
	cd $(FRONTEND) && npm run dev

test: test-backend test-frontend ## Run the backend and frontend tests.

test-backend:
	cd $(BACKEND) && go test -race -cover ./...

test-frontend: $(FRONTEND)/node_modules
	cd $(FRONTEND) && npm test

lint: lint-backend lint-frontend lint-e2e ## Run all linters and type checks.

lint-backend:
	cd $(BACKEND) && go vet ./... && golangci-lint run

lint-frontend: $(FRONTEND)/node_modules
	cd $(FRONTEND) && npm run lint && npm run typecheck

lint-e2e: $(E2E)/node_modules
	cd $(E2E) && npm run lint

coverage: coverage-backend coverage-frontend ## Make the HTML and text coverage reports in coverage/.

coverage-backend:
	@mkdir -p $(COVERAGE)/backend
	cd $(BACKEND) && go test -race -covermode=atomic -coverpkg=./internal/... \
		-coverprofile=../$(COVERAGE)/backend/coverage.out ./...
	cd $(BACKEND) && go tool cover -func=../$(COVERAGE)/backend/coverage.out \
		> ../$(COVERAGE)/backend/coverage.txt
	cat $(COVERAGE)/backend/coverage.txt
	cd $(BACKEND) && go tool cover -html=../$(COVERAGE)/backend/coverage.out \
		-o ../$(COVERAGE)/backend/index.html
	@total=$$(awk '/^total:/ {sub("%", "", $$3); print $$3}' $(COVERAGE)/backend/coverage.txt); \
		echo "Backend coverage (calculator + api): $$total% (minimum $(BACKEND_COVERAGE_MIN)%)"; \
		awk -v t="$$total" -v m="$(BACKEND_COVERAGE_MIN)" 'BEGIN { exit (t + 0 >= m + 0) ? 0 : 1 }'

coverage-frontend: $(FRONTEND)/node_modules
	cd $(FRONTEND) && npm test -- --coverage

up: ## Build and start the Compose stack, and wait until it is healthy.
	docker compose up --build -d --wait

down: ## Stop the Compose stack.
	docker compose down

logs: ## Show the Compose logs.
	docker compose logs

e2e: $(E2E)/node_modules ## Run the Playwright tests against the running stack (make up first).
	cd $(E2E) && npx playwright install chromium && npm test

check: lint test ## Run lint, tests, the stack and the end-to-end tests.
	$(MAKE) up
	$(MAKE) e2e || { status=$$?; $(MAKE) logs; $(MAKE) down; exit $$status; }
	$(MAKE) down
