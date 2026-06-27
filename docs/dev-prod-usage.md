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

## Development With APISIX

APISIX is available beside Kong for learning gateway behavior.

Standalone APISIX, YAML configured:

```sh
make apisix-dev
```

Standalone APISIX URLs:

```text
APISIX proxy:     http://localhost:9080
Auth service:     http://localhost:9080/auth
Task tracker:     http://localhost:9080/tasks
Task GraphQL:     http://localhost:9080/tasks/graphql
Expense tracker:  http://localhost:9080/expenses
Zipkin tracing:   http://localhost:9411
```

APISIX with etcd and Dashboard:

```sh
make apisix-gui-up
```

APISIX GUI URLs:

```text
APISIX GUI proxy:   http://localhost:9088
APISIX Dashboard:   http://localhost:9181
APISIX Admin API:   http://localhost:9180
Zipkin tracing:     http://localhost:9411
```

Dashboard login:

```text
username: admin
password: admin
```

The official APISIX Dashboard plugin catalog can be unstable with the available
dashboard image. For local plugin editing, use:

```text
apisix/plugin-manager.html
```

That helper is development-only because it uses the local Admin API key in the
browser.

## Frontend

The frontend reads `VITE_API_URL`.

Use Kong:

```env
VITE_API_URL=http://localhost:8000
```

Use standalone APISIX:

```env
VITE_API_URL=http://localhost:9080
```

Use APISIX GUI mode:

```env
VITE_API_URL=http://localhost:9088
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

Kong and APISIX both send gateway traces to Zipkin in local dev/prod-style
compose:

```text
http://localhost:9411
```

Run one gateway stack at a time if you use the default Zipkin port. If you need
Kong and APISIX running together, change one compose file's host port mapping.

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
  - APISIX Admin API: `9180`
  - APISIX Dashboard: `9181`
