package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const applicationID = 1398228308 // ASCII "SWET"

// migrationTableDDL is part of the persisted database identity. Change it only
// through an explicitly designed metadata migration, never as a formatting edit.
const migrationTableDDL = `CREATE TABLE schema_migrations (
	version INTEGER PRIMARY KEY CHECK (version > 0),
	name TEXT NOT NULL UNIQUE,
	checksum TEXT NOT NULL UNIQUE CHECK (
		length(checksum) = 64
		AND checksum NOT GLOB '*[^0-9a-f]*'
	),
	applied_at_unix_ms INTEGER NOT NULL CHECK (applied_at_unix_ms > 0)
) STRICT`

var (
	ErrUnrecognizedDatabase = errors.New("database is not an empty or recognized Sweeters database")
	ErrWrongApplication     = errors.New("database belongs to another application")
	ErrDatabaseTooNew       = errors.New("database schema is newer than this application")
	ErrMigrationHistory     = errors.New("database migration history is invalid")
	ErrMigrationChecksum    = errors.New("database migration checksum does not match")
)

//go:embed schemas/*.sql
var schemaFS embed.FS

var migrationNamePattern = regexp.MustCompile(`^(\d{4})_([a-z0-9]+(?:_[a-z0-9]+)*)\.sql$`)

type migration struct {
	version  int
	name     string
	checksum string
	content  []byte
}

func migrateDatabase(db *sql.DB, source fs.FS, directory string) error {
	migrations, err := loadMigrations(source, directory)
	if err != nil {
		return err
	}
	if err := prepareMigrationStore(db); err != nil {
		return err
	}
	if err := validateMigrationHistory(db, migrations, false); err != nil {
		return err
	}
	for _, item := range migrations {
		if err := ensureMigrationApplied(db, item); err != nil {
			return fmt.Errorf("apply migration %s: %w", item.name, err)
		}
	}
	return validateMigrationHistory(db, migrations, true)
}

func loadMigrations(source fs.FS, directory string) ([]migration, error) {
	entries, err := fs.ReadDir(source, directory)
	if err != nil {
		return nil, fmt.Errorf("read migration directory: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, fmt.Errorf("migration directory contains subdirectory %q", entry.Name())
		}
		matches := migrationNamePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			return nil, fmt.Errorf("invalid migration filename %q", entry.Name())
		}
		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", matches[1], err)
		}
		expectedVersion := len(migrations) + 1
		if version != expectedVersion {
			return nil, fmt.Errorf("migration %q has version %d; expected %d", entry.Name(), version, expectedVersion)
		}
		content, err := fs.ReadFile(source, path.Join(directory, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}
		checksum := fmt.Sprintf("%x", sha256.Sum256(content))
		migrations = append(migrations, migration{
			version:  version,
			name:     entry.Name(),
			checksum: checksum,
			content:  content,
		})
	}
	if len(migrations) == 0 {
		return nil, errors.New("no embedded migrations found")
	}
	return migrations, nil
}

func prepareMigrationStore(db *sql.DB) error {
	return withImmediateTransaction(db, func(conn *sql.Conn) error {
		application, err := readPragmaInt(conn, "application_id")
		if err != nil {
			return err
		}
		exists, err := schemaMigrationTableExists(conn)
		if err != nil {
			return err
		}
		if exists {
			if err := validateMigrationTableShape(conn); err != nil {
				return err
			}
			if application != applicationID {
				if application == 0 {
					return fmt.Errorf("%w: legacy migration table has no application identity", ErrUnrecognizedDatabase)
				}
				return fmt.Errorf("%w: application_id=%d", ErrWrongApplication, application)
			}
			var recordedMigrations int
			if err := conn.QueryRowContext(
				context.Background(),
				"SELECT COUNT(*) FROM schema_migrations",
			).Scan(&recordedMigrations); err != nil {
				return err
			}
			userVersion, err := readPragmaInt(conn, "user_version")
			if err != nil {
				return err
			}
			if recordedMigrations == 0 && userVersion == 0 {
				objectCount, err := userObjectCount(conn)
				if err != nil {
					return err
				}
				if objectCount != 1 {
					return fmt.Errorf(
						"%w: version-zero database contains %d user objects",
						ErrUnrecognizedDatabase,
						objectCount,
					)
				}
			}
			return nil
		}

		objectCount, err := userObjectCount(conn)
		if err != nil {
			return err
		}
		if objectCount != 0 || application != 0 {
			if application != 0 && application != applicationID {
				return fmt.Errorf("%w: application_id=%d", ErrWrongApplication, application)
			}
			return fmt.Errorf("%w: found %d user objects", ErrUnrecognizedDatabase, objectCount)
		}

		if _, err := conn.ExecContext(context.Background(), migrationTableDDL); err != nil {
			return fmt.Errorf("create migration table: %w", err)
		}
		if _, err := conn.ExecContext(context.Background(), fmt.Sprintf("PRAGMA application_id = %d", applicationID)); err != nil {
			return fmt.Errorf("set application identity: %w", err)
		}
		return nil
	})
}

func schemaMigrationTableExists(queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}) (bool, error) {
	var count int
	err := queryer.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM sqlite_schema
		WHERE type = 'table' AND name = 'schema_migrations'
	`).Scan(&count)
	return count == 1, err
}

func validateMigrationTableShape(conn *sql.Conn) error {
	var storedDDL string
	if err := conn.QueryRowContext(context.Background(), `
		SELECT sql FROM sqlite_schema
		WHERE type = 'table' AND name = 'schema_migrations'
	`).Scan(&storedDDL); err != nil {
		return err
	}
	if strings.Join(strings.Fields(storedDDL), " ") !=
		strings.Join(strings.Fields(migrationTableDDL), " ") {
		return fmt.Errorf("%w: schema_migrations definition differs from the application contract", ErrUnrecognizedDatabase)
	}

	var strict int
	if err := conn.QueryRowContext(context.Background(), `
		SELECT strict FROM pragma_table_list
		WHERE schema = 'main' AND type = 'table' AND name = 'schema_migrations'
	`).Scan(&strict); err != nil {
		return err
	}
	if strict != 1 {
		return fmt.Errorf("%w: schema_migrations is not STRICT", ErrUnrecognizedDatabase)
	}

	rows, err := conn.QueryContext(context.Background(), "PRAGMA table_xinfo(schema_migrations)")
	if err != nil {
		return err
	}
	defer rows.Close()

	type wantedColumn struct {
		columnType string
		notNull    int
		primaryKey int
	}
	wanted := map[string]wantedColumn{
		"version":            {columnType: "INTEGER", notNull: 0, primaryKey: 1},
		"name":               {columnType: "TEXT", notNull: 1, primaryKey: 0},
		"checksum":           {columnType: "TEXT", notNull: 1, primaryKey: 0},
		"applied_at_unix_ms": {columnType: "INTEGER", notNull: 1, primaryKey: 0},
	}
	columnCount := 0
	for rows.Next() {
		var cid, notNull, primaryKey, hidden int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey, &hidden); err != nil {
			return err
		}
		columnCount++
		expected, ok := wanted[name]
		if !ok || hidden != 0 || !strings.EqualFold(columnType, expected.columnType) ||
			notNull != expected.notNull || primaryKey != expected.primaryKey {
			return fmt.Errorf("%w: unexpected schema_migrations column %q", ErrUnrecognizedDatabase, name)
		}
		delete(wanted, name)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for name := range wanted {
		return fmt.Errorf("%w: schema_migrations is missing %q", ErrUnrecognizedDatabase, name)
	}
	if columnCount != 4 {
		return fmt.Errorf("%w: schema_migrations has %d columns; expected 4", ErrMigrationHistory, columnCount)
	}
	return nil
}

func userObjectCount(conn *sql.Conn) (int, error) {
	var count int
	err := conn.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM sqlite_schema
		WHERE name NOT LIKE 'sqlite_%'
		  AND type IN ('table', 'index', 'trigger', 'view')
	`).Scan(&count)
	return count, err
}

func validateMigrationHistory(db *sql.DB, migrations []migration, requireComplete bool) error {
	application, err := readPragmaInt(db, "application_id")
	if err != nil {
		return err
	}
	if application != applicationID {
		return fmt.Errorf("%w: application_id=%d", ErrWrongApplication, application)
	}

	rows, err := db.Query(`SELECT version, name, checksum FROM schema_migrations ORDER BY version`)
	if err != nil {
		return err
	}
	defer rows.Close()

	applied := 0
	for rows.Next() {
		var version int
		var name, checksum string
		if err := rows.Scan(&version, &name, &checksum); err != nil {
			return err
		}
		applied++
		if version > len(migrations) {
			return fmt.Errorf("%w: migration version %d", ErrDatabaseTooNew, version)
		}
		if version != applied {
			return fmt.Errorf("%w: expected version %d, found %d", ErrMigrationHistory, applied, version)
		}
		expected := migrations[version-1]
		if name != expected.name {
			return fmt.Errorf("%w: version %d is %q, expected %q", ErrMigrationHistory, version, name, expected.name)
		}
		if checksum != expected.checksum {
			return fmt.Errorf("%w: %s", ErrMigrationChecksum, name)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	userVersion, err := readPragmaInt(db, "user_version")
	if err != nil {
		return err
	}
	if userVersion > len(migrations) {
		return fmt.Errorf("%w: user_version=%d", ErrDatabaseTooNew, userVersion)
	}
	if userVersion != applied {
		return fmt.Errorf("%w: user_version=%d but %d migrations are recorded", ErrMigrationHistory, userVersion, applied)
	}
	if requireComplete && applied != len(migrations) {
		return fmt.Errorf("%w: applied %d of %d migrations", ErrMigrationHistory, applied, len(migrations))
	}
	return nil
}

func ensureMigrationApplied(db *sql.DB, item migration) error {
	return withImmediateTransaction(db, func(conn *sql.Conn) error {
		var name, checksum string
		err := conn.QueryRowContext(context.Background(), `
			SELECT name, checksum FROM schema_migrations WHERE version = ?
		`, item.version).Scan(&name, &checksum)
		if err == nil {
			if name != item.name {
				return fmt.Errorf("%w: version %d is %q, expected %q", ErrMigrationHistory, item.version, name, item.name)
			}
			if checksum != item.checksum {
				return fmt.Errorf("%w: %s", ErrMigrationChecksum, item.name)
			}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		userVersion, err := readPragmaInt(conn, "user_version")
		if err != nil {
			return err
		}
		if userVersion != item.version-1 {
			return fmt.Errorf("%w: cannot apply version %d after user_version %d", ErrMigrationHistory, item.version, userVersion)
		}
		if _, err := conn.ExecContext(context.Background(), string(item.content)); err != nil {
			return err
		}
		if _, err := conn.ExecContext(context.Background(), `
			INSERT INTO schema_migrations (version, name, checksum, applied_at_unix_ms)
			VALUES (?, ?, ?, ?)
		`, item.version, item.name, item.checksum, time.Now().UTC().UnixMilli()); err != nil {
			return err
		}
		if _, err := conn.ExecContext(context.Background(), fmt.Sprintf("PRAGMA user_version = %d", item.version)); err != nil {
			return err
		}
		return nil
	})
}

func withImmediateTransaction(db *sql.DB, operation func(*sql.Conn) error) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(ctx, "ROLLBACK")
		}
	}()
	if err := operation(conn); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		return err
	}
	committed = true
	return nil
}

func readPragmaInt(queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, name string) (int, error) {
	if name != "application_id" && name != "user_version" {
		return 0, fmt.Errorf("unsupported integer pragma %q", name)
	}
	var value int
	if err := queryer.QueryRowContext(context.Background(), "PRAGMA "+name).Scan(&value); err != nil {
		return 0, err
	}
	return value, nil
}

func validateDatabase(db *sql.DB) error {
	application, err := readPragmaInt(db, "application_id")
	if err != nil {
		return err
	}
	if application != applicationID {
		return fmt.Errorf("%w: application_id=%d", ErrWrongApplication, application)
	}

	rows, err := db.Query("PRAGMA quick_check")
	if err != nil {
		return err
	}
	for rows.Next() {
		var result string
		if err := rows.Scan(&result); err != nil {
			rows.Close()
			return err
		}
		if !strings.EqualFold(result, "ok") {
			rows.Close()
			return fmt.Errorf("sqlite quick_check failed: %s", result)
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}

	foreignKeys, err := db.Query("PRAGMA foreign_key_check")
	if err != nil {
		return err
	}
	defer foreignKeys.Close()
	if foreignKeys.Next() {
		var table, parent string
		var rowID sql.NullInt64
		var foreignKeyID int
		if err := foreignKeys.Scan(&table, &rowID, &parent, &foreignKeyID); err != nil {
			return err
		}
		return fmt.Errorf("foreign key violation in %s row %v referencing %s (fk %d)", table, rowID, parent, foreignKeyID)
	}
	return foreignKeys.Err()
}
