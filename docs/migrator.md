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

## CLI Commands

Use these commands in normal development. Run them from the project root.

Create a migration:

```sh
go -C services/auth run ./cmd/migrate make create_users_table
```

The library also accepts `create` and `new` as aliases:

```sh
go -C services/auth run ./cmd/migrate create create_users_table
go -C services/auth run ./cmd/migrate new create_users_table
```

Run migrations:

```sh
go -C services/auth run ./cmd/migrate up
```

Show status:

```sh
go -C services/auth run ./cmd/migrate status
```

Rollback latest batch:

```sh
go -C services/auth run ./cmd/migrate rollback
```

Use the same commands with:

```text
services/task-tracker
services/expense-tracker
```

That means each service keeps its own DB config and migration folder, while the
developer can still run from the root.

If your Go version does not support `go -C`, run the command inside the service
folder:

```sh
cd services/auth
go run ./cmd/migrate up
```

## Optional Make Commands

Make is not required by the library.

This project keeps Make shortcuts only for convenience:

```sh
make migrate-make service=auth name=create_users_table
make migrate-up service=auth
make migrate-status service=auth
make migrate-rollback service=auth
```

Those shortcuts call the selected service's migration entrypoint:

```sh
cd services/auth && go run ./cmd/migrate up
```

## Adding The Library Somewhere Else

For any new app or service, add:

```text
cmd/migrate/main.go
migrations/
```

In `cmd/migrate/main.go`, load env, connect to DB, then use:

```go
args := os.Args[1:]
runnerConfig := migration.Config{
	Dir:         "migrations",
	ServiceName: "my-service",
}
if !migration.NeedsDatabase(args) {
	runner := migration.NewRunner(nil, runnerConfig)
	if err := runner.Run(context.Background(), args); err != nil {
		log.Fatal(err)
	}
	return
}

// Load env and connect to DB here for up/status/rollback.
runner := migration.NewRunner(db, runnerConfig)

if err := runner.Run(context.Background(), args); err != nil {
	log.Fatal(err)
}
```

This keeps file generation commands from requiring a live database connection.

Then run inside that app:

```sh
go run ./cmd/migrate make create_users_table
go run ./cmd/migrate create create_users_table
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/migrate rollback
```

In a monorepo, use Go's `-C` flag for root commands:

```sh
go -C services/my-service run ./cmd/migrate up
```

Make shortcuts can be added later if the team wants them, but they are not part
of the migrator requirement.

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

The library also prevents generated files from overwriting existing migration
files, validates 14-digit timestamp versions, and lets callers inject a writer
for command output.

## Why Not Public Yet

Keep this private until:

- it works well across all three services
- we add tests around checksum and rollback behavior
- we decide whether to support only PostgreSQL or more databases
- we decide whether to support embedded migrations
- the API feels stable enough for other people
