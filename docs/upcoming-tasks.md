# Upcoming Learning Tasks

This is the recommended order for growing this project. Complete one milestone
and test it before starting the next one.

## 1. Application Foundation

- Add an `internal/constants` package for shared application constants.
- Add a small `internal/utils` package only for genuinely reusable helpers.
- Define a consistent JSON response shape for successful data and errors.
- Add a central API route list, for example `internal/router/app_api.go`, so all
  endpoint registration is easy to find.

**Done when:** every endpoint returns the same response envelope and routes are
registered from one clear place per service.

## 2. Structured Logging

- Replace `fmt.Println` and the default logger with structured logs.
- Include request ID, service name, method, path, status, duration, and error.
- Configure the log level with `APP_LOG_LEVEL`.

**Done when:** a failed request can be traced from the gateway log to the Go
service log.

## 3. Database Migrations

- Add a migration tool and a `cmd/migrate/main.go` entry point per service.
- Add `migrations/` folders with ordered SQL files such as
  `000001_create_users.up.sql` and `000001_create_users.down.sql`.
- Add Makefile commands to run migrations up and down.

**Done when:** an empty PostgreSQL database can be created entirely from the
migration files, with no manual SQL steps.

## 4. Simple User CRUD in Auth

- Create a `users` table with migration files.
- Add model, repository, service, handler, request validation, and routes.
- Implement create, list, get by ID, update, and delete endpoints.
- Add handler and repository tests.

**Done when:** the Auth service can manage users through Kong at
`http://localhost:8000/auth`.

## 5. Redis Connection and First Use Case

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

## 6. CORS and Rate Limiting

- Configure CORS centrally in Kong for browser clients.
- Add Kong rate limiting for public endpoints first.
- Use a Redis-backed rate-limit policy when multiple Kong instances need shared
  counters.
- Keep business authorization inside the Go services; gateway rate limits are
  not authorization.

**Done when:** a browser origin is allowed intentionally and repeated requests
receive `429 Too Many Requests` at the configured limit.

## 7. MongoDB Practice Module

- Add MongoDB only for a document-oriented use case, such as audit events,
  activity feeds, or flexible user preferences.
- Create a dedicated Mongo connection package with context timeout, ping, and
  graceful close handling.
- Do not duplicate the PostgreSQL `users` table in MongoDB.

**Done when:** one clearly defined document use case is stored and queried from
MongoDB without affecting PostgreSQL ownership.

## 8. External Avatar API

- Create an HTTP client with timeouts, context propagation, and typed request
  and response structs.
- Call an avatar provider from Auth when a user is created, or expose a separate
  avatar endpoint.
- Handle provider failures gracefully and return your standard error shape.
- Add tests using `httptest`; do not call the real provider in tests.

**Done when:** the Auth service can safely use an avatar provider and still
respond predictably when the provider is unavailable.

## 9. gRPC Between Services

- Start only after REST CRUD and service boundaries are clear.
- Define a small `.proto` contract, generate Go code, and add a gRPC server to
  one service.
- Let another service call one read-only method first.
- Add deadlines, error mapping, and a local integration test.

**Done when:** one service-to-service request works through gRPC with a stable,
versioned contract.

## 10. Review and Harden

- Add health and readiness endpoints for PostgreSQL, Redis, and MongoDB where
  used.
- Add integration tests with disposable local dependencies.
- Document environment variables, migrations, and API examples for every
  completed milestone.
