# Migrator

Reusable SQL migration library for this project.

For now this library is private inside the monorepo. We use it in real services
first, then later decide if it is stable enough to publish.

## Fast Usage In This Project

You usually do not need to touch Go code.

Use the Go CLI directly from the project root.

Create a migration:

```sh
go -C services/auth run ./cmd/migrate make create_users_table
```

Run pending migrations:

```sh
go -C services/auth run ./cmd/migrate up
```

Check status:

```sh
go -C services/auth run ./cmd/migrate status
```

Rollback the latest batch:

```sh
go -C services/auth run ./cmd/migrate rollback
```

Use the same pattern with:

```text
services/auth
services/task-tracker
services/expense-tracker
```

That is the normal developer workflow.

Task tracker example:

```sh
go -C services/task-tracker run ./cmd/migrate make create_tasks_table
go -C services/task-tracker run ./cmd/migrate up
go -C services/task-tracker run ./cmd/migrate status
go -C services/task-tracker run ./cmd/migrate rollback
```

Expense tracker example:

```sh
go -C services/expense-tracker run ./cmd/migrate make create_expenses_table
go -C services/expense-tracker run ./cmd/migrate up
go -C services/expense-tracker run ./cmd/migrate status
go -C services/expense-tracker run ./cmd/migrate rollback
```

If your Go version does not support `go -C`, use `cd`:

```sh
cd services/auth
go run ./cmd/migrate up
```

So the service controls its own `.env`, database connection, and `migrations`
folder.

## Add To Another Service Or Project

After installing or adding the package, create:

```text
cmd/migrate/main.go
migrations/
```

The migration main file should connect to the database, create the runner, then
pass command-line args to the library:

```go
package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mohit838/learn-go-with-project/libs/migrator/migration"
)

func main() {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	runner := migration.NewRunner(db, migration.Config{
		Dir:         "migrations",
		ServiceName: "my-service",
	})

	if err := runner.Run(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
```

Then run from that app root:

```sh
go run ./cmd/migrate make create_users_table
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/migrate rollback
```

For a root-level command, use Go's `-C` flag:

```sh
go -C services/my-service run ./cmd/migrate up
```

## Optional Make Shortcuts

Make is not required by the library.

This project keeps root Makefile shortcuts only for convenience:

```sh
make migrate-make service=auth name=create_users_table
make migrate-up service=auth
make migrate-status service=auth
make migrate-rollback service=auth
```

Those shortcuts only call the same Go CLI commands:

```makefile
migrate-up:
	cd services/$(service) && go run ./cmd/migrate up

migrate-status:
	cd services/$(service) && go run ./cmd/migrate status

migrate-rollback:
	cd services/$(service) && go run ./cmd/migrate rollback

migrate-make:
	cd services/$(service) && go run ./cmd/migrate make $(name)
```

## What The Library Does

The library handles the boring migration parts:

- creates the migration tracking table
- creates paired `.up.sql` and `.down.sql` files
- reads migration files from a folder
- runs pending migrations in order
- rolls back the latest batch
- stores checksums so old migrations cannot silently change
- prints migration status

The service still handles:

- loading `.env`
- connecting to the database
- choosing the migration folder
- passing the service name

This keeps the library portable for microservice and non-microservice projects.

## Migration Files

Each migration has two files:

```text
20260627120000_create_users_table.up.sql
20260627120000_create_users_table.down.sql
```

The `.up.sql` file applies the change.
The `.down.sql` file rolls it back.

Do not edit an already-applied `.up.sql` file. The checksum check will fail.
Create a new migration instead.

## Folders

Each service owns its own migrations:

```text
services/auth/migrations
services/task-tracker/migrations
services/expense-tracker/migrations
```

For a normal single app, use the same simple shape:

```text
my-app/migrations
```

## Config

```go
type Config struct {
	Dir         string
	TableName   string
	ServiceName string
}
```

- `Dir`: migration folder. Default is `migrations`.
- `TableName`: tracking table. Default is `schema_migrations`.
- `ServiceName`: optional label for logs and tracking rows.

## Database Support

Current implementation is tested with PostgreSQL through Go's `database/sql`.

The default tracking table is:

```text
schema_migrations
```

It stores:

- `version`
- `name`
- `service_name`
- `batch`
- `checksum`
- `applied_at`
- `execution_ms`

## Batch And Rollback

Every `up` run creates a new batch number.

`rollback` rolls back the latest batch in reverse order, similar to Laravel
migrations.

## Production Usage

Do not run migrations forever inside the API server.

In production, run migrations only as a deploy step when schema changes:

```sh
make migrate-up service=auth
```

If there is no schema change, do not run migration commands.

## Publishing Later

Before publishing this library publicly, decide:

- PostgreSQL-only or multi-database support
- public module path and repository name
- tests using a temporary PostgreSQL container
- custom logger support
- embedded migration support with `fs.FS`
