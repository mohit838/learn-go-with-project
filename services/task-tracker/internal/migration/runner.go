package migration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Runner struct {
	db  *sql.DB
	dir string
}

type fileMigration struct {
	version  string
	name     string
	upPath   string
	downPath string
	checksum string
}

type appliedMigration struct {
	version  string
	name     string
	batch    int
	checksum string
}

func NewRunner(db *sql.DB, dir string) *Runner {
	return &Runner{db: db, dir: dir}
}

func (r *Runner) Up(ctx context.Context) error {
	if err := r.ensureStore(ctx); err != nil {
		return err
	}

	files, err := r.loadFiles()
	if err != nil {
		return err
	}
	applied, err := r.applied(ctx)
	if err != nil {
		return err
	}
	if err := validateChecksums(files, applied); err != nil {
		return err
	}

	batch := latestBatch(applied) + 1
	ran := 0
	for _, file := range files {
		if _, ok := applied[file.version]; ok {
			continue
		}
		if err := r.runUp(ctx, file, batch); err != nil {
			return err
		}
		ran++
	}
	if ran == 0 {
		fmt.Println("Nothing to migrate.")
		return nil
	}
	fmt.Printf("Migrated %d migration(s).\n", ran)
	return nil
}

func (r *Runner) Rollback(ctx context.Context) error {
	if err := r.ensureStore(ctx); err != nil {
		return err
	}

	files, err := r.loadFiles()
	if err != nil {
		return err
	}
	applied, err := r.applied(ctx)
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		fmt.Println("Nothing to rollback.")
		return nil
	}

	batch := latestBatch(applied)
	toRollback := make([]fileMigration, 0)
	for _, file := range files {
		if appliedFile, ok := applied[file.version]; ok && appliedFile.batch == batch {
			toRollback = append(toRollback, file)
		}
	}
	sort.Slice(toRollback, func(i, j int) bool {
		return toRollback[i].version > toRollback[j].version
	})

	for _, file := range toRollback {
		if err := r.runDown(ctx, file); err != nil {
			return err
		}
	}
	fmt.Printf("Rolled back %d migration(s) from batch %d.\n", len(toRollback), batch)
	return nil
}

func (r *Runner) Status(ctx context.Context) error {
	if err := r.ensureStore(ctx); err != nil {
		return err
	}

	files, err := r.loadFiles()
	if err != nil {
		return err
	}
	applied, err := r.applied(ctx)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("No migration files found.")
		return nil
	}
	for _, file := range files {
		state := "pending"
		if appliedFile, ok := applied[file.version]; ok {
			state = fmt.Sprintf("ran batch=%d", appliedFile.batch)
		}
		fmt.Printf("%s %-8s %s\n", file.version, state, file.name)
	}
	return nil
}

func (r *Runner) Make(name string) error {
	cleanName := sanitizeName(name)
	if cleanName == "" {
		return fmt.Errorf("migration name cannot be empty")
	}
	if err := os.MkdirAll(r.dir, 0755); err != nil {
		return err
	}

	version := time.Now().UTC().Format("20060102150405")
	upPath := filepath.Join(r.dir, version+"_"+cleanName+".up.sql")
	downPath := filepath.Join(r.dir, version+"_"+cleanName+".down.sql")
	if err := os.WriteFile(upPath, []byte("-- Write migration SQL here.\n"), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(downPath, []byte("-- Write rollback SQL here.\n"), 0644); err != nil {
		return err
	}

	fmt.Println("Created", upPath)
	fmt.Println("Created", downPath)
	return nil
}

func (r *Runner) ensureStore(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	id BIGSERIAL PRIMARY KEY,
	version TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	batch INTEGER NOT NULL,
	checksum TEXT NOT NULL,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	execution_ms BIGINT NOT NULL DEFAULT 0
)`)
	return err
}

func (r *Runner) runUp(ctx context.Context, file fileMigration, batch int) error {
	sqlBody, err := os.ReadFile(file.upPath)
	if err != nil {
		return err
	}

	start := time.Now()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, string(sqlBody)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("run %s: %w", file.upPath, err)
	}
	executionMS := time.Since(start).Milliseconds()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO schema_migrations (version, name, batch, checksum, execution_ms)
VALUES ($1, $2, $3, $4, $5)`, file.version, file.name, batch, file.checksum, executionMS); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Println("Migrated", filepath.Base(file.upPath))
	return nil
}

func (r *Runner) runDown(ctx context.Context, file fileMigration) error {
	sqlBody, err := os.ReadFile(file.downPath)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, string(sqlBody)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("rollback %s: %w", file.downPath, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = $1`, file.version); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Println("Rolled back", filepath.Base(file.downPath))
	return nil
}

func (r *Runner) applied(ctx context.Context) (map[string]appliedMigration, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT version, name, batch, checksum FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]appliedMigration)
	for rows.Next() {
		var item appliedMigration
		if err := rows.Scan(&item.version, &item.name, &item.batch, &item.checksum); err != nil {
			return nil, err
		}
		applied[item.version] = item
	}
	return applied, rows.Err()
}

func (r *Runner) loadFiles() ([]fileMigration, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	byVersion := make(map[string]*fileMigration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		direction := ""
		switch {
		case strings.HasSuffix(filename, ".up.sql"):
			direction = "up"
		case strings.HasSuffix(filename, ".down.sql"):
			direction = "down"
		default:
			continue
		}

		parts := strings.SplitN(strings.TrimSuffix(strings.TrimSuffix(filename, ".up.sql"), ".down.sql"), "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid migration filename: %s", filename)
		}
		version, name := parts[0], parts[1]
		item := byVersion[version]
		if item == nil {
			item = &fileMigration{version: version, name: name}
			byVersion[version] = item
		}
		path := filepath.Join(r.dir, filename)
		if direction == "up" {
			item.upPath = path
			checksum, err := checksumFile(path)
			if err != nil {
				return nil, err
			}
			item.checksum = checksum
		} else {
			item.downPath = path
		}
	}

	files := make([]fileMigration, 0, len(byVersion))
	for _, item := range byVersion {
		if item.upPath == "" || item.downPath == "" {
			return nil, fmt.Errorf("migration %s_%s must have up and down files", item.version, item.name)
		}
		files = append(files, *item)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].version < files[j].version
	})
	return files, nil
}

func validateChecksums(files []fileMigration, applied map[string]appliedMigration) error {
	for _, file := range files {
		appliedFile, ok := applied[file.version]
		if ok && appliedFile.checksum != file.checksum {
			return fmt.Errorf("migration checksum changed after apply: %s_%s", file.version, file.name)
		}
	}
	return nil
}

func latestBatch(applied map[string]appliedMigration) int {
	latest := 0
	for _, item := range applied {
		if item.batch > latest {
			latest = item.batch
		}
	}
	return latest
}

func checksumFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func sanitizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "_")
	return strings.Trim(name, "_")
}
