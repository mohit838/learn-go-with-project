# Upcoming Learning Tasks

This is the recommended order for growing this project. Complete one milestone
and test it before starting the next one.

## 1. Application Foundation (Completed)

- Add an `internal/constants` package for shared application constants.
- Add a small `internal/utils` package only when a genuinely reusable helper appears.
- Define a consistent JSON response shape for successful data and errors.
- Add a central API route list, for example `internal/router/app_api.go`, so all
  endpoint registration is easy to find.

**Done when:** every endpoint returns the same response envelope and routes are
registered from one clear place per service.

## 2. Structured Logging (Completed)

- Replace `fmt.Println` and the default logger with structured logs.
- Include request ID, service name, method, path, status, duration, and error.
- Configure the log level with `APP_LOG_LEVEL`.

**Done when:** a failed request can be traced from the gateway log to the Go
service log.

## 3. Database Migrations (Completed)

- Add a migration tool and a `cmd/migrate/main.go` entry point per service.
- Add `migrations/` folders with ordered SQL files such as
  `000001_create_users.up.sql` and `000001_create_users.down.sql`.
- Add Makefile commands to run migrations up and down.

**Done when:** an empty PostgreSQL database can be created entirely from the
migration files, with no manual SQL steps.

## 4. Simple User CRUD in Auth (Completed)

- Create a `users` table with migration files.
- Add domain, application, infrastructure, transport, request validation, and routes.
- Implement create, list, get by ID, update, and delete endpoints.
- Add application, transport, and repository tests.

**Done when:** the Auth service can manage users through Kong at
`http://localhost:8000/auth`.

## 5. Redis Connection and First Use Case (Completed)

Use Redis after CRUD works. Redis should have a purpose: cache data, store
short-lived sessions, or keep rate-limit counters. Do not use it as a second
database for user records.

1. Run a local Redis server on port `1236`, matching your service `.env` files.
2. Add `github.com/redis/go-redis/v9` to each service that needs Redis.
3. Create `internal/database/redis.go` with a function that parses `REDIS_URL`,
   creates a client, and verifies it with `PING` using a context timeout.
4. Create the client in `cmd/api/main.go`, close it on shutdown, and pass it
   only to the layer that needs it.
5. Start with caching `GET /users/{id}` in Auth. Use a key such as
   `auth:user:{id}`, a short TTL, and delete that key when the user changes or
   is deleted.
6. Add a test for cache hit, cache miss, and cache invalidation.

For local development, a standard Redis instance normally has no username or
password. A suitable local value is `REDIS_URL=redis://localhost:1236/0`.
Production credentials belong only in the real `.env` or secret store, never in
Git.

**Done when:** a second user lookup reads from Redis, and updates correctly
remove the stale cache value.

## 6. CORS and Rate Limiting (Completed)

- Configure CORS centrally in Kong for the local browser origin
  `http://localhost:3000`.
- Add a Kong local rate limit of 60 requests per minute to each public service.
- Switch the policy to Redis when multiple Kong instances need shared counters.
- Keep business authorization inside the Go services; gateway rate limits are
  not authorization.

**Done when:** a browser origin is allowed intentionally and repeated requests
receive `429 Too Many Requests` at the configured limit.

## 7. MongoDB Practice Module (Completed)

- Add MongoDB only for a document-oriented use case, such as audit events,
  activity feeds, or flexible user preferences.
- Create a dedicated Mongo connection package with context timeout, ping, and
  graceful close handling.
- Do not duplicate the PostgreSQL `users` table in MongoDB.

**Done when:** one clearly defined document use case is stored and queried from
MongoDB without affecting PostgreSQL ownership.

## 8. External Avatar API (Completed)

- Create an HTTP client with timeouts, context propagation, and typed request
  and response structs.
- Auth exposes `GET /users/{id}/avatar`. It uses the user's username as a seed
  for the configured provider and remains disabled until `AVATAR_API_URL` is set.
- Handle provider failures gracefully and return your standard error shape.
- Add tests using `httptest`; do not call the real provider in tests.

**Done when:** the Auth service can safely use an avatar provider and still
respond predictably when the provider is unavailable.

## 9. gRPC Between Services (Completed)

- Initial learning version is now added:
  - Auth exposes `CheckUser` and `UserStats` over gRPC.
  - Kong validates JWTs for task routes and forwards trusted identity headers.
  - Task-tracker trusts gateway identity for HTTP requests.
  - Task-tracker uses Auth gRPC user stats in the superadmin dashboard.
- Next hardening step: replace the temporary hand-written protobuf-compatible
  gRPC structs with generated protobuf contracts.
- Add deadlines, richer error mapping, and a local integration test.

**Done when:** one service-to-service request works through gRPC with a stable,
versioned contract.

## 10. Review and Harden

- `/health` reports liveness and `/ready` verifies PostgreSQL connectivity
  before reporting readiness.
- Add integration tests with disposable local dependencies.
- Document environment variables, migrations, and API examples for every
  completed milestone.

## 11. Observability Stack

Add observability after the services have real routes, database calls, cache
usage, and gateway traffic. Start small in local Docker, then shape a
production-friendly version.

- Add Prometheus metrics for HTTP request count, status, duration, and in-flight
  requests.
- Add Grafana dashboards for service health, route latency, error rate, and
  dependency checks.
- Add Loki for centralized structured logs from Kong and the Go services.
- Add distributed tracing with OpenTelemetry so one request can be followed
  through Kong and service code.
- Evaluate Elastic APM later if you want a richer APM UI, error grouping,
  traces, service maps, and searchable transaction data.
- Add Docker Compose services for local observability, such as Prometheus,
  Grafana, Loki, Promtail or Alloy, and an OpenTelemetry Collector.
- Add Makefile commands for developer workflows, for example
  `make observe-dev`, `make observe-dev-down`, `make observe-dev-logs`, and
  `make observe-dev-ps`.
- Add production-oriented Compose or deployment notes with persistent volumes,
  retention settings, authentication, and resource limits.

**Done when:** a request through Kong can be viewed in metrics, logs, and traces,
and the local observability stack can be started with one Make command.
