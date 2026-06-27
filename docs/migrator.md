# Migrator Library Guideline

The shared migration code lives in:

```text
libs/migrator
```

This is a private monorepo library for now. We use it in the services first,
then decide later if it is good enough to publish as open source.

## Responsibility Split

The library handles:

- creating the migration tracking table
- reading `.up.sql` and `.down.sql` files
- calculating checksums
- running pending migrations
- rolling back the latest batch
- printing status
- creating paired migration files

Each service handles:

- loading `.env`
- creating the database connection
- choosing the migration folder path
- passing a service name for logs

This keeps the library portable for microservice and non-microservice projects.

## Current Database

The current implementation targets PostgreSQL through Go's `database/sql`.

Each service passes a ready `*sql.DB` to the migrator:

```go
runner := migration.NewRunner(db, migration.Config{
	Dir:         "migrations",
	ServiceName: "auth",
})
```

The migrator does not know or care where `db` came from.

## Service Usage

Auth:

```text
services/auth/cmd/migrate/main.go
services/auth/migrations
```

Task tracker:

```text
services/task-tracker/cmd/migrate/main.go
services/task-tracker/migrations
```

Expense tracker:

```text
services/expense-tracker/cmd/migrate/main.go
services/expense-tracker/migrations
```

Each service imports:

```go
github.com/mohit838/learn-go-with-project/libs/migrator/migration
```

The service wrapper should stay small:

```go
runner := migration.NewRunner(db, migration.Config{
	Dir:         "migrations",
	ServiceName: "auth",
})

if err := runner.Run(context.Background(), os.Args[1:]); err != nil {
	log.Fatal(err)
}
```

Each service `go.mod` keeps the library local:

```go
require github.com/mohit838/learn-go-with-project/libs/migrator v0.0.0

replace github.com/mohit838/learn-go-with-project/libs/migrator => ../../libs/migrator
```

## Makefile Commands

Use these commands in normal development. This is the preferred workflow, so
developers do not need to think about the library internals.

Run them from the project root.

Create a migration:

```sh
make migrate-make service=auth name=create_users_table
```

Run migrations:

```sh
make migrate-up service=auth
```

Show status:

```sh
make migrate-status service=auth
```

Rollback latest batch:

```sh
make migrate-rollback service=auth
```

Use the same commands with:

```text
service=task-tracker
service=expense-tracker
```

The root Makefile calls the selected service's migration entrypoint:

```sh
cd services/auth && go run ./cmd/migrate up
```

That means each service keeps its own DB config and migration folder, while the
developer uses one command style from the root.

## Without Make

If `make` is not installed, use the Go CLI directly from the project root:

```sh
go -C services/auth run ./cmd/migrate make create_users_table
go -C services/auth run ./cmd/migrate up
go -C services/auth run ./cmd/migrate status
go -C services/auth run ./cmd/migrate rollback
```

Change the service folder as needed:

```sh
go -C services/task-tracker run ./cmd/migrate up
go -C services/expense-tracker run ./cmd/migrate up
```

If your Go version does not support `go -C`, run the command inside the service
folder:

```sh
cd services/auth
go run ./cmd/migrate up
```

## Adding The Library Somewhere Else

For any new app or service, add:

```text
cmd/migrate/main.go
migrations/
```

In `cmd/migrate/main.go`, load env, connect to DB, then use:

```go
runner := migration.NewRunner(db, migration.Config{
	Dir:         "migrations",
	ServiceName: "my-service",
})

if err := runner.Run(context.Background(), os.Args[1:]); err != nil {
	log.Fatal(err)
}
```

Then run inside that app:

```sh
go run ./cmd/migrate make create_users_table
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/migrate rollback
```

In a monorepo, add root Makefile shortcuts so users can run the same commands
from the root.

Without Make, use:

```sh
go -C services/my-service run ./cmd/migrate up
```

## File Rules

Migration files must be paired:

```text
YYYYMMDDHHMMSS_name.up.sql
YYYYMMDDHHMMSS_name.down.sql
```

Example:

```text
20260627120000_create_tasks_table.up.sql
20260627120000_create_tasks_table.down.sql
```

Do not edit an already applied `.up.sql` file. The checksum check will fail.
Create a new migration instead.

## Why Not Public Yet

Keep this private until:

- it works well across all three services
- we add tests around checksum and rollback behavior
- we decide whether to support only PostgreSQL or more databases
- we decide whether to support embedded migrations
- the API feels stable enough for other people
