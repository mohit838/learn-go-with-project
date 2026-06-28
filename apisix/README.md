# APISIX Gateway

This folder is a side-by-side APISIX gateway setup for learning. It does not replace the existing Kong config.

APISIX runs in standalone YAML mode for local development using `apache/apisix:3.17.0-debian`:

```bash
make apisix-dev
```

Local APISIX URLs:

```text
APISIX proxy:     http://localhost:9080
Auth service:     http://localhost:9080/auth
Task tracker:     http://localhost:9080/tasks
Expense tracker:  http://localhost:9080/expenses
Task GraphQL:     http://localhost:9080/tasks/graphql
Zipkin tracing:   http://localhost:9411
```

The config mirrors the Kong dev gateway and forwards to locally running Go services through `host.docker.internal`.

This mode does not need etcd or a curl bootstrap container. If we want to learn the Admin API flow later, add it as a separate compose file so this standalone setup remains clean.

Auth tokens use `iss=auth-service`, so task routes configure APISIX `jwt-auth`
with `key_claim_name: iss`.

## Gateway Responsibility

APISIX is responsible for edge concerns:

- CORS
- rate limiting
- JWT validation
- trusted identity headers
- Zipkin tracing

Auth issues tokens. APISIX validates those tokens for protected routes and then
forwards:

```text
X-User-ID
X-Tenant-ID
X-Tenant-Slug
X-User-Role
```

Task-tracker trusts those headers because only the gateway should be exposed to
frontend traffic.

For the frontend, switch the API base URL from:

```text
http://localhost:8000
```

to:

```text
http://localhost:9080
```

Keep Kong on `8000` and APISIX on `9080` so both gateways can exist without port conflicts.

## Dashboard Mode

Run APISIX with etcd, Admin API, Dashboard UI, and a bootstrap container:

```bash
make apisix-gui-up
```

Local APISIX GUI URLs:

```text
APISIX GUI proxy:   http://localhost:9088
APISIX Dashboard:   http://localhost:9181
APISIX Admin API:   http://localhost:9180
Zipkin tracing:     http://localhost:9411
```

Dashboard login for local development:

```text
username: admin
password: admin
```

Use `VITE_API_URL=http://localhost:9088` when the frontend should call this
dynamic APISIX gateway.

Dashboard mode intentionally uses `apache/apisix:3.11.0-debian` with
`apache/apisix-dashboard:3.0.1-alpine`. The dashboard image is older than the
latest APISIX gateway image, and the plugin management pages can return
`data:null` or crash when paired with newer APISIX versions. Standalone mode can
stay on newer APISIX because it does not rely on the dashboard UI.

If the official Dashboard plugin pages crash, use the local helper:

```text
apisix/plugin-manager.html
```

It talks directly to the local APISIX Admin API at
`http://localhost:9180/apisix/admin` and can enable route-level presets for
CORS, JWT auth, and local rate limiting. This helper is for local development
only because it uses the Admin API key in the browser.

## Zipkin

Both standalone and Dashboard modes start Zipkin:

```text
http://localhost:9411
```

APISIX sends gateway spans through the `zipkin` plugin. If you run multiple
gateway stacks at the same time, only one can bind host port `9411` unless you
change the compose port mapping.
