# Service Auth, gRPC, And Dashboard

## Request Flow

The service shape is:

```text
Client -> Kong -> Auth validates user -> Kong forwards trusted identity -> Service
```

Task-tracker does not validate client Bearer tokens directly. Kong validates the
JWT and forwards trusted identity headers to task-tracker.

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

## Important Security Rule

Only Kong or another trusted internal gateway should set identity headers.
Frontend clients must not be allowed to spoof `X-User-ID`, `X-Tenant-ID`, or
`X-User-Role`.

Kong validates the token and then forwards trusted identity claims to upstream
services. Services should still keep business authorization checks, such as
`superadmin` access for dashboards.

## gRPC

Auth gRPC methods:

```text
auth.v1.AuthService/CheckUser
auth.v1.AuthService/UserStats
```

For this learning step, the project uses a small JSON gRPC codec instead of
generated protobuf files. Later, replace it with:

```text
proto/auth/v1/auth.proto
protoc generated Go code
typed generated clients and servers
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
