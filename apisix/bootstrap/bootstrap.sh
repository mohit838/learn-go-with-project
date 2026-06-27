#!/bin/sh
set -eu

ADMIN_URL="http://apisix:9180/apisix/admin"
ADMIN_KEY="dev_apisix_admin_key"

until curl -fsS "$ADMIN_URL/routes" -H "X-API-KEY: $ADMIN_KEY" >/dev/null; do
  sleep 1
done

curl -fsS -X PUT "$ADMIN_URL/consumers/auth_service" \
  -H "X-API-KEY: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  --data-binary "@/bootstrap/consumers/auth_service.json"

curl -fsS -X PUT "$ADMIN_URL/global_rules/1" \
  -H "X-API-KEY: $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  --data-binary "@/bootstrap/global_rules/1.json"

for route in /bootstrap/routes/*.json; do
  id="$(basename "$route" .json)"
  curl -fsS -X PUT "$ADMIN_URL/routes/$id" \
    -H "X-API-KEY: $ADMIN_KEY" \
    -H "Content-Type: application/json" \
    --data-binary "@$route"
done

printf "APISIX GUI gateway bootstrap complete\n"
