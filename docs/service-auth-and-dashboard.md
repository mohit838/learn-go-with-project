# Service Auth, gRPC, And Dashboard

## Request Flow

The intended production shape is:

```text
Client -> Kong -> Auth validates user -> Kong forwards trusted identity -> Service
```

For local development, task-tracker still accepts a Bearer access token directly.
That keeps the service easy to test before Kong gets full JWT or OIDC validation.

## Current Behavior

Auth service:

- Handles `/register` and `/login`.
- Issues access and refresh tokens.
- Runs an internal gRPC server on `GRPC_PORT`.
- Exposes gRPC checks for user validity and user dashboard counts.

Task-tracker service:

- Prefers trusted gateway identity headers:
  - `X-User-ID`
  - `X-Tenant-ID`
  - `X-Tenant-Slug`
  - `X-User-Role`
- Falls back to Bearer token verification for local development.
- Calls Auth gRPC after Bearer token verification to confirm the user is still
  active and still belongs to the tenant.
- Exposes a superadmin-only `/graphql` dashboard endpoint.

## Important Security Rule

Only Kong or another trusted internal gateway should set identity headers.
Frontend clients must not be allowed to spoof `X-User-ID`, `X-Tenant-ID`, or
`X-User-Role`.

When Kong auth is added, configure Kong to validate the token and then inject or
forward trusted identity claims to upstream services. Services should still keep
business authorization checks, such as `superadmin` access for dashboards.

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
POST /graphql
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
JWT_SECRET=replace_with_the_same_secret_used_by_auth
JWT_ISSUER=auth-service
```
