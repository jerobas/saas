package database

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDatabaseExportIncludesLatestCommittedWALData(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "active.db")
	db := newBackupTestDatabase(t, dbPath)

	var journalMode string
	if err := db.conn.QueryRow(`PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(journalMode, "wal") {
		t.Fatalf("journal_mode = %q, want WAL", journalMode)
	}

	const latestBusinessName = "Latest committed business name"
	if _, err := db.conn.Exec(`
		UPDATE app_settings
		SET business_name = ?, updated_at_ms = updated_at_ms + 1
		WHERE id = 1
	`, latestBusinessName); err != nil {
		t.Fatal(err)
	}
	walInfo, err := os.Stat(dbPath + "-wal")
	if err != nil {
		t.Fatalf("stat WAL after committed write: %v", err)
	}
	if walInfo.Size() == 0 {
		t.Fatal("WAL is empty after committed write")
	}

	backupPath := filepath.Join(t.TempDir(), "backup.db")
	if err := db.Export(backupPath); err != nil {
		t.Fatalf("export database: %v", err)
	}
	if err := validateDatabaseFile(backupPath); err != nil {
		t.Fatalf("validate exported database: %v", err)
	}

	reopened := newBackupTestDatabase(t, backupPath)
	var backedUpBusinessName string
	if err := reopened.conn.QueryRow(`
		SELECT business_name FROM app_settings WHERE id = 1
	`).Scan(&backedUpBusinessName); err != nil {
		t.Fatal(err)
	}
	if backedUpBusinessName != latestBusinessName {
		t.Fatalf("backed-up business name = %q, want %q", backedUpBusinessName, latestBusinessName)
	}
}

func TestDatabaseExportRefusesExistingDestination(t *testing.T) {
	db := newBackupTestDatabase(t, filepath.Join(t.TempDir(), "active.db"))
	destination := filepath.Join(t.TempDir(), "existing.db")
	original := []byte("do not overwrite this file")
	if err := os.WriteFile(destination, original, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := db.Export(destination); err == nil {
		t.Fatal("Export returned nil error for an existing destination")
	}
	content, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(original) {
		t.Fatalf("existing destination was modified: got %q, want %q", content, original)
	}
}

func TestDatabaseExportRemovesTemporaryFileAfterValidationFailure(t *testing.T) {
	directory := t.TempDir()
	db := newBackupTestDatabase(t, filepath.Join(directory, "active.db"))

	if _, err := db.conn.Exec(`DROP TRIGGER schema_migrations_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.conn.Exec(`
		UPDATE schema_migrations SET checksum = ? WHERE version = 1
	`, strings.Repeat("0", 64)); err != nil {
		t.Fatal(err)
	}

	destination := filepath.Join(directory, "invalid-backup.db")
	err := db.Export(destination)
	if !errors.Is(err, ErrMigrationChecksum) {
		t.Fatalf("Export error = %v, want ErrMigrationChecksum", err)
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("invalid backup destination exists or cannot be inspected: %v", err)
	}
	temporaryFiles, err := filepath.Glob(filepath.Join(directory, ".sweeters-backup-*.db"))
	if err != nil {
		t.Fatal(err)
	}
	if len(temporaryFiles) != 0 {
		t.Fatalf("temporary backup files were not removed: %v", temporaryFiles)
	}
}

func TestDatabaseImportRequiresRestartAndLeavesConnectionOpen(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "active.db")
	db := newBackupTestDatabase(t, dbPath)
	originalConnection := db.conn

	const beforeImport = "Live before import"
	if _, err := db.conn.Exec(`UPDATE app_settings SET business_name = ? WHERE id = 1`, beforeImport); err != nil {
		t.Fatal(err)
	}

	for _, source := range []string{"", filepath.Join(t.TempDir(), "missing.db"), dbPath} {
		err := db.Import(source)
		if !errors.Is(err, ErrRestoreRequiresRestart) {
			t.Fatalf("Import(%q) error = %v, want ErrRestoreRequiresRestart", source, err)
		}
		if db.conn != originalConnection {
			t.Fatalf("Import(%q) replaced the live connection", source)
		}
		var businessName string
		if err := db.conn.QueryRow(`SELECT business_name FROM app_settings WHERE id = 1`).Scan(&businessName); err != nil {
			t.Fatalf("query live database after Import(%q): %v", source, err)
		}
		if businessName != beforeImport {
			t.Fatalf("business name after Import(%q) = %q, want %q", source, businessName, beforeImport)
		}
	}

	const afterImport = "Live and writable after import"
	if _, err := db.conn.Exec(`UPDATE app_settings SET business_name = ? WHERE id = 1`, afterImport); err != nil {
		t.Fatalf("write live database after Import: %v", err)
	}
	var businessName string
	if err := db.conn.QueryRow(`SELECT business_name FROM app_settings WHERE id = 1`).Scan(&businessName); err != nil {
		t.Fatal(err)
	}
	if businessName != afterImport {
		t.Fatalf("business name after final write = %q, want %q", businessName, afterImport)
	}
}

func newBackupTestDatabase(t *testing.T, path string) *Database {
	t.Helper()
	db, err := NewDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})
	return db
}
