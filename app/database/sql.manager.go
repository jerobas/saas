package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrRestoreRequiresRestart = errors.New("database restore is disabled until restart-safe activation is implemented")

func (d *Database) Export(destPath string) error {
	if destPath == "" {
		return errors.New("backup destination is empty")
	}
	absPath, err := filepath.Abs(destPath)
	if err != nil {
		return fmt.Errorf("resolve backup destination: %w", err)
	}
	if _, err := os.Stat(absPath); err == nil {
		return fmt.Errorf("backup destination already exists: %s", absPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect backup destination: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(absPath), ".sweeters-backup-*.db")
	if err != nil {
		return fmt.Errorf("reserve backup path: %w", err)
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("close backup placeholder: %w", err)
	}
	if err := os.Remove(temporaryPath); err != nil {
		return fmt.Errorf("prepare backup path: %w", err)
	}
	defer os.Remove(temporaryPath)

	if _, err := d.ExecContext(context.Background(), "VACUUM INTO ?", temporaryPath); err != nil {
		return fmt.Errorf("export database: %w", err)
	}
	if err := validateDatabaseFile(temporaryPath); err != nil {
		return fmt.Errorf("validate exported database: %w", err)
	}
	if err := os.Rename(temporaryPath, absPath); err != nil {
		return fmt.Errorf("publish backup: %w", err)
	}
	return nil
}

func (d *Database) Import(srcPath string) error {
	_ = srcPath
	return ErrRestoreRequiresRestart
}

func validateDatabaseFile(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	if err := db.Ping(); err != nil {
		return err
	}
	migrations, err := loadMigrations(schemaFS, "schemas")
	if err != nil {
		return err
	}
	if err := validateMigrationHistory(db, migrations, true); err != nil {
		return err
	}
	return validateDatabase(db)
}
