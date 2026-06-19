COMPOSE ?= docker compose
DEV_COMPOSE ?= docker compose -f docker-compose.dev.yml
MIGRATION_SERVICE ?= auth

.PHONY: help build up down restart logs ps gateway gateway-dev gateway-dev-down gateway-dev-logs gateway-dev-ps migrate-up migrate-down migrate-version test fmt tidy clean

help:
	@printf "Available targets:\n"
	@printf "  make gateway-dev       Start Kong only for local service development\n"
	@printf "  make gateway-dev-down  Stop Kong dev gateway\n"
	@printf "  make gateway-dev-logs  Follow Kong dev logs\n"
	@printf "  make build     Build all service images\n"
	@printf "  make up        Start services and Kong\n"
	@printf "  make down      Stop and remove containers\n"
	@printf "  make restart   Restart the Docker Compose stack\n"
	@printf "  make logs      Follow stack logs\n"
	@printf "  make ps        Show stack containers\n"
	@printf "  make gateway   Show gateway URLs\n"
	@printf "  make migrate-up MIGRATION_SERVICE=auth  Apply a service's migrations\n"
	@printf "  make migrate-down MIGRATION_SERVICE=auth  Revert one migration\n"
	@printf "  make migrate-version MIGRATION_SERVICE=auth  Show migration version\n"
	@printf "  make test      Run Go tests in every service\n"
	@printf "  make fmt       Format Go code in every service\n"
	@printf "  make tidy      Run go mod tidy in every service\n"
	@printf "  make clean     Remove containers and volumes\n"

build:
	$(COMPOSE) build

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

restart: down up

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

gateway:
	@printf "Kong proxy:       http://localhost:8000\n"
	@printf "Auth service:     http://localhost:8000/auth\n"
	@printf "Task tracker:     http://localhost:8000/tasks\n"
	@printf "Expense tracker:  http://localhost:8000/expenses\n"
	@printf "Kong admin API:   http://localhost:8001\n"

gateway-dev:
	$(DEV_COMPOSE) up -d
	@$(MAKE) gateway

gateway-dev-down:
	$(DEV_COMPOSE) down

gateway-dev-logs:
	$(DEV_COMPOSE) logs -f

gateway-dev-ps:
	$(DEV_COMPOSE) ps

migrate-up:
	cd services/$(MIGRATION_SERVICE) && go run ./cmd/migrate up

migrate-down:
	cd services/$(MIGRATION_SERVICE) && go run ./cmd/migrate down

migrate-version:
	cd services/$(MIGRATION_SERVICE) && go run ./cmd/migrate version

test:
	@for service in services/*; do \
		if [ -f "$$service/go.mod" ]; then \
			printf "\n==> Testing $$service\n"; \
			(cd "$$service" && go test ./...); \
		fi; \
	done

fmt:
	@for service in services/*; do \
		if [ -f "$$service/go.mod" ]; then \
			printf "\n==> Formatting $$service\n"; \
			(cd "$$service" && gofmt -w .); \
		fi; \
	done

tidy:
	@for service in services/*; do \
		if [ -f "$$service/go.mod" ]; then \
			printf "\n==> Tidying $$service\n"; \
			(cd "$$service" && go mod tidy); \
		fi; \
	done

clean:
	$(COMPOSE) down -v --remove-orphans
