# gRPC Between Auth And Task Tracker

gRPC is a way for services to call each other directly using a small, typed
contract. REST is still the public API in this project; gRPC is an internal
service-to-service API. The contract is written once in
[`proto/task/v1/task.proto`](../proto/task/v1/task.proto), then Go code is
generated for both services.

## The First Contract

`TaskService.GetTask` accepts a task ID and returns a task. It is deliberately
read-only: it is the smallest useful cross-service call and keeps ownership of
the `tasks` table inside Task Tracker.

```
Browser -> Kong -> Auth GET /auth/tasks/42
                         |
                         | gRPC GetTask(id=42), deadline: 2 seconds
                         v
                    Task Tracker :8487 -> PostgreSQL
```

Auth maps the gRPC result back into its normal JSON response envelope. If Task
Tracker says `NotFound`, callers receive HTTP `404`; an unavailable service or
a deadline failure becomes HTTP `502`. Auth never queries the Task Tracker
database directly.

## Running It Locally

1. Start Task Tracker with `GRPC_PORT=8487` (the default).
2. Set `TASK_GRPC_ADDR=localhost:8487` in Auth's `.env`.
3. Start Auth, then request `GET http://localhost:8000/auth/tasks/42` through
   Kong, or `GET http://localhost:8484/tasks/42` directly.

When the services run with `docker-compose.yml`, set
`TASK_GRPC_ADDR=task-tracker:8487` in Auth's container environment. Port 8487
is also published for local inspection.

## Real-Life Scenarios

### An expense needs task details

An expense service may store a `task_id`, but it should not copy the complete
task record. When a user views an expense, Expense Tracker can call
`GetTask(task_id)` to show the current task title and status. A rename is
visible immediately because Task Tracker remains the source of truth.

### A notification checks whether a task is still active

Before sending a reminder, a notification service calls `GetTask`. If the task
is inactive or complete, it skips the message. The two-second deadline prevents
a slow Task Tracker from holding notification workers forever.

### Safe failures during an outage

If Task Tracker is unavailable, Auth returns `502` for the task-preview
endpoint instead of guessing from stale database copies. The client can retry
later; the request does not leak database credentials or internal SQL details.

## Updating The Contract

Treat protobuf fields as a public API. Add new fields with new numbers; never
reuse or renumber existing fields. For a breaking change, create `task/v2`.

After changing the `.proto` file, regenerate both service copies:

```bash
PATH=/tmp/protoc-29.3/bin:$PATH protoc -I ../../proto \
  --go_out=. --go_opt=module=github.com/mohit838/learn-go-with-project \
  --go-grpc_out=. --go-grpc_opt=module=github.com/mohit838/learn-go-with-project \
  ../../proto/task/v1/task.proto
```

Run that command from `services/auth` and `services/task-tracker`. In a normal
developer setup, replace `/tmp/protoc-29.3/bin` with the directory containing
your installed `protoc` and generator plugins.
