package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mohit838/learn-go-with-project/internal/config"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run ./cmd/migrate [up|down|version]")
	}

	cfg, err := config.LoadConfig("./.env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if cfg.DBURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	m, err := migrate.New("file://migrations", cfg.DBURL)
	if err != nil {
		log.Fatalf("create migration client: %v", err)
	}
	defer m.Close()

	switch os.Args[1] {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	case "version":
		version, dirty, versionErr := m.Version()
		if errors.Is(versionErr, migrate.ErrNilVersion) {
			fmt.Println("version: no migrations applied")
			return
		}
		if versionErr != nil {
			log.Fatalf("read migration version: %v", versionErr)
		}
		fmt.Printf("version: %d, dirty: %t\n", version, dirty)
		return
	default:
		log.Fatal("usage: go run ./cmd/migrate [up|down|version]")
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("run migration: %v", err)
	}

	log.Println("migration completed")
}
