# Notification Service

Notification Service is a small learning service for Go concurrency. It accepts
notification jobs over HTTP or gRPC, sends them into a buffered channel, and lets
worker goroutines process them in the background.

## Run Locally

```sh
cd services/notification-service
go run ./cmd/api
```

Ports:

```text
HTTP: 8487
gRPC: 8587
```

## HTTP API

HTTP is for frontend and gateway traffic:

```text
POST /notifications
GET  /notifications/{id}
GET  /notifications/stats
GET  /health
GET  /ready
```

Protected routes require trusted gateway headers:

```text
X-User-ID
X-Tenant-ID
X-Tenant-Slug
X-User-Role
```

Kong adds those headers after validating the access token.

## gRPC API

gRPC is for backend service-to-service traffic:

```text
notification.v1.NotificationService/SubmitNotification
notification.v1.NotificationService/GetNotification
notification.v1.NotificationService/Stats
```

The proto contract lives in:

```text
proto/notification/v1/notification.proto
```

## Concurrency Flow

```text
HTTP/gRPC request
  -> application.Service.Submit
  -> jobs channel
  -> worker goroutine
  -> results channel
  -> collector goroutine
  -> in-memory status map
```

The service is intentionally in-memory for now. That keeps the lesson focused on
goroutines, channels, cancellation, and graceful shutdown.
