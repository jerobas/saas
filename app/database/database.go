package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Database struct {
	Conn *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := openConnection(dbPath)
	if err != nil {
		return nil, err
	}
	database := &Database{Conn: db}
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

func openConnection(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	// SQLite PRAGMAs are connection-scoped. A single connection is sufficient
	// for this local desktop app and keeps these settings deterministic.
	db.SetMaxOpenConns(1)
	// These connection-local settings do not alter an unidentified database.
	// File-persistent settings are applied only after migration identity has
	// been accepted, so an old or foreign database is rejected in place.
	for _, statement := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA trusted_schema = OFF",
	} {
		if _, err := db.Exec(statement); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("configure sqlite: %w", err)
		}
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
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
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		return fmt.Errorf("configure sqlite runtime: %w", err)
	}
	return nil
}

func (d *Database) Close() error {
	return d.Conn.Close()
}

func (d *Database) GetConnection() *sql.DB {
	return d.Conn
}
