package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"
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
		"busy_timeout":   5000,
		"synchronous":    1,
		"application_id": applicationID,
		"user_version":   2,
	}
	for name, want := range pragmas {
		var got int
		if err := db.conn.QueryRow("PRAGMA " + name).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("PRAGMA %s = %d, want %d", name, got, want)
		}
	}

	var journalMode string
	if err := db.conn.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatal(err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", journalMode)
	}

	var migrations int
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrations); err != nil {
		t.Fatal(err)
	}
	if migrations != 2 {
		t.Fatalf("migration count = %d, want 2", migrations)
	}

	var domainTables, strictTables int
	if err := db.conn.QueryRow(`
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

func TestConnectionPragmasSurviveDriverReplacement(t *testing.T) {
	options := DefaultOpenOptions()
	options.BusyTimeout = 37 * time.Millisecond
	db, err := NewDatabaseWithOptions(filepath.Join(t.TempDir(), "replacement.db"), options)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	want := map[string]int{
		"foreign_keys":   1,
		"trusted_schema": 0,
		"busy_timeout":   37,
		"synchronous":    1,
	}
	assertIntegerPragmas(t, db.conn, want)

	ctx := context.Background()
	connection, err := db.conn.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"CREATE TEMP TABLE replacement_marker (value INTEGER)",
		"PRAGMA foreign_keys = OFF",
		"PRAGMA trusted_schema = ON",
		"PRAGMA busy_timeout = 1",
		"PRAGMA synchronous = FULL",
	} {
		if _, err := connection.ExecContext(ctx, statement); err != nil {
			_ = connection.Close()
			t.Fatalf("poison original connection with %q: %v", statement, err)
		}
	}
	assertIntegerPragmas(t, connection, map[string]int{
		"foreign_keys": 0, "trusted_schema": 1, "busy_timeout": 1, "synchronous": 2,
	})

	rawErr := connection.Raw(func(any) error { return driver.ErrBadConn })
	if rawErr != nil && !errors.Is(rawErr, driver.ErrBadConn) {
		_ = connection.Close()
		t.Fatalf("invalidate physical connection: %v", rawErr)
	}
	if err := connection.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("close invalidated connection: %v", err)
	}

	var markerCount int
	if err := db.conn.QueryRowContext(ctx, `
		SELECT count(*) FROM sqlite_temp_master WHERE name = 'replacement_marker'
	`).Scan(&markerCount); err != nil {
		t.Fatal(err)
	}
	if markerCount != 0 {
		t.Fatal("temporary marker survived; database/sql did not replace the physical connection")
	}
	assertIntegerPragmas(t, db.conn, want)

	var journalMode string
	if err := db.conn.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatal(err)
	}
	if journalMode != "wal" {
		t.Fatalf("replacement journal_mode = %q, want wal", journalMode)
	}
	if _, err := db.conn.ExecContext(ctx, `
		INSERT INTO counterparty_roles (counterparty_id, role, created_at_ms)
		VALUES (999, 'SUPPLIER', 1)
	`); err == nil {
		t.Fatal("replacement connection did not enforce foreign keys")
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
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrations); err != nil {
		t.Fatal(err)
	}
	if migrations != 2 {
		t.Fatalf("migration count after concurrent open = %d, want 2", migrations)
	}
}

func TestNewDatabaseCanReopenV2Database(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "app.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.conn.Exec(`
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
	if err := db.conn.QueryRow("SELECT name FROM items").Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Flour" {
		t.Fatalf("item name after reopen = %q, want Flour", name)
	}
}

type integerPragmaQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func assertIntegerPragmas(t *testing.T, queryer integerPragmaQueryer, want map[string]int) {
	t.Helper()
	for name, expected := range want {
		var actual int
		if err := queryer.QueryRowContext(context.Background(), "PRAGMA "+name).Scan(&actual); err != nil {
			t.Fatalf("read PRAGMA %s: %v", name, err)
		}
		if actual != expected {
			t.Errorf("PRAGMA %s = %d, want %d", name, actual, expected)
		}
	}
}
