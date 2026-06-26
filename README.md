# Expense Task Tracker Microservices POC

Expense Task Tracker is a Go microservices proof of concept with three HTTP services:

- `auth`
- `task-tracker`
- `expense-tracker`

The local Docker stack runs the services behind Kong Gateway. Each service uses its own local `.env` file from its service directory.

## Gateway Choice

This project uses Kong in DB-less mode. Kong is a good fit here because it gives you clean path-based routing now and can later add authentication, rate limiting, request transforms, logging, and observability without putting that logic into every Go service.

For this project size, lighter options like Caddy, Traefik, or Nginx would also work. Kong is the better choice if you expect this repo to grow into a more complete microservices API platform.

## Local Setup

For daily development, run only Kong in Docker and run the Go services locally in terminal tabs:

```sh
make gateway-dev
```

Then start the services locally:

```sh
cd services/auth && go run ./cmd/api
cd services/task-tracker && go run ./cmd/api
cd services/expense-tracker && go run ./cmd/api
```

Kong will route to your locally running services through `host.docker.internal`.

Start everything in Docker when you want a full container check:

```sh
make up
```

Open the services through Kong:

- Auth: http://localhost:8000/auth
- Task tracker: http://localhost:8000/tasks
- Expense tracker: http://localhost:8000/expenses
- Kong admin API: http://localhost:8001

Stop the stack:

```sh
make down
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
```

Migration helpers use the `service` name from the `services/` folder:

```sh
make migrate-make service=auth name=create_users_table
make migrate-status service=auth
make migrate-up service=auth
make migrate-rollback service=auth
```

## Learning Docs

- [Step by step learning plan](docs/step-by-step.md)
- [Upcoming tasks](docs/upcoming-tasks.md)
- [Kong gateway notes](docs/kong.md)
- [OpenAPI documentation](docs/openapi.md)

## Project Layout

```text
.
├── docker-compose.yml
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
