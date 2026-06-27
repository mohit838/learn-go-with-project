COMPOSE ?= docker compose
DEV_COMPOSE ?= docker compose -f docker-compose.dev.yml
APISIX_COMPOSE ?= docker compose -f docker-compose.apisix.yml

.PHONY: help build up down restart logs ps gateway gateway-dev gateway-dev-down gateway-dev-logs gateway-dev-ps apisix apisix-dev apisix-dev-down apisix-dev-logs apisix-dev-ps swagger-auth swagger-task-tracker swagger-expense-tracker swagger-generate swagger-generate-auth swagger-generate-task-tracker swagger-generate-expense-tracker migrate-up migrate-down migrate-rollback migrate-status migrate-make test fmt tidy clean

help:
	@printf "Available targets:\n"
	@printf "  make gateway-dev       Start Kong only for local service development\n"
	@printf "  make gateway-dev-down  Stop Kong dev gateway\n"
	@printf "  make gateway-dev-logs  Follow Kong dev logs\n"
	@printf "  make apisix-dev        Start APISIX only for local service development\n"
	@printf "  make apisix-dev-down   Stop APISIX dev gateway\n"
	@printf "  make apisix-dev-logs   Follow APISIX dev logs\n"
	@printf "  make swagger-auth     Open Auth API documentation at http://localhost:8081\n"
	@printf "  make swagger-task-tracker Open Task Tracker API documentation at http://localhost:8082\n"
	@printf "  make swagger-expense-tracker Open Expense Tracker API documentation at http://localhost:8083\n"
	@printf "  make swagger-generate Generate Swagger files for all services\n"
	@printf "  make swagger-generate-auth Generate Swagger files for Auth\n"
	@printf "  make swagger-generate-task-tracker Generate Swagger files for Task Tracker\n"
	@printf "  make swagger-generate-expense-tracker Generate Swagger files for Expense Tracker\n"
	@printf "  make migrate-up service=auth Run pending migrations for one service\n"
	@printf "  make migrate-rollback service=auth Roll back latest migration batch for one service\n"
	@printf "  make migrate-status service=auth Show migration status for one service\n"
	@printf "  make migrate-make service=auth name=create_users_table Create paired migration files\n"
	@printf "  make build     Build all service images\n"
	@printf "  make up        Start services and Kong\n"
	@printf "  make down      Stop and remove containers\n"
	@printf "  make restart   Restart the Docker Compose stack\n"
	@printf "  make logs      Follow stack logs\n"
	@printf "  make ps        Show stack containers\n"
	@printf "  make gateway   Show gateway URLs\n"
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

apisix:
	@printf "APISIX proxy:     http://localhost:9080\n"
	@printf "Auth service:     http://localhost:9080/auth\n"
	@printf "Task tracker:     http://localhost:9080/tasks\n"
	@printf "Task GraphQL:     http://localhost:9080/tasks/graphql\n"
	@printf "Expense tracker:  http://localhost:9080/expenses\n"

apisix-dev:
	$(APISIX_COMPOSE) up -d
	@$(MAKE) apisix

apisix-dev-down:
	$(APISIX_COMPOSE) down

apisix-dev-logs:
	$(APISIX_COMPOSE) logs -f

apisix-dev-ps:
	$(APISIX_COMPOSE) ps

swagger-auth:
	docker run --rm -p 8081:8080 -e SWAGGER_JSON=/spec/openapi.yaml -v "$(CURDIR)/services/auth/docs/openapi.yaml:/spec/openapi.yaml:ro" swaggerapi/swagger-ui

swagger-task-tracker:
	docker run --rm -p 8082:8080 -e SWAGGER_JSON=/spec/openapi.yaml -v "$(CURDIR)/services/task-tracker/docs/openapi.yaml:/spec/openapi.yaml:ro" swaggerapi/swagger-ui

swagger-expense-tracker:
	docker run --rm -p 8083:8080 -e SWAGGER_JSON=/spec/openapi.yaml -v "$(CURDIR)/services/expense-tracker/docs/openapi.yaml:/spec/openapi.yaml:ro" swaggerapi/swagger-ui

swagger-generate: swagger-generate-auth swagger-generate-task-tracker swagger-generate-expense-tracker

swagger-generate-auth:
	cd services/auth && swag init -g cmd/api/main.go -o docs

swagger-generate-task-tracker:
	cd services/task-tracker && swag init -g cmd/api/main.go -o docs

swagger-generate-expense-tracker:
	cd services/expense-tracker && swag init -g cmd/api/main.go -o docs

migrate-up:
	@test -n "$(service)" || (printf "service is required, example: make migrate-up service=auth\n" && exit 1)
	cd services/$(service) && go run ./cmd/migrate up

migrate-down:
	@test -n "$(service)" || (printf "service is required, example: make migrate-down service=auth\n" && exit 1)
	cd services/$(service) && go run ./cmd/migrate down

migrate-rollback:
	@test -n "$(service)" || (printf "service is required, example: make migrate-rollback service=auth\n" && exit 1)
	cd services/$(service) && go run ./cmd/migrate rollback

migrate-status:
	@test -n "$(service)" || (printf "service is required, example: make migrate-status service=auth\n" && exit 1)
	cd services/$(service) && go run ./cmd/migrate status

migrate-make:
	@test -n "$(service)" || (printf "service is required, example: make migrate-make service=auth name=create_users_table\n" && exit 1)
	@test -n "$(name)" || (printf "name is required, example: make migrate-make service=auth name=create_users_table\n" && exit 1)
	cd services/$(service) && go run ./cmd/migrate make $(name)

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
