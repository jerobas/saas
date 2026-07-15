package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Database struct {
	conn *sql.DB
}

// OpenOptions configures connection behavior that needs to differ in bounded
// integration tests. Production callers should start with DefaultOpenOptions.
type OpenOptions struct {
	BusyTimeout time.Duration
}

func DefaultOpenOptions() OpenOptions {
	return OpenOptions{BusyTimeout: 5 * time.Second}
}

func NewDatabase(dbPath string) (*Database, error) {
	return NewDatabaseWithOptions(dbPath, DefaultOpenOptions())
}

func NewDatabaseWithOptions(dbPath string, options OpenOptions) (*Database, error) {
	db, err := openConnection(dbPath, options)
	if err != nil {
		return nil, err
	}
	database := &Database{conn: db}
	if err := migrateDatabase(db, schemaFS, "schemas"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := configureRuntimeConnection(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := validateDatabase(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return database, nil
}

func openConnection(dbPath string, options OpenOptions) (*sql.DB, error) {
	dsn, err := sqliteConnectionDSN(dbPath, options)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// One connection is sufficient for the local desktop app. The DSN still
	// configures every physical connection because database/sql may discard and
	// replace this connection after an interruption or driver error.
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open sqlite connection: %w", err)
	}
	return db, nil
}

func sqliteConnectionDSN(dbPath string, options OpenOptions) (string, error) {
	if options.BusyTimeout < 0 {
		return "", errors.New("database busy timeout cannot be negative")
	}
	busyTimeoutMS := options.BusyTimeout.Milliseconds()
	if options.BusyTimeout > 0 && busyTimeoutMS == 0 {
		busyTimeoutMS = 1
	}

	query := url.Values{}
	for _, pragma := range []string{
		fmt.Sprintf("busy_timeout=%d", busyTimeoutMS),
		"foreign_keys=ON",
		"synchronous=NORMAL",
		"trusted_schema=OFF",
	} {
		query.Add("_pragma", pragma)
	}
	separator := "?"
	if strings.Contains(dbPath, "?") {
		separator = "&"
	}
	return dbPath + separator + query.Encode(), nil
}

func configureRuntimeConnection(db *sql.DB) error {
	deadline := time.Now().Add(5 * time.Second)
	for {
		var journalMode string
		err := db.QueryRow("PRAGMA journal_mode = WAL").Scan(&journalMode)
		if err == nil {
			if !strings.EqualFold(journalMode, "wal") && journalMode != "memory" {
				return fmt.Errorf("configure sqlite runtime: journal_mode=%s", journalMode)
			}
			break
		}
		message := strings.ToLower(err.Error())
		if time.Now().After(deadline) ||
			(!strings.Contains(message, "busy") && !strings.Contains(message, "locked")) {
			return fmt.Errorf("configure sqlite runtime: %w", err)
		}
		time.Sleep(25 * time.Millisecond)
	}
	return nil
}

func (d *Database) Close() error {
	return d.conn.Close()
}

// ExecContext, PrepareContext, QueryContext, and QueryRowContext intentionally
// make Database satisfy sqlc's DBTX contract without exposing the underlying
// connection pool.
func (d *Database) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.conn.ExecContext(ctx, query, args...)
}

func (d *Database) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return d.conn.PrepareContext(ctx, query)
}

func (d *Database) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.conn.QueryContext(ctx, query, args...)
}

func (d *Database) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.conn.QueryRowContext(ctx, query, args...)
}
