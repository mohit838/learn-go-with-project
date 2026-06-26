# Kong Gateway Notes

Kong is the API gateway for this project. It sits in front of the Go services and gives one public entry point for clients while keeping each service independent behind the gateway.

In this project, clients call Kong:

```text
Client -> Kong -> Go service
```

Current routes:

| Public URL | Kong route | Internal service |
| --- | --- | --- |
| `http://localhost:8000/auth` | `/auth` | `http://auth:8484` |
| `http://localhost:8000/tasks` | `/tasks` | `http://task-tracker:8485` |
| `http://localhost:8000/graphql` | `/graphql` | `http://task-tracker:8485` |
| `http://localhost:8000/expenses` | `/expenses` | `http://expense-tracker:8486` |

The config lives in [`kong/kong.yml`](../kong/kong.yml).

## Why Use Kong

Without a gateway, every client needs to know every service URL:

```text
Client -> auth service
Client -> task tracker service
Client -> expense tracker service
```

That becomes painful when services grow. Mobile apps, frontend apps, and external clients would need to track many hosts, ports, auth rules, rate limits, and API versions.

With Kong, clients only need one base URL:

```text
http://localhost:8000
```

Kong handles where each request goes.

## DB-less Mode

This project uses Kong in DB-less mode:

```yaml
KONG_DATABASE: "off"
KONG_DECLARATIVE_CONFIG: /kong/declarative/kong.yml
```

That means Kong does not need its own database. Instead, it reads all route and service configuration from `kong/kong.yml`.

This is good for local development and early microservice projects because:

- The setup is simple.
- Routes are version-controlled.
- No Kong database migrations are needed.
- Docker Compose can start the full gateway setup quickly.

For a larger production platform, you may later move to database-backed Kong or Kong Konnect, but DB-less mode is a strong starting point.

## Services And Routes

In Kong, a **service** is the upstream app Kong forwards traffic to.

Example:

```yaml
- name: auth-service
  url: http://auth:8484
```

This means Kong can reach the `auth` container on port `8484`.

A **route** tells Kong which public path should map to that service.

Example:

```yaml
routes:
  - name: auth-route
    paths:
      - /auth
    strip_path: true
```

Now this request:

```text
GET http://localhost:8000/auth
```

is forwarded to:

```text
GET http://auth:8484/
```

because `strip_path: true` removes `/auth` before sending the request to the service.

## Real Scenario: Frontend App

Imagine a frontend dashboard needs login, tasks, and expenses.

Without Kong, the frontend might need:

```text
AUTH_API_URL=http://localhost:8484
TASK_API_URL=http://localhost:8485
EXPENSE_API_URL=http://localhost:8486
```

With Kong, the frontend can use:

```text
API_URL=http://localhost:8000
```

Then it calls:

```text
POST /auth/login
GET /tasks
GET /expenses
```

The frontend stays simple while Kong handles service routing.

## Real Scenario: Adding A New Service

Suppose you add a notification service running on port `8487`.

In `docker-compose.yml`, you would add:

```yaml
notification:
  build:
    context: ./services/notification
  image: learn-go/notification:local
  env_file:
    - ./services/notification/.env
  ports:
    - "8487:8487"
  networks:
    - app-network
  restart: unless-stopped
```

Then add this to `kong/kong.yml`:

```yaml
- name: notification-service
  url: http://notification:8487
  routes:
    - name: notification-route
      paths:
        - /notifications
      strip_path: true
```

Now clients can call:

```text
GET http://localhost:8000/notifications
```

Kong forwards the request to:

```text
GET http://notification:8487/
```

## Real Scenario: API Versioning

If the task service later has a v2 API, Kong can route by version path.

Example:

```yaml
- name: task-tracker-v1
  url: http://task-tracker:8485
  routes:
    - name: task-tracker-v1-route
      paths:
        - /v1/tasks
      strip_path: true

- name: task-tracker-v2
  url: http://task-tracker-v2:8495
  routes:
    - name: task-tracker-v2-route
      paths:
        - /v2/tasks
      strip_path: true
```

Clients can move from:

```text
GET /v1/tasks
```

to:

```text
GET /v2/tasks
```

without changing the whole gateway architecture.

## Real Scenario: Authentication At The Gateway

Right now, Kong only routes traffic. Later, you can ask Kong to protect some routes.

For example, public routes might be:

```text
POST /auth/login
POST /auth/register
```

Protected routes might be:

```text
GET /tasks
POST /expenses
```

Kong can run plugins before the request reaches a service. Common authentication plugins include:

- JWT
- Key Auth
- OAuth2
- OpenID Connect, usually in enterprise or Konnect setups

Example idea:

```text
Client -> Kong checks token -> service receives trusted request
```

This keeps repeated security checks out of every service. The service should still validate important business rules, but Kong can handle common edge security.

Current project shape:

- Auth login/register remains public through `/auth`.
- Task-tracker accepts gateway identity headers when Kong provides them:
  `X-User-ID`, `X-Tenant-ID`, `X-Tenant-Slug`, and `X-User-Role`.
- Until Kong JWT/OIDC is configured, task-tracker keeps a Bearer-token fallback
  for local development.
- The task dashboard is exposed through `/graphql` and still checks the
  `superadmin` business rule inside the service.

Important: browser clients should not be allowed to spoof identity headers.
Only Kong or another trusted internal gateway should set those headers.

## Real Scenario: Rate Limiting

Imagine someone calls:

```text
GET /expenses
```

thousands of times per minute. Without gateway protection, the expense service takes the full hit.

With Kong rate limiting, you can set rules like:

```text
100 requests per minute per consumer
```

Then the flow becomes:

```text
Client -> Kong checks rate limit -> service only receives allowed traffic
```

This is useful for:

- Protecting services from abuse.
- Preventing accidental frontend loops from overwhelming APIs.
- Giving different limits to different clients later.

This project uses Kong route-level rate limits:

| Route | Limit | Why |
| --- | --- | --- |
| `/auth` | 60 requests per minute per IP | Public login/register routes need tighter protection. |
| `/tasks` | 300 requests per minute per IP | Normal API traffic during local development. |
| `/expenses` | 300 requests per minute per IP | Normal API traffic during local development. |

Auth also keeps a service-level limiter on `POST /register` and `POST /login`.
That gives sensitive endpoints a final guard even if gateway config changes.

## Real Scenario: Browser CORS

Kong handles browser CORS at the gateway so frontend apps can use one public API
base URL:

```text
http://localhost:8000
```

Allowed local origins:

```text
http://localhost:3000
http://localhost:5173
```

The current CORS plugin allows common API methods, `Authorization`,
`Content-Type`, and `X-Request-ID`. Add staging or production frontend origins
to `kong/kong.yml` before exposing those environments.

## Real Scenario: Logging And Observability

Kong can log every request before it reaches a service.

Useful questions Kong can help answer:

- Which service gets the most traffic?
- Which route has the most errors?
- Which clients are calling the API too much?
- How long do upstream services take to respond?

For local development, this project sends Kong logs to Docker output:

```yaml
KONG_PROXY_ACCESS_LOG: /dev/stdout
KONG_PROXY_ERROR_LOG: /dev/stderr
```

You can watch logs with:

```sh
make logs
```

## Common Commands

Start only Kong for local service development:

```sh
make gateway-dev
```

Run services in local terminals:

```sh
cd services/auth && go run ./cmd/api
cd services/task-tracker && go run ./cmd/api
cd services/expense-tracker && go run ./cmd/api
```

In this mode, Kong uses [`kong/kong.dev.yml`](../kong/kong.dev.yml) and forwards traffic to `host.docker.internal`.

Start the full container stack:

```sh
make up
```

Show running containers:

```sh
make ps
```

View logs:

```sh
make logs
```

Test routes:

```sh
curl http://localhost:8000/auth
curl http://localhost:8000/tasks
curl http://localhost:8000/expenses
```

Kong admin API:

```sh
curl http://localhost:8001/services
curl http://localhost:8001/routes
```

## Things To Remember

- Kong is the public API entry point.
- Go services stay private behind Docker networking.
- `kong/kong.yml` defines gateway routes.
- `strip_path: true` removes the public prefix before forwarding.
- DB-less mode means Kong config is stored in Git, not in a Kong database.
- Plugins can later add auth, rate limits, logging, request transforms, and more.

## Current Project Recommendation

For this project, keep Kong responsible for edge concerns:

- Use path routing only.
- Use gateway-level CORS for browser clients.
- Use gateway-level rate limits for public traffic.
- Keep each service responsible for its own business logic.
- Keep app-level guards on sensitive routes such as login/register.
- Add more routes as new services are created.

This keeps the project easy to understand while still using a gateway structure that can grow.
