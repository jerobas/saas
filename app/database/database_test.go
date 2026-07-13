package database

import (
	"path/filepath"
	"testing"
)

func TestNewDatabaseAppliesSchemasAndEnablesForeignKeys(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "app.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var foreignKeys int
	if err := db.Conn.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys = %d, want 1", foreignKeys)
	}

	var migrations int
	if err := db.Conn.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrations); err != nil {
		t.Fatal(err)
	}
	if migrations != 7 {
		t.Fatalf("migration count = %d, want 7", migrations)
	}
}

func TestNewDatabaseCanReopenMigratedDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "app.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
}
