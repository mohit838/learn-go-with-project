# Expense Task Tracker Microservices POC

Expense Task Tracker is a Go microservices proof of concept with four HTTP services:

- `auth`
- `task-tracker`
- `expense-tracker`
- `notification-service`

The local Docker stack runs the services behind Kong Gateway by default. APISIX is also available as a separate optional gateway for learning. Each service uses its own local `.env` file from its service directory.

## Gateway Choice

This project uses Kong in DB-less mode. Kong is a good fit here because it gives you clean path-based routing now and can later add authentication, rate limiting, request transforms, logging, and observability without putting that logic into every Go service.

For this project size, lighter options like Caddy, Traefik, or Nginx would also work. Kong is the better choice if you expect this repo to grow into a more complete microservices API platform.

APISIX is included side-by-side in standalone YAML mode so we can compare gateway behavior without changing the Kong setup.

## Local Setup

For daily development, run only Kong in Docker and run the Go services locally in terminal tabs:

```sh
make dev-up
```

Then start the services locally:

```sh
cd services/auth && go run ./cmd/api
cd services/task-tracker && go run ./cmd/api
cd services/expense-tracker && go run ./cmd/api
cd services/notification-service && go run ./cmd/api
```

Kong will route to your locally running services through `host.docker.internal`.

To try APISIX instead of Kong, run:

```sh
make apisix-dev
```

APISIX also routes to locally running services through `host.docker.internal`.

To try APISIX with the Dashboard UI, run:

```sh
make apisix-gui-up
```

Start everything in Docker when you want a full container check:

```sh
make prod-up
```

Open the services through Kong:

- Kong proxy: http://localhost:8000
- Auth: http://localhost:8000/auth
- Task tracker: http://localhost:8000/tasks
- Expense tracker: http://localhost:8000/expenses
- Notifications: http://localhost:8000/notifications
- Kong admin API: http://localhost:8001

Open the services through APISIX:

- APISIX proxy: http://localhost:9080
- Auth: http://localhost:9080/auth
- Task tracker: http://localhost:9080/tasks
- Task GraphQL: http://localhost:9080/tasks/graphql
- Expense tracker: http://localhost:9080/expenses
- Notifications: http://localhost:9080/notifications

Open APISIX Dashboard mode:

- APISIX GUI proxy: http://localhost:9088
- APISIX Dashboard: http://localhost:9181
- APISIX Admin API: http://localhost:9180
- Notifications: http://localhost:9088/notifications

Stop the stack:

```sh
make prod-down
```

Remove containers and volumes:

```sh
make clean
```

## Direct Service Ports

The services are also exposed directly for debugging:

- Auth: http://localhost:8484
- Task tracker: http://localhost:8485
- Expense tracker: http://localhost:8486
- Notification service: http://localhost:8487
- Notification gRPC: localhost:8587

Inside Docker, the services use the `APP_PORT` values from their own `.env` files.

## Useful Commands

```sh
make build
make logs
make ps
make test
make fmt
make tidy
make gateway
make apisix
make apisix-gui
make dev-check
make prod-check
```

Migration helpers use the `service` name from the `services/` folder:

```sh
make migrate-make service=auth name=create_users_table
make migrate-status service=auth
make migrate-up service=auth
make migrate-rollback service=auth
```

## Learning Docs

- [Development and production usage](docs/dev-prod-usage.md)
- [Step by step learning plan](docs/step-by-step.md)
- [Upcoming tasks](docs/upcoming-tasks.md)
- [Migrator library guideline](docs/migrator.md)
- [Kong gateway notes](docs/kong.md)
- [APISIX gateway notes](docs/apisix.md)
- [OpenAPI documentation](docs/openapi.md)
- [Go concurrency with notification service](docs/go-concurrency-notifications.md)

## Project Layout

```text
.
├── docker-compose.yml
├── docker-compose.apisix.yml
├── docker-compose.apisix-gui.yml
├── apisix/
│   └── apisix.dev.yaml
├── kong/
│   └── kong.yml
├── services/
│   ├── auth/
│   ├── expense-tracker/
│   └── task-tracker/
├── .env.example
├── .dockerignore
├── Makefile
└── README.md
```

Each service has its own Go module, Dockerfile, and local `.env` file.
