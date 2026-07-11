# Go Concurrency With Notification Service

This service exists for learning core Go concurrency in a real microservice
shape. It is intentionally simple: no database, no Redis, no MinIO. The important
part is how a request becomes background work.

## Where To Read

```text
services/notification-service/cmd/api/main.go
services/notification-service/internal/notification/application/service.go
services/notification-service/internal/notification/transport/http_handler.go
services/notification-service/internal/notification/transport/grpc_server.go
proto/notification/v1/notification.proto
```

## Ports

```text
HTTP: 8487
gRPC: 8587
Kong route: http://localhost:8000/notifications
```

## Run It

Start the service:

```sh
cd services/notification-service
go run ./cmd/api
```

Start a gateway in another terminal:

```sh
make dev-up
```

The frontend has a `Notifications` page. It calls the gateway, not the direct
service port.

## The Concurrency Model

The application service owns two channels:

```go
jobs    chan job
results chan result
```

`jobs` is the input channel. HTTP and gRPC handlers call `Submit`, and `Submit`
tries to send a job into this channel.

`results` is the output channel. Worker goroutines send completed work into this
channel.

One collector goroutine reads from `results` and updates the in-memory map. This
keeps status writes in one place and makes the code easier to reason about.

## Request Flow

```text
1. Client calls POST /notifications through gateway.
2. Gateway validates JWT and forwards X-User-* headers.
3. HTTP handler decodes JSON and calls application.Service.Submit.
4. Submit creates a notification with status=queued.
5. Submit sends the job into jobs channel.
6. A worker goroutine receives from jobs.
7. Worker marks status=sending and simulates delivery.
8. Worker sends result into results channel.
9. Collector receives result and marks delivered or failed.
```

The same application service is used by gRPC, so HTTP and gRPC share the worker
pool.

## Why Buffered Channels

The jobs channel is buffered:

```go
jobs: make(chan job, queueSize)
```

That means a short burst of requests can wait in memory while workers are busy.
If the queue stays full, `Submit` returns `ErrQueueFull` instead of making the
client wait forever.

## Why `select`

`Submit` uses `select` to handle multiple outcomes:

```text
job accepted
request cancelled
service shutting down
queue timeout
```

This is a professional Go pattern: never block forever when the caller, process,
or queue state says the work should stop.

## Why WaitGroups

The service uses separate wait groups:

```text
workerWG    waits for worker goroutines
collectorWG waits for the collector goroutine
```

Shutdown order matters:

```text
close done
close jobs
wait for workers
close results
wait for collector
```

Only the owner of a channel should close it. In this service:

```text
Service closes jobs
Service closes results after workers finish
Workers never close channels
Handlers never close channels
```

## gRPC In This Service

The gRPC contract is internal:

```text
notification.v1.NotificationService
```

Methods:

```text
SubmitNotification
GetNotification
Stats
```

Use gRPC when another backend service needs to talk to notification-service
directly. Do not call gRPC from the browser. Browser/frontend traffic should go
through HTTP and the gateway.

The current code manually registers gRPC methods while keeping the proto file in
the repo. Later, generated protobuf code can replace the small
`internal/notificationrpc` compatibility structs.

## Test

```sh
cd services/notification-service
go test ./...
```

The main test submits a notification, then polls until a worker delivers it.
That proves the request moved through the jobs channel, worker goroutine, results
channel, and collector.
