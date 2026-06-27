# APISIX Gateway Notes

APISIX is available as an optional gateway beside Kong. This is for learning
and comparison; it does not replace the existing Kong setup.

## Why Separate

Kong and APISIX both want to act as the public gateway. To avoid config and port
conflicts, APISIX has its own files:

```text
docker-compose.apisix.yml
apisix/
  config.yaml
  apisix.dev.yaml
  README.md
```

Kong stays on:

```text
http://localhost:8000
```

APISIX uses:

```text
http://localhost:9080
```

## Run APISIX Locally

Run Auth, Task Tracker, and Expense Tracker locally in terminal tabs, then start
only APISIX in Docker:

```sh
make apisix-dev
```

Stop it with:

```sh
make apisix-dev-down
```

Follow logs:

```sh
make apisix-dev-logs
```

## Image And Mode

This dev setup uses:

```text
apache/apisix:3.17.0-debian
```

It runs APISIX in standalone YAML mode, so it does not need:

```text
gcr.io/etcd-development/etcd:v3.6.12
curlimages/curl:8.11.1
```

Those images are useful for an APISIX Admin API + etcd setup. We can add that
as a separate compose file later, but this first APISIX gateway stays simple and
Git-configured like the current Kong DB-less setup.

## Local URLs

```text
APISIX proxy:     http://localhost:9080
Auth service:     http://localhost:9080/auth
Task tracker:     http://localhost:9080/tasks
Task GraphQL:     http://localhost:9080/tasks/graphql
Expense tracker:  http://localhost:9080/expenses
```

## Frontend

The frontend reads `VITE_API_URL`.

Use Kong:

```env
VITE_API_URL=http://localhost:8000
```

Use APISIX:

```env
VITE_API_URL=http://localhost:9080
```

Restart the Vite dev server after changing this value.

## Current APISIX Behavior

- Routes `/auth` to Auth and strips the `/auth` prefix.
- Routes `/tasks` and `/tasks/graphql` to Task Tracker.
- Routes `/expenses` to Expense Tracker and strips the `/expenses` prefix.
- Handles browser CORS at gateway level.
- Applies local rate limits.
- Validates Auth JWTs for task routes.
- Forwards trusted identity headers to Task Tracker:
  - `X-User-ID`
  - `X-Tenant-ID`
  - `X-Tenant-Slug`
  - `X-User-Role`

Auth `JWT_SECRET` must match the APISIX `jwt-auth` consumer secret in
`apisix/apisix.dev.yaml`.

Auth tokens use the issuer claim:

```json
{ "iss": "auth-service" }
```

So protected APISIX routes must configure:

```yaml
jwt-auth:
  key_claim_name: iss
  claims_to_verify:
    - exp
```

Without `key_claim_name: iss`, APISIX defaults to looking for a claim named
`key` and returns `401` with `missing user key in JWT token`.
