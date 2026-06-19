# Backend Structure Study Guide

## Request Flow

Each CRUD request follows one direction:

```text
HTTP request -> handler -> service -> repository -> PostgreSQL
```

The handler owns HTTP: route parameters, JSON decoding, status codes, and the
shared response envelope. The service owns business rules: validation, default
values, and password hashing. The repository owns SQL only. This separation
keeps database details out of HTTP code and makes business rules testable.

## DTOs

DTO means Data Transfer Object. Request DTOs describe input accepted from the
client, such as `CreateRequest`. Response DTOs describe only data safe to return
to the client, such as `Response`.

Auth deliberately has a private `passwordHash` field. The client sends a
password in `CreateRequest`; the service hashes it with bcrypt; PostgreSQL stores
only `password_hash`; the response never contains either value.

## Shared Table Fields

The three first tables use `is_active`, `created_at`, and `updated_at`.

- `is_active` supports disabling a record without deleting it.
- `created_at` records when it was first created.
- `updated_at` records the most recent change.

These are useful defaults, not a rule. Do not add `is_active` when a record is
either permanent or should use an explicit lifecycle state. Do not add timestamps
to short-lived join tables unless you need audit history.

## Migrations

`000001` establishes the migration sequence. `000002` creates the first real
tables. Apply them before starting CRUD:

```sh
make migrate-up MIGRATION_SERVICE=auth
make migrate-up MIGRATION_SERVICE=task-tracker
make migrate-up MIGRATION_SERVICE=expense-tracker
```

Never edit a migration that has already run in a shared environment. Create the
next numbered migration instead.

## CRUD Routes

| Service | Gateway collection route |
| --- | --- |
| Auth users | `http://localhost:8000/auth/users` |
| Tasks | `http://localhost:8000/tasks/tasks` |
| Expenses | `http://localhost:8000/expenses/expenses` |

Each collection supports `POST` and `GET`. Each item route with `/{id}` supports
`GET`, `PUT`, and `DELETE`. Successful responses use `{ "data": ... }`; errors
use `{ "error": { "code": "...", "message": "..." } }`.

## Structured Logs

The request logger writes JSON with the request ID, method, path, status, byte
count, and duration. JSON logs are machine-readable, so a log collector can
filter requests such as `status >= 500` without parsing human text.

## Redis And MongoDB Next

Redis should cache read-heavy data, beginning with `GET /users/{id}`. The cache
must be invalidated after update or delete. MongoDB audit logs should be a
separate write model: store an `audit_logs` document with `service`, `action`,
`method`, `path`, `status`, `request_id`, and `created_at`.

Use one Mongo database named `appdb` for this learning project and include the
service name in every document. Split into three databases only when services
need independent retention policies, access control, backups, or deployment
ownership.
