# Author: Al Amin Ahamed
# Sitemon Platform — dev workflow entrypoints.
COMPOSE := docker compose -f infra/docker-compose.yml

.PHONY: help up down logs ps build test test-race lint tidy fmt vet run-cli \
        gateway obs-up clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n",$$1,$$2}'

# ---- Docker stack -----------------------------------------------------------
up: ## Start the local stack (mongo, redis, nats, gateway)
	$(COMPOSE) up -d

down: ## Stop the stack and remove volumes
	$(COMPOSE) down -v

logs: ## Tail logs (s=<service> to filter, e.g. make logs s=gateway)
	$(COMPOSE) logs -f $(s)

ps: ## Show stack status
	$(COMPOSE) ps

# ---- Go ---------------------------------------------------------------------
build: ## Build all service binaries into ./bin
	@mkdir -p bin
	@for svc in gateway checker scheduler notifier ai cli; do \
	  echo "building $$svc"; \
	  go build -trimpath -o bin/$$svc ./apps/$$svc/cmd || exit 1; \
	done

test: ## Run unit tests
	go test ./...

test-race: ## Run tests with the race detector
	CGO_ENABLED=1 go test -race ./...

lint: ## Vet + gofmt check
	go vet ./...
	@test -z "$$(gofmt -l apps packages)" || { echo "gofmt needed:"; gofmt -l apps packages; exit 1; }

fmt: ## Format code
	gofmt -w apps packages

vet: ## Static checks
	go vet ./...

tidy: ## Sync go.mod/go.sum
	go mod tidy

# ---- Local run --------------------------------------------------------------
run-cli: ## Run the CLI (ARGS="check -u https://example.com")
	go run ./apps/cli $(ARGS)

gateway: ## Run the gateway locally
	go run ./apps/gateway/cmd

clean: ## Remove build artifacts
	rm -rf bin
