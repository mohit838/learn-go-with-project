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
```

The config mirrors the Kong dev gateway and forwards to locally running Go services through `host.docker.internal`.

This mode does not need etcd or a curl bootstrap container. If we want to learn the Admin API flow later, add it as a separate compose file so this standalone setup remains clean.

Auth tokens use `iss=auth-service`, so task routes configure APISIX `jwt-auth`
with `key_claim_name: iss`.

For the frontend, switch the API base URL from:

```text
http://localhost:8000
```

to:

```text
http://localhost:9080
```

Keep Kong on `8000` and APISIX on `9080` so both gateways can exist without port conflicts.
