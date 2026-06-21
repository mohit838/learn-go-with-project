# OpenAPI Documentation

Each Go service owns its API contract. The OpenAPI files are the source of
truth for API documentation and should be updated in the same pull request as
any route, request, or response change.

| Service | Contract | Direct local URL | Kong gateway URL |
| --- | --- | --- | --- |
| Auth | [`services/auth/docs/openapi.yaml`](../services/auth/docs/openapi.yaml) | `http://localhost:8484` | `http://localhost:8000/auth` |
| Task tracker | [`services/task-tracker/docs/openapi.yaml`](../services/task-tracker/docs/openapi.yaml) | `http://localhost:8485` | `http://localhost:8000/tasks` |
| Expense tracker | [`services/expense-tracker/docs/openapi.yaml`](../services/expense-tracker/docs/openapi.yaml) | `http://localhost:8486` | `http://localhost:8000/expenses` |

## Development

Start Kong and run the Go services locally:

```sh
make gateway-dev

cd services/auth && go run ./cmd/api
cd services/task-tracker && go run ./cmd/api
cd services/expense-tracker && go run ./cmd/api
```

In Swagger UI, choose the **Kong gateway** server for an end-to-end check. Its
default `gatewayUrl` value is `http://localhost:8000`, so the service prefix is
included automatically. Choose the **Direct local service** server only when
debugging a service without Kong.

Use one of the service-specific Make targets to preview a contract locally in
Swagger UI:

```sh
make swagger-auth            # http://localhost:8081
make swagger-task-tracker    # http://localhost:8082
make swagger-expense-tracker # http://localhost:8083
```

Each command runs Swagger UI in the foreground; stop it with `Ctrl+C`. Select
the Kong server, then use **Try it out**. For example, the Auth health check is sent to
`http://localhost:8000/auth/`; Kong removes `/auth` before forwarding the
request to the service's `/` route.

These commands render the committed OpenAPI YAML; they do not generate it from
Go code. Update the owning `docs/openapi.yaml` file whenever the service API
changes.

## Generating Swagger from Go Annotations

If a service uses [swaggo](https://github.com/swaggo/swag) annotations, generate
its Swagger files with a service-specific command:

```sh
make swagger-generate-auth
make swagger-generate-task-tracker
make swagger-generate-expense-tracker

# Generate files for every service.
make swagger-generate
```

These targets run `swag init -g cmd/api/main.go -o docs` from the relevant
service directory. Install the `swag` command once before using them:

```sh
go install github.com/swaggo/swag/cmd/swag@latest
```

`swag` creates Swagger 2.0 output such as `docs/swagger.yaml` and
`docs/swagger.json`; it does not update this project's OpenAPI 3
`docs/openapi.yaml` files. Keep the OpenAPI files as the versioned API contract
unless the project deliberately migrates fully to annotation-generated Swagger.

## Staging

Deploy the same committed OpenAPI files with the service release. In the
Swagger UI server selector, choose the **Kong gateway** server and change
`gatewayUrl` from `http://localhost:8000` to the staging public gateway URL,
for example `https://api.staging.example.com`. The resulting Auth URL is
`https://api.staging.example.com/auth`.

Do not publish direct service ports in staging. Clients and documentation
should use Kong's public URL and its service prefixes (`/auth`, `/tasks`, and
`/expenses`). This keeps staging aligned with production routing and avoids
exposing internal service addresses.

For a permanent staging documentation site, serve Swagger UI as a static
container or static site and configure it to load the versioned OpenAPI file
from the same release. Restrict "Try it out" access if the staging API is not
intended for everyone.

## Keeping Contracts Correct

Before merging an API change:

1. Update the owning service's `docs/openapi.yaml` file.
2. Ensure every documented method and path matches `internal/router/router.go`.
3. Lint the contract with a pinned OpenAPI linter in CI, such as Redocly CLI.
4. Compare the branch contract with the target branch to flag breaking changes
   (removed paths, operations, or response fields).

The current endpoints return plain text. When the services move to structured
JSON responses, update both the handler and the response schemas together.
