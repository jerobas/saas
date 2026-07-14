package database

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestNewDatabaseCreatesStrictV2Baseline(t *testing.T) {
	db, err := NewDatabase(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	pragmas := map[string]int{
		"foreign_keys":   1,
		"trusted_schema": 0,
		"application_id": applicationID,
		"user_version":   1,
	}
	for name, want := range pragmas {
		var got int
		if err := db.Conn.QueryRow("PRAGMA " + name).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("PRAGMA %s = %d, want %d", name, got, want)
		}
	}

	var journalMode string
	if err := db.Conn.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatal(err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", journalMode)
	}

	var migrations int
	if err := db.Conn.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrations); err != nil {
		t.Fatal(err)
	}
	if migrations != 1 {
		t.Fatalf("migration count = %d, want 1", migrations)
	}

	var domainTables, strictTables int
	if err := db.Conn.QueryRow(`
		SELECT COUNT(*), SUM(strict)
		FROM pragma_table_list
		WHERE schema = 'main' AND type = 'table' AND name NOT LIKE 'sqlite_%'
	`).Scan(&domainTables, &strictTables); err != nil {
		t.Fatal(err)
	}
	if domainTables != 17 || strictTables != domainTables {
		t.Fatalf("domain tables = %d and strict tables = %d, want 17 strict tables", domainTables, strictTables)
	}
}

func TestNewDatabaseHandlesConcurrentFirstOpen(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "concurrent.db")
	start := make(chan struct{})
	errors := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			db, err := NewDatabase(dbPath)
			if err == nil {
				err = db.Close()
			}
			errors <- err
		}()
	}
	close(start)
	wait.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent first open: %v", err)
		}
	}

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var migrations int
	if err := db.Conn.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrations); err != nil {
		t.Fatal(err)
	}
	if migrations != 1 {
		t.Fatalf("migration count after concurrent open = %d, want 1", migrations)
	}
}

func TestNewDatabaseCanReopenV2Database(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "app.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Conn.Exec(`
		INSERT INTO items (
			name, normalized_name, base_unit_code,
			is_purchasable, is_producible, is_sellable,
			created_at_ms, updated_at_ms
		) VALUES ('Flour', 'flour', 'g', 1, 0, 0, 1, 1)
	`); err != nil {
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

	var name string
	if err := db.Conn.QueryRow("SELECT name FROM items").Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Flour" {
		t.Fatalf("item name after reopen = %q, want Flour", name)
	}
}
