# Development And Production Usage

This project has two practical run modes:

- Development: run Go services locally, then run only a gateway in Docker.
- Production-style: run services and gateway together through Docker Compose.

The current production-style stack is still local-friendly. Before using it as a
real production deployment, move secrets out of `.env` files, lock down gateway
admin ports, add TLS, and add observability.

## Quick Checks

Validate development gateway compose files:

```sh
make dev-check
```

Validate the production-style compose file:

```sh
make prod-check
```

Run backend tests:

```sh
make test
```

Build and lint the frontend:

```sh
cd frontend
npm run build
npm run lint
```

## Development Mode

Use this when you want fast feedback while editing Go services.

Start the Go services in separate terminals:

```sh
cd services/auth && go run ./cmd/api
cd services/task-tracker && go run ./cmd/api
cd services/expense-tracker && go run ./cmd/api
cd services/notification-service && go run ./cmd/api
```

Start the default development gateway with Kong:

```sh
make dev-up
```

Development URLs:

```text
Kong proxy:       http://localhost:8000
Auth service:     http://localhost:8000/auth
Task tracker:     http://localhost:8000/tasks
Expense tracker:  http://localhost:8000/expenses
Notifications:    http://localhost:8000/notifications
Kong admin API:   http://localhost:8001
Zipkin tracing:   http://localhost:9411
```

Stop the gateway:

```sh
make dev-down
```

Follow logs:

```sh
make dev-logs
```

Show containers:

```sh
make dev-ps
```

## Development Seed Users

Auth migrations seed these demo tenant users for local testing:

```text
tenant_slug: demo

superadmin@example.com / admin123
admin@example.com      / admin123
owner@example.com      / admin123
staff@example.com      / admin123
guest@example.com      / admin123
```

Roles are `superadmin`, `admin`, `owner`, `staff`, and `guest`. Task ownership
for edit/delete still comes from the task row's `user_id`.

## Frontend

The frontend reads `VITE_API_URL`.

Use Kong:

```env
VITE_API_URL=http://localhost:8000
```

Run the frontend:

```sh
cd frontend
npm run dev
```

Restart Vite after changing `.env`.

## Production-Style Mode

Use this when you want the full stack in Docker.

Validate first:

```sh
make prod-check
```

Build images:

```sh
make prod-build
```

Start the full stack:

```sh
make prod-up
```

Production-style URLs:

```text
Kong proxy:       http://localhost:8000
Auth service:     http://localhost:8000/auth
Task tracker:     http://localhost:8000/tasks
Expense tracker:  http://localhost:8000/expenses
Kong admin API:   http://localhost:8001
Zipkin tracing:   http://localhost:9411
```

## Gateway Auth Contract

Auth handles login/register and issues JWTs. The gateway validates access tokens
for protected service routes. Downstream services trust only these gateway-set
headers:

```text
X-User-ID
X-Tenant-ID
X-Tenant-Slug
X-User-Role
```

Task-tracker does not re-validate browser JWTs. It keeps business checks, such
as `superadmin` dashboard access. Use gRPC only for internal service-to-service
data, for example task-tracker asking auth for user stats.

## Gateway Observability

Kong sends gateway traces to Zipkin in local dev/prod-style compose:

```text
http://localhost:9411
```

Run one gateway stack at a time if you use the default Zipkin port. If you need
to customize ports, change the compose file's host port mapping.

Show containers:

```sh
make prod-ps
```

Follow logs:

```sh
make prod-logs
```

Restart:

```sh
make prod-restart
```

Stop:

```sh
make prod-down
```

## Migrations

Create migration files:

```sh
make migrate-make service=auth name=create_users_table
```

Run migrations:

```sh
make migrate-up service=auth
make migrate-up service=task-tracker
make migrate-up service=expense-tracker
```

Check migration status:

```sh
make migrate-status service=auth
```

Rollback latest batch:

```sh
make migrate-rollback service=auth
```

## Readiness Notes

- `make dev-check` and `make prod-check` validate Compose syntax, not database
  credentials or service health.
- Run service health endpoints after startup:
  - `GET /health`
  - `GET /ready`
  - `GET /health/postgres`
  - `GET /health/redis`
  - `GET /health/mongo`
  - `GET /health/minio`
- Keep gateway admin ports local only:
  - Kong Admin API: `8001`
