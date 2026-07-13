package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"slices"

	_ "modernc.org/sqlite"
)

//go:embed schemas/*.sql
var schemaFS embed.FS

type Database struct {
	Conn *sql.DB
	path string
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := openConnection(dbPath)
	if err != nil {
		return nil, err
	}
	database := &Database{Conn: db, path: dbPath}
	if err := database.createTables(); err != nil {
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
	for _, statement := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA journal_mode = WAL",
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

func (d *Database) createTables() error {
	schema, err := schemaFS.ReadDir("schemas")
	if err != nil {
		return fmt.Errorf("read schemas: %w", err)
	}
	slices.SortFunc(schema, func(a, b fs.DirEntry) int {
		if a.Name() < b.Name() {
			return -1
		}
		if a.Name() > b.Name() {
			return 1
		}
		return 0
	})

	tx, err := d.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	for _, file := range schema {
		var applied int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, file.Name()).Scan(&applied); err != nil {
			return err
		}
		if applied > 0 {
			continue
		}
		content, err := schemaFS.ReadFile("schemas/" + file.Name())
		if err != nil {
			return fmt.Errorf("read schema %s: %w", file.Name(), err)
		}
		if _, err := tx.Exec(string(content)); err != nil {
			log.Printf("failed applying schema %s: %v", file.Name(), err)
			return err
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, file.Name()); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *Database) Close() error {
	return d.Conn.Close()
}

func (d *Database) GetConnection() *sql.DB {
	return d.Conn
}
