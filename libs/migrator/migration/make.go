package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

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

	fmt.Println(r.logPrefix()+"Created", upPath)
	fmt.Println(r.logPrefix()+"Created", downPath)
	return nil
}

func sanitizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "_")
	return strings.Trim(name, "_")
}
