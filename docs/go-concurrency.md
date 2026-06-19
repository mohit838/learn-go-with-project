# Go Concurrency: A Friendly Microservice Study Guide

Concurrency means making progress on more than one piece of work at a time.
In Go, that usually means goroutines, channels, and `context`. Think of a
restaurant: a waiter takes an order, the kitchen cooks it, and another person
can take the next order. They are cooperating; they are not all trying to hold
the same frying pan.

The rule that keeps this pleasant: start a goroutine only when it has a clear
owner, a stopping condition, and a reason to run separately.

## 1. Goroutines: Do Independent Work

A goroutine runs a function concurrently. Use one when waiting on that work
does not need to block the current request.

```go
go func() {
    auditStore.Record(ctx, service, requestID, method, path, status, duration)
}()
```

This project already uses that shape for Mongo audit logging. The response is
sent first; writing a non-critical audit event happens afterward.

### Scenario: send two independent lookups at once

If an API needs a user profile and a task summary, start both calls, wait for
both results, and cancel the other call if one fails:

```go
group, ctx := errgroup.WithContext(request.Context())

group.Go(func() error { profile, err = users.Get(ctx, userID); return err })
group.Go(func() error { task, err = tasks.Get(ctx, taskID); return err })

if err := group.Wait(); err != nil {
    return err
}
```

Use this for independent I/O. Do not use it merely to make a short piece of
ordinary code look exciting. Tiny goroutines are still tiny pets: someone must
feed, clean up, and eventually stop them.

## 2. Channels: Pass Work, Not Shared Memory

A channel is a typed pipe. One goroutine sends values and another receives
them. Prefer a channel when it makes ownership and sequencing clearer.

```go
jobs := make(chan string)
results := make(chan error)

go func() {
    for email := range jobs {
        results <- sendEmail(email)
    }
}()

jobs <- "ada@example.com"
close(jobs) // no more jobs will be sent
if err := <-results; err != nil { /* handle it */ }
```

### Scenario: worker pool for slow background jobs

For image processing, CSV imports, or notification delivery, use a small fixed
number of workers. This protects the database and external APIs from a sudden
flood of 10,000 goroutines.

```go
jobs := make(chan Job)
var workers sync.WaitGroup

for i := 0; i < 4; i++ {
    workers.Add(1)
    go func() {
        defer workers.Done()
        for job := range jobs {
            process(job)
        }
    }()
}

// producer sends jobs, then closes jobs
close(jobs)
workers.Wait()
```

Four workers is not magical. Pick a small starting number, measure, and tune
it based on CPU, database connections, and provider limits.

## 3. Buffered vs Unbuffered Channels

- An unbuffered channel (`make(chan T)`) is a handoff: sender and receiver meet.
- A buffered channel (`make(chan T, 10)`) is a small waiting room.

A buffer smooths brief bursts; it is not a cure for a worker that is slower
than its producer forever. A full channel is useful backpressure: it tells the
producer to slow down instead of filling memory.

## 4. Context: The Stop Button

Pass `context.Context` into database, Redis, Mongo, HTTP, and gRPC work. It
carries cancellation and deadlines down the call chain.

```go
ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
defer cancel()

task, err := taskClient.Get(ctx, id)
```

### Scenario: client went away

If a browser closes a page while a request is waiting on Task Tracker, the
request context is cancelled. The gRPC call can stop instead of doing work for
a response nobody will receive.

Never replace a request context with `context.Background()` for request work.
`Background` is suitable only for deliberately detached work, such as the
bounded audit write above.

## 5. Select: Listen to More Than One Thing

`select` waits until one channel operation can proceed. It is especially useful
for a worker that must either receive work or stop cleanly.

```go
for {
    select {
    case job, ok := <-jobs:
        if !ok { return } // producer closed the channel
        process(job)
    case <-ctx.Done():
        return // shutdown or deadline
    }
}
```

## 6. Graceful Shutdown: Let Work Finish Politely

When Docker, Kubernetes, or Ctrl+C sends `SIGTERM`/`SIGINT`, a service should:

1. Stop accepting new HTTP and gRPC requests.
2. Give in-flight requests a short deadline to finish.
3. Stop background workers through context cancellation or closed job channels.
4. Close gRPC, Redis, MongoDB, and PostgreSQL connections.

Here is the basic HTTP shape:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

server := &http.Server{Addr: ":8484", Handler: handler}
go func() {
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        logger.Error("http server stopped", "error", err)
    }
}()

<-ctx.Done() // wait for Ctrl+C or SIGTERM
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
_ = server.Shutdown(shutdownCtx)
```

For gRPC, call `GracefulStop` after the HTTP server has stopped accepting new
work. If it cannot finish before your shutdown deadline, call `Stop` as the
last-resort escape hatch.

## 7. Common Bugs to Hunt Like a Detective

| Smell | What it means | Fix |
| --- | --- | --- |
| A goroutine has no context or exit path | It may leak forever | Give it a context, close its channel, or both. |
| Sending to a closed channel | The program panics | The sender that owns production should decide when to close. |
| Two places close the same channel | One eventually panics | One channel, one closer. |
| Reading shared maps from many goroutines | Data races or crashes | Use a mutex, a dedicated owner goroutine, or `sync.Map` when appropriate. |
| `time.Sleep` to coordinate work | Timing-dependent tests and bugs | Use channels, `WaitGroup`, or contexts. |
| Spawning one goroutine per unlimited request | Resource exhaustion | Use limits, a worker pool, or backpressure. |

Run race detection while learning:

```bash
go test -race ./...
```

The race detector is a wonderful grumpy librarian: it notices when two people
touch the same book at once and asks both of them to please stop.

## Suggested Practice Missions

1. Add a cancellable worker pool that processes fake email jobs.
2. Add a channel-backed audit queue with a bounded size and a dropped-event log
   when it is full.
3. Change one API server from `http.ListenAndServe` to `http.Server` and add
   signal-based graceful shutdown.
4. Write a test that cancels a context and proves a worker exits.
5. Run the test with `-race`, then intentionally introduce a shared-counter
   race so you can see the detector report it. Fix it with a mutex.

Learn these in order. Concurrency is less about cleverness and more about
making the next person—often future you—able to explain when work starts, who
owns it, and how it stops.
