package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/mohit838/learn-go-with-project/internal/config"
	"github.com/mohit838/learn-go-with-project/internal/database"
	"github.com/mohit838/learn-go-with-project/internal/migration"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.LoadConfig("./.env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.ConnectDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer closeDB(db)

	runner := migration.NewRunner(db, "migrations")
	ctx := context.Background()

	switch os.Args[1] {
	case "up":
		err = runner.Up(ctx)
	case "down", "rollback":
		err = runner.Rollback(ctx)
	case "status":
		err = runner.Status(ctx)
	case "make":
		if len(os.Args) < 3 {
			log.Fatal("migration name is required")
		}
		err = runner.Make(os.Args[2])
	default:
		printUsage()
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Println("usage: go run ./cmd/migrate [up|down|rollback|status|make] [name]")
}

func closeDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("close database: %v", err)
	}
}
