// Package database
package database

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed migrations/*.sql
var migrations embed.FS

func (h *Handler) Migrate() error {
	files, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, file := range files {
		stmt, err := migrations.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(stmt)); err != nil {
			return fmt.Errorf("migration %s: %w", file, err)
		}
	}

	return tx.Commit()
}
