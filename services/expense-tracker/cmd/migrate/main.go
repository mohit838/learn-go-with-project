package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/mohit838/learn-go-with-project/internal/config"
	"github.com/mohit838/learn-go-with-project/internal/database"
	"github.com/mohit838/learn-go-with-project/libs/migrator/migration"
)

func main() {
	cfg, err := config.LoadConfig("./.env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.ConnectDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer closeDB(db)

	runner := migration.NewRunner(db, migration.Config{
		Dir:         "migrations",
		ServiceName: "expense-tracker",
	})

	if err := runner.Run(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func closeDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("close database: %v", err)
	}
}
