# Step By Step Learning Plan

This project is a Go microservices learning repo. The goal is to build the same
basic shape in all three services first, then decide later what should stay,
what should be removed, and what should become service-specific.

Services:

| Service | Folder | Local port | Kong route |
| --- | --- | --- | --- |
| Auth | `services/auth` | `8484` | `/auth` |
| Task tracker | `services/task-tracker` | `8485` | `/tasks` |
| Expense tracker | `services/expense-tracker` | `8486` | `/expenses` |

## 1. Install Go

Install Go from the official download page:

```sh
https://go.dev/dl/
```

Check the installation:

```sh
go version
go env GOPATH
go env GOMODCACHE
```

For this repo, each service is its own Go module:

```sh
cd services/auth && go mod tidy
cd services/task-tracker && go mod tidy
cd services/expense-tracker && go mod tidy
```

## 2. Create And Load Environment Files

Each service owns its own `.env` file.

```sh
cp services/auth/.env.example services/auth/.env
cp services/task-tracker/.env.example services/task-tracker/.env
cp services/expense-tracker/.env.example services/expense-tracker/.env
```

Current important environment groups:

- App: `APP_ENV`, `APP_NAME`, `APP_VERSION`, `APP_PORT`, `APP_DEBUG`,
  `APP_LOG_LEVEL`
- Postgres: `DATABASE_URL`
- Redis: `REDIS_URL`
- MongoDB: `MONGO_URL`, `MONGO_DB`
- MinIO: `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`,
  `MINIO_BUCKET`, `MINIO_REGION`, `MINIO_USE_SSL`

Learning note: the config loader already supports MongoDB settings. Keep the
`.env.example` files symmetric so every service shows the same required groups,
even if one service does not use every dependency yet.

## 3. Run A Service Locally

Run one service directly while developing:

```sh
cd services/auth
go run ./cmd/api
```

Expected direct service URLs:

```text
http://localhost:8484
http://localhost:8485
http://localhost:8486
```

Run all tests:

```sh
make test
```

Format all services:

```sh
make fmt
```

## 4. Add Chi Router

Each service uses `chi` for HTTP routing and middleware.

Install it inside a service module when needed:

```sh
cd services/auth
go get github.com/go-chi/chi/v5
```

Basic router responsibilities:

- Create `chi.NewRouter()`.
- Register common middleware.
- Register health and readiness routes.
- Register feature routes from handler packages.
- Return `http.Handler` to `cmd/api/main.go`.

Recommended router files:

```text
internal/router/router.go
internal/router/app_api.go
internal/router/health.go
```

## 5. Add Database Connections

Start with small connection packages under `internal/database`.

Recommended files:

```text
internal/database/postgres.go
internal/database/redis.go
internal/database/mongo.go
internal/database/minio.go
```

Connection rules:

- Create clients in `cmd/api/main.go`.
- Verify the connection with `Ping` or an equivalent health check.
- Close clients on shutdown.
- Pass dependencies into the router or app container.
- Do not open database connections inside handlers.

## 6. Add Postgres

Postgres is the primary relational database. Use it for core owned data such as
users, tasks, expenses, categories, and records that need relational queries.

Current shape:

```go
db, err := database.ConnectDB(cfg.DBURL)
```

Next improvements:

- Add context timeout to `Ping`.
- Configure connection pool settings.
- Add migrations.
- Add repository tests around SQL behavior.

## 7. Add Redis

Redis should have a real purpose, not duplicate Postgres data.

Good first use cases:

- Cache `GET /users/{id}` in Auth.
- Store short-lived sessions.
- Store rate-limit counters.

Suggested key style:

```text
auth:user:{id}
task:user:{user_id}:summary
expense:user:{user_id}:monthly-summary:{yyyy-mm}
```

Rules:

- Use short TTLs for cache.
- Delete cache keys when the source record changes.
- Test cache hit, cache miss, and cache invalidation.

## 8. Add MongoDB

MongoDB is useful for document-oriented data. Do not copy the same relational
tables from Postgres into MongoDB.

Good first use cases:

- Audit logs.
- Activity feeds.
- Flexible user preferences.
- Event payloads where shape can vary.

Suggested collections:

```text
auth_logs
task_activity_logs
expense_activity_logs
```

Rules:

- Keep MongoDB ownership clear.
- Add indexes when queries need them.
- Add context timeouts to reads and writes.

## 9. Add MinIO

MinIO is object storage. Use it for files, images, exports, receipts, and
attachments.

Good first use cases:

- Auth user avatar.
- Task attachment.
- Expense receipt image.

Rules:

- Ensure the bucket exists at startup or in a setup command.
- Store object metadata in Postgres.
- Store object bytes in MinIO.
- Never store large files directly in Postgres.

## 10. Add Health And Readiness Routes

Use simple health routes for learning first, then split liveness and readiness.

Recommended routes:

```text
GET /
GET /health
GET /ready
GET /health/postgres
GET /health/redis
GET /health/mongo
GET /health/minio
```

Response style should become JSON:

```json
{
  "success": true,
  "message": "service is ready",
  "data": {
    "postgres": "ok",
    "redis": "ok",
    "mongo": "ok",
    "minio": "ok"
  }
}
```

## 11. Add Kong

Kong is the public gateway in front of the services.

Start only Kong for local service development:

```sh
make gateway-dev
```

Then run Go services locally:

```sh
cd services/auth && go run ./cmd/api
cd services/task-tracker && go run ./cmd/api
cd services/expense-tracker && go run ./cmd/api
```

Gateway URLs:

```text
http://localhost:8000/auth
http://localhost:8000/tasks
http://localhost:8000/expenses
```

Full Docker stack:

```sh
make up
```

Kong config files:

```text
kong/kong.yml
kong/kong.dev.yml
```

## 12. Next: Build A Go Migration Tool

After the service skeleton is stable, add a migration system that feels familiar
if you have used Laravel migrations.

Goal commands:

```sh
cd services/auth
go run ./cmd/migrate up
go run ./cmd/migrate down
go run ./cmd/migrate rollback
go run ./cmd/migrate status
go run ./cmd/migrate make create_users_table
```

Later, add root Makefile shortcuts:

```sh
make migrate-up service=auth
make migrate-down service=auth
make migrate-rollback service=auth
make migrate-status service=auth
make migrate-make service=auth name=create_users_table
```

Migration folder shape:

```text
services/auth/migrations/
  20260101010101_create_users_table.up.sql
  20260101010101_create_users_table.down.sql
```

Migration tracking table:

```text
schema_migrations
```

Suggested fields:

```text
id
version
name
batch
checksum
applied_at
execution_ms
```

Behavior:

- `up` runs all pending `.up.sql` files in order.
- `down` rolls back the latest batch or one migration, depending on the command.
- `rollback` rolls back the latest batch like Laravel.
- `status` shows ran and pending migrations.
- `make` creates paired `.up.sql` and `.down.sql` files.
- Every migration runs in a transaction when possible.
- The tool refuses to run if a checksum changed after being applied.

Recommended implementation:

```text
cmd/migrate/main.go
internal/migration/runner.go
internal/migration/store.go
internal/migration/file.go
```

Keep this symmetric across the three services first. Later, if duplication feels
annoying, extract it into a shared package or a small internal CLI.

## 13. Next: DDD-Friendly Project Layers

After migrations, build each service around feature modules. Keep shared
infrastructure helpers at `internal/`, but put business features in their own
bounded-context folder.

Recommended Auth shape:

```text
internal/auth/
  domain/
  application/
  infrastructure/
  transport/
internal/config/
internal/constants/
internal/database/
internal/migration/
internal/response/
internal/router/
internal/utils/
```

Suggested purpose:

| Folder | Purpose |
| --- | --- |
| `auth/domain` | Entities, domain types, repository interfaces, domain errors |
| `auth/application` | Use cases, request/response DTOs, token/password workflow |
| `auth/infrastructure` | Postgres, Redis, MongoDB, external provider implementations |
| `auth/transport` | HTTP handlers and request/response mapping |
| `config` | Environment loading |
| `constants` | App constants, route names, cache prefixes, default values |
| `database` | Database/client connection helpers |
| `migration` | Migration runner |
| `response` | Standard success and error JSON response helpers |
| `router` | Route registration, middleware, dependency wiring |
| `utils` | Small reusable helpers only |

Request/response style:

```text
application.RegisterRequest
application.UserResponse
response.Body
utils.APITime
```

Transport rule:

- Decode request.
- Call application use case.
- Map result to JSON response.

Application rule:

- Own use-case flow.
- Validate request data.
- Coordinate domain ports, cache, object storage, and events.
- Return domain errors.

Domain rule:

- Define business data and interfaces.
- Do not import SQL, HTTP, Redis, MongoDB, or framework packages.

Infrastructure rule:

- Implement domain interfaces.
- Own SQL or external storage calls.
- Accept context and return domain types.

## 14. Later: Add Observability

Add observability after the APIs have real behavior. It is easier to learn when
there are real requests, database calls, cache hits, errors, and gateway traffic
to inspect.

Recommended local stack:

```text
Prometheus
Grafana
Loki
Promtail or Grafana Alloy
OpenTelemetry Collector
```

Optional later stack:

```text
Elastic APM
Elasticsearch
Kibana
```

Start with these signals:

- Metrics: request count, status code, duration, in-flight requests, dependency
  health.
- Logs: structured JSON logs from Go services and Kong.
- Traces: request path through Kong, handler, service, repository, and external
  dependencies.
- APM: transaction timing, error grouping, slow endpoint discovery, and service
  maps.

Suggested files:

```text
docker-compose.observability.yml
observability/prometheus/prometheus.yml
observability/grafana/provisioning/
observability/loki/loki.yml
observability/otel-collector/config.yml
```

Suggested Makefile commands for local development:

```sh
make observe-dev
make observe-dev-down
make observe-dev-logs
make observe-dev-ps
```

Possible production-friendly commands later:

```sh
make observe-prod-config-check
make observe-prod-up
make observe-prod-down
```

Production notes to remember:

- Add authentication for Grafana and any public dashboards.
- Use persistent volumes for Grafana, Prometheus, Loki, and Elasticsearch.
- Set retention periods so logs and metrics do not grow forever.
- Add resource limits.
- Keep secrets out of Git.
- Prefer OpenTelemetry instrumentation in Go code so the vendor can change
  later without rewriting all app logic.

Done when one request through Kong can be seen in all three places:

```text
Prometheus/Grafana metrics
Loki logs
OpenTelemetry traces or Elastic APM traces
```

## 15. Suggested Build Order From Here

1. Make `.env.example` symmetric for all three services.
2. Add standard JSON response helpers.
3. Add health/readiness JSON routes.
4. Add migration runner for Auth.
5. Add Auth `users` migration.
6. Add Auth domain, application, infrastructure, transport, and routes.
7. Add tests for Auth application use cases, transport handlers, and repositories.
8. Copy the same structure to Task and Expense for learning symmetry.
9. Add Redis cache to one read endpoint.
10. Add MongoDB audit events.
11. Add MinIO file upload for one real use case.
12. Update OpenAPI docs after each route change.
13. Add local observability Docker stack.
14. Add Go metrics, structured logs, and traces.
15. Add production-friendly observability notes and Make targets.

## 16. Keep Or Remove Later

For learning, symmetry is good. It helps you understand the pattern.

Later, simplify:

- Remove folders that stay empty.
- Keep helpers only when used by more than one feature.
- Avoid generic abstractions until duplication becomes painful.
- Let each service keep its own business language.
- Keep infrastructure patterns consistent, but do not force every service to
  have identical features.
