package migration

import (
	"context"
	"fmt"
	"strings"
)

func (r *Runner) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%s", Usage())
	}

	switch args[0] {
	case "up":
		return r.Up(ctx)
	case "down", "rollback":
		return r.Rollback(ctx)
	case "status":
		return r.Status(ctx)
	case "make":
		if len(args) < 2 {
			return fmt.Errorf("migration name is required\n\n%s", Usage())
		}
		return r.Make(args[1])
	case "help", "-h", "--help":
		fmt.Println(Usage())
		return nil
	default:
		return fmt.Errorf("unknown migration command: %s\n\n%s", args[0], Usage())
	}
}

func Usage() string {
	return strings.TrimSpace(`usage: go run ./cmd/migrate [command] [name]

commands:
  up                 run all pending migrations
  status             show migration status
  rollback, down     rollback the latest migration batch
  make <name>        create paired .up.sql and .down.sql files
  help               show this help`)
}
