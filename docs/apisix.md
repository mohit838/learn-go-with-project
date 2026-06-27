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

## Dashboard Mode

Dashboard mode uses a separate compose file:

```text
docker-compose.apisix-gui.yml
```

It includes:

```text
apache/apisix:3.11.0-debian
gcr.io/etcd-development/etcd:v3.6.12
apache/apisix-dashboard:3.0.1-alpine
curlimages/curl:8.11.1
```

Dashboard mode uses APISIX `3.11.0` on purpose. The available dashboard image
is older, and its plugin management pages can crash or return `data:null` when
paired with newer APISIX gateway builds such as `3.17.0`. Standalone APISIX
still uses the newer image because it is YAML-driven and does not depend on the
dashboard UI.

Start it:

```sh
make apisix-gui-up
```

Stop it:

```sh
make apisix-gui-down
```

Dashboard mode URLs:

```text
APISIX GUI proxy:   http://localhost:9088
APISIX Dashboard:   http://localhost:9181
APISIX Admin API:   http://localhost:9180
```

Dashboard login for local development:

```text
username: admin
password: admin
```

This mode stores routes, consumers, and plugins in etcd. The dashboard can then
add, edit, or delete routes without changing YAML.

The official Dashboard image can still be unstable on the plugin catalog pages.
If `/apisix/admin/plugins?all=true` returns `data:null` and the browser crashes
with `Cannot read properties of null`, use the local development helper instead:

```text
apisix/plugin-manager.html
```

That helper talks directly to the local APISIX Admin API and can enable common
route plugins such as CORS, JWT auth, and `limit-count`. Keep it local only; do
not expose the Admin API key to a real frontend or public network.

Dashboard mode intentionally does not pin a tiny plugin allow-list in
`apisix/config-gui.yaml`. Letting APISIX expose its normal plugin catalog avoids
Dashboard pages receiving incomplete plugin metadata.

If Dashboard shows this validation error:

```text
redis_host is required
```

check the `limit-count` plugin config. For local development use `policy:
local`. Only use `policy: redis` when `redis_host` and the related Redis fields
are configured.

The bootstrap container seeds the same local routes we use elsewhere:

```text
/auth
/tasks
/tasks/graphql
/expenses
/notifications
```

For frontend testing with dashboard mode:

```env
VITE_API_URL=http://localhost:9088
```

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
