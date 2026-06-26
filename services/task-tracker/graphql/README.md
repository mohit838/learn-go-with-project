# Task Tracker GraphQL

Task-tracker GraphQL is exposed through Kong at:

```text
POST /tasks/graphql
```

This endpoint is for dashboard-style reads. REST remains the main API for task
CRUD.

## Files

```text
services/task-tracker/graphql/
  schema.graphql
  queries/
    dashboard.graphql
  README.md
```

## Auth

Call through Kong with the Auth service access token:

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

Kong validates the JWT and forwards trusted identity headers to task-tracker.
Only `superadmin` can use the dashboard query for now.

## Dashboard Query

Use the query in:

```text
services/task-tracker/graphql/queries/dashboard.graphql
```

HTTP body example:

```json
{
  "query": "query TaskDashboard { dashboard { users { total active inactive by_role { role count } } tasks { total active inactive by_status { status count } } } }"
}
```

## Frontend Usage

Frontend teams can copy queries from `queries/` and use `schema.graphql` for
type generation or editor autocomplete.

Later, if we add GraphQL for expense-tracker, use:

```text
/expenses/graphql
```

Do not use one global `/graphql` until there is a dedicated GraphQL gateway or
federation service.
