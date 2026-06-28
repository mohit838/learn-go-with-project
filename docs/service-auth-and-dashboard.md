# Service Auth, gRPC, And Dashboard

## Request Flow

The service shape is:

```text
Client -> Gateway -> Auth issues JWT -> Gateway validates JWT -> Service
```

Auth authenticates the user during login/register and issues tokens. After that,
APISIX or Kong validates the access token at the gateway and forwards trusted
identity headers to downstream services.

Task-tracker does not validate client Bearer tokens directly. It trusts the
gateway identity headers and keeps only business authorization checks, such as
`superadmin` dashboard access.

## Current Behavior

Auth service:

- Handles `/register` and `/login`.
- Issues access and refresh tokens.
- Runs an internal gRPC server on `GRPC_PORT`.
- Exposes gRPC checks for user validity and user dashboard counts.

Task-tracker service:

- Requires trusted gateway identity headers:
  - `X-User-ID`
  - `X-Tenant-ID`
  - `X-Tenant-Slug`
  - `X-User-Role`
- Exposes a superadmin-only `/tasks/graphql` dashboard endpoint.
- Uses Auth gRPC for internal service-to-service calls, such as dashboard user
  stats.

Notification service:

- Requires the same trusted gateway identity headers for HTTP routes.
- Demonstrates goroutines and channels with an in-memory worker pool.
- Exposes internal gRPC methods for service-to-service notification calls.

## Important Security Rule

Only APISIX, Kong, or another trusted internal gateway should set identity headers.
Frontend clients must not be allowed to spoof `X-User-ID`, `X-Tenant-ID`, or
`X-User-Role`.

The gateway validates the token and then forwards trusted identity claims to
upstream services. Services should still keep business authorization checks, such as
`superadmin` access for dashboards.

## Gateway Switch

Both APISIX and Kong follow the same protected-route contract:

```text
Authorization: Bearer <access_token>
Gateway validates JWT signature and exp
Gateway forwards X-User-ID, X-Tenant-ID, X-Tenant-Slug, X-User-Role
Task-tracker handles the request
```

Switching gateway should not require service code changes. Change only the
frontend base URL and gateway config:

```env
VITE_API_URL=http://localhost:8000  # Kong
VITE_API_URL=http://localhost:9080  # APISIX standalone
VITE_API_URL=http://localhost:9088  # APISIX GUI mode
```

## gRPC

Auth gRPC methods:

```text
auth.v1.AuthService/CheckUser
auth.v1.AuthService/UserStats
```

The source contract lives in:

```text
proto/auth/v1/auth.proto
```

Notification gRPC methods:

```text
notification.v1.NotificationService/SubmitNotification
notification.v1.NotificationService/GetNotification
notification.v1.NotificationService/Stats
```

The source contract lives in:

```text
proto/notification/v1/notification.proto
```

The current implementation uses protobuf wire encoding with small typed
compatibility structs in each service's `internal/authrpc` package. This keeps
the code simple while `protoc` is not installed in the local environment.

Later, replace those compatibility structs with generated Go code:

```text
protoc generated Go code
typed generated clients and servers
shared versioned protobuf module/package
```

## GraphQL Dashboard

Endpoint:

```text
POST /tasks/graphql
```

Each service should own its GraphQL path while we are using Kong as a simple API
gateway:

```text
/tasks/graphql
/expenses/graphql
```

A single public `/graphql` should only be used later if we add a dedicated
GraphQL gateway or federation service that knows how to route and compose
schemas across services.

Schema file:

```text
services/task-tracker/graphql/schema.graphql
```

Example query:

```graphql
query {
  dashboard {
    users {
      total
      active
      inactive
      by_role { role count }
    }
    tasks {
      total
      active
      inactive
      by_status { status count }
      by_priority { priority count }
      by_user { user_id total active inactive }
    }
  }
}
```

Only `superadmin` can access this dashboard for now.

## Env

Auth:

```env
GRPC_PORT=8584
```

Task-tracker:

```env
AUTH_GRPC_ADDRESS=localhost:8584
```

Kong dev config must use the same JWT secret as Auth:

```yaml
consumers:
  - username: auth-service
    jwt_secrets:
      - key: auth-service
        algorithm: HS256
        secret: replace_with_a_long_random_secret
```

APISIX uses the same shared secret in its `jwt-auth` consumer config.

## Observability

Gateway tracing is handled at the gateway with Zipkin. Services do not need
Zipkin code for this first step.

Local Zipkin URL:

```text
http://localhost:9411
```

The gateway sends spans to:

```text
http://zipkin:9411/api/v2/spans
```
