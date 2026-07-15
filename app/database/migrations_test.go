package database

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestMigrateDatabaseInitializesIdentityVersionAndHistory(t *testing.T) {
	db := openMigrationTestDatabase(t)
	source := migrationTestFS(map[string]string{
		"0001_create_widgets.sql":  `CREATE TABLE widgets (id INTEGER PRIMARY KEY) STRICT;`,
		"0002_add_widget_name.sql": `ALTER TABLE widgets ADD COLUMN name TEXT;`,
	})

	if err := migrateDatabase(db, source, "migrations"); err != nil {
		t.Fatalf("migrate fresh database: %v", err)
	}

	application, err := readPragmaInt(db, "application_id")
	if err != nil {
		t.Fatal(err)
	}
	if application != applicationID {
		t.Fatalf("application_id = %d, want %d", application, applicationID)
	}

	version, err := readPragmaInt(db, "user_version")
	if err != nil {
		t.Fatal(err)
	}
	if version != 2 {
		t.Fatalf("user_version = %d, want 2", version)
	}

	type appliedMigration struct {
		version     int
		name        string
		checksum    string
		appliedAtMS int64
	}
	rows, err := db.Query(`
		SELECT version, name, checksum, applied_at_unix_ms
		FROM schema_migrations
		ORDER BY version
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var applied []appliedMigration
	for rows.Next() {
		var item appliedMigration
		if err := rows.Scan(&item.version, &item.name, &item.checksum, &item.appliedAtMS); err != nil {
			t.Fatal(err)
		}
		applied = append(applied, item)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(applied) != 2 {
		t.Fatalf("applied migration count = %d, want 2", len(applied))
	}

	want := []struct {
		name    string
		content string
	}{
		{"0001_create_widgets.sql", `CREATE TABLE widgets (id INTEGER PRIMARY KEY) STRICT;`},
		{"0002_add_widget_name.sql", `ALTER TABLE widgets ADD COLUMN name TEXT;`},
	}
	for index, expected := range want {
		item := applied[index]
		if item.version != index+1 {
			t.Errorf("migration %d version = %d, want %d", index, item.version, index+1)
		}
		if item.name != expected.name {
			t.Errorf("migration %d name = %q, want %q", index, item.name, expected.name)
		}
		if item.checksum != migrationChecksum(expected.content) {
			t.Errorf("migration %d checksum = %q, want SHA-256 of embedded content", index, item.checksum)
		}
		if item.appliedAtMS <= 0 || item.appliedAtMS > time.Now().Add(time.Minute).UnixMilli() {
			t.Errorf("migration %d applied_at_unix_ms = %d, want a current positive timestamp", index, item.appliedAtMS)
		}
	}
}

func TestRecipeOutputGuardForwardMigration(t *testing.T) {
	db := openMigrationTestDatabase(t)
	if err := migrateDatabase(db, embeddedBaselineOnlyFS(t), "schemas"); err != nil {
		t.Fatalf("apply embedded baseline: %v", err)
	}
	result, err := db.Exec(`
		INSERT INTO items (
			name, normalized_name, base_unit_code,
			is_purchasable, is_producible, is_sellable,
			created_at_ms, updated_at_ms
		) VALUES ('Output', 'output', 'g', 1, 1, 0, 1, 1)
	`)
	if err != nil {
		t.Fatal(err)
	}
	outputID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	result, err = db.Exec(`
		INSERT INTO recipes (
			name, normalized_name, output_item_id, created_at_ms, updated_at_ms
		) VALUES ('Migrated recipe', 'migrated recipe', ?, 1, 1)
	`, outputID)
	if err != nil {
		t.Fatal(err)
	}
	recipeID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO recipe_revisions (
			recipe_id, revision_number, standard_yield_quantity_atomic,
			instructions, preparation_time_minutes, created_at_ms
		) VALUES (?, 1, 1000, '', 0, 1)
	`, recipeID); err != nil {
		t.Fatal(err)
	}

	if err := migrateDatabase(db, schemaFS, "schemas"); err != nil {
		t.Fatalf("apply forward migrations: %v", err)
	}
	version, err := readPragmaInt(db, "user_version")
	if err != nil {
		t.Fatal(err)
	}
	if version != 2 {
		t.Fatalf("user_version = %d, want 2", version)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("migration count = %d, want 2", count)
	}
	expectExecError(t, db, `UPDATE items SET is_producible = 0, updated_at_ms = 2 WHERE id = ?`, outputID)
	expectExecError(t, db, `UPDATE items SET archived_at_ms = 2, updated_at_ms = 2 WHERE id = ?`, outputID)
}

func TestRecipeChainMigrationRejectsExistingGapAtomically(t *testing.T) {
	db := openMigrationTestDatabase(t)
	if err := migrateDatabase(db, embeddedBaselineOnlyFS(t), "schemas"); err != nil {
		t.Fatalf("apply embedded baseline: %v", err)
	}
	result, err := db.Exec(`
		INSERT INTO items (
			name, normalized_name, base_unit_code,
			is_purchasable, is_producible, is_sellable,
			created_at_ms, updated_at_ms
		) VALUES ('Gap output', 'gap output', 'g', 0, 1, 0, 1, 1)
	`)
	if err != nil {
		t.Fatal(err)
	}
	outputID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	result, err = db.Exec(`
		INSERT INTO recipes (
			name, normalized_name, output_item_id, created_at_ms, updated_at_ms
		) VALUES ('Gap recipe', 'gap recipe', ?, 1, 1)
	`, outputID)
	if err != nil {
		t.Fatal(err)
	}
	recipeID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO recipe_revisions (
			recipe_id, revision_number, standard_yield_quantity_atomic,
			instructions, preparation_time_minutes, created_at_ms
		) VALUES (?, 2, 1000, '', 0, 1)
	`, recipeID); err != nil {
		t.Fatal(err)
	}

	if err := migrateDatabase(db, schemaFS, "schemas"); err == nil {
		t.Fatal("migration accepted an existing discontinuous recipe chain")
	}
	version, err := readPragmaInt(db, "user_version")
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Fatalf("user_version after failed migration = %d, want 1", version)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("migration count after failed migration = %d, want 1", count)
	}
	if databaseObjectExists(t, db, "trigger", "recipe_revisions_require_next_number") {
		t.Fatal("failed migration left its recipe sequencing trigger installed")
	}
}

func TestArchiveVersionMigrationRejectsExistingMismatchAtomically(t *testing.T) {
	db := openMigrationTestDatabase(t)
	if err := migrateDatabase(db, embeddedBaselineOnlyFS(t), "schemas"); err != nil {
		t.Fatalf("apply embedded baseline: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO items (
			name, normalized_name, base_unit_code,
			is_purchasable, is_producible, is_sellable,
			created_at_ms, updated_at_ms, archived_at_ms
		) VALUES ('Old archive', 'old archive', 'g', 0, 0, 0, 1, 1, 2)
	`); err != nil {
		t.Fatal(err)
	}

	if err := migrateDatabase(db, schemaFS, "schemas"); err == nil {
		t.Fatal("migration accepted an archive timestamp newer than the optimistic version")
	}
	version, err := readPragmaInt(db, "user_version")
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Fatalf("user_version after failed migration = %d, want 1", version)
	}
	if databaseObjectExists(t, db, "trigger", "items_archive_version_update") {
		t.Fatal("failed migration left its archive-version trigger installed")
	}
}

func TestMigrateDatabaseIsIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "reopen.db")
	db := openMigrationTestDatabaseAt(t, dbPath)
	source := migrationTestFS(map[string]string{
		"0001_create_widgets.sql": `CREATE TABLE widgets (id INTEGER PRIMARY KEY) STRICT;`,
	})

	if err := migrateDatabase(db, source, "migrations"); err != nil {
		t.Fatalf("first migration run: %v", err)
	}
	var firstAppliedAt int64
	if err := db.QueryRow(`SELECT applied_at_unix_ms FROM schema_migrations WHERE version = 1`).Scan(&firstAppliedAt); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close migrated database: %v", err)
	}

	db = openMigrationTestDatabaseAt(t, dbPath)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close reopened test database: %v", err)
		}
	})

	if err := migrateDatabase(db, source, "migrations"); err != nil {
		t.Fatalf("migration after reopen: %v", err)
	}

	var count int
	var secondAppliedAt int64
	if err := db.QueryRow(`SELECT COUNT(*), MIN(applied_at_unix_ms) FROM schema_migrations`).Scan(&count, &secondAppliedAt); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("migration count after reopen = %d, want 1", count)
	}
	if secondAppliedAt != firstAppliedAt {
		t.Fatalf("applied timestamp changed from %d to %d", firstAppliedAt, secondAppliedAt)
	}
}

func TestLoadMigrationsRejectsInvalidAndGappedFilenames(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			name:  "invalid filename",
			files: map[string]string{"001_bad.sql": `SELECT 1;`},
			want:  "invalid migration filename",
		},
		{
			name: "version gap",
			files: map[string]string{
				"0001_first.sql": `SELECT 1;`,
				"0003_third.sql": `SELECT 3;`,
			},
			want: "expected 2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadMigrations(migrationTestFS(test.files), "migrations")
			if err == nil {
				t.Fatal("loadMigrations returned nil error")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %q, want it to contain %q", err, test.want)
			}
		})
	}
}

func TestMigrateDatabaseRollsBackFailedMigration(t *testing.T) {
	db := openMigrationTestDatabase(t)
	source := migrationTestFS(map[string]string{
		"0001_create_widgets.sql": `CREATE TABLE widgets (id INTEGER PRIMARY KEY) STRICT;`,
		"0002_broken_change.sql": `
			CREATE TABLE partial_change (id INTEGER PRIMARY KEY) STRICT;
			INSERT INTO missing_table (id) VALUES (1);
		`,
	})

	err := migrateDatabase(db, source, "migrations")
	if err == nil {
		t.Fatal("migrateDatabase returned nil error for broken migration")
	}
	if !databaseObjectExists(t, db, "table", "widgets") {
		t.Fatal("first committed migration was lost")
	}
	if databaseObjectExists(t, db, "table", "partial_change") {
		t.Fatal("table from failed migration was not rolled back")
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("recorded migration count = %d, want 1", count)
	}
	version, err := readPragmaInt(db, "user_version")
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Fatalf("user_version = %d, want 1 after rollback", version)
	}
}

func TestMigrateDatabaseRejectsChecksumMismatch(t *testing.T) {
	db := openMigrationTestDatabase(t)
	original := migrationTestFS(map[string]string{
		"0001_create_widgets.sql": `CREATE TABLE widgets (id INTEGER PRIMARY KEY) STRICT;`,
	})
	if err := migrateDatabase(db, original, "migrations"); err != nil {
		t.Fatal(err)
	}

	var recordedChecksum string
	if err := db.QueryRow(`SELECT checksum FROM schema_migrations WHERE version = 1`).Scan(&recordedChecksum); err != nil {
		t.Fatal(err)
	}
	changed := migrationTestFS(map[string]string{
		"0001_create_widgets.sql": `CREATE TABLE widgets (id INTEGER PRIMARY KEY, name TEXT) STRICT;`,
	})

	err := migrateDatabase(db, changed, "migrations")
	if !errors.Is(err, ErrMigrationChecksum) {
		t.Fatalf("error = %v, want ErrMigrationChecksum", err)
	}
	var checksumAfterFailure string
	if err := db.QueryRow(`SELECT checksum FROM schema_migrations WHERE version = 1`).Scan(&checksumAfterFailure); err != nil {
		t.Fatal(err)
	}
	if checksumAfterFailure != recordedChecksum {
		t.Fatalf("recorded checksum changed from %q to %q", recordedChecksum, checksumAfterFailure)
	}
}

func TestMigrateDatabaseRejectsLegacyDatabaseWithoutAdoptingIt(t *testing.T) {
	db := openMigrationTestDatabase(t)
	if _, err := db.Exec(`CREATE TABLE legacy_items (id INTEGER PRIMARY KEY, name TEXT); INSERT INTO legacy_items VALUES (1, 'kept');`); err != nil {
		t.Fatal(err)
	}

	err := migrateDatabase(db, oneMigrationFS(), "migrations")
	if !errors.Is(err, ErrUnrecognizedDatabase) {
		t.Fatalf("error = %v, want ErrUnrecognizedDatabase", err)
	}
	if databaseObjectExists(t, db, "table", "schema_migrations") {
		t.Fatal("legacy database was destructively adopted")
	}
	var name string
	if err := db.QueryRow(`SELECT name FROM legacy_items WHERE id = 1`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "kept" {
		t.Fatalf("legacy data = %q, want %q", name, "kept")
	}
	application, err := readPragmaInt(db, "application_id")
	if err != nil {
		t.Fatal(err)
	}
	if application != 0 {
		t.Fatalf("legacy application_id changed to %d", application)
	}
}

func TestMigrateDatabaseRejectsForeignDatabaseWithoutAdoptingIt(t *testing.T) {
	db := openMigrationTestDatabase(t)
	const foreignApplicationID = 123456789
	if _, err := db.Exec(fmt.Sprintf("PRAGMA application_id = %d", foreignApplicationID)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE foreign_data (value TEXT); INSERT INTO foreign_data VALUES ('kept');`); err != nil {
		t.Fatal(err)
	}

	err := migrateDatabase(db, oneMigrationFS(), "migrations")
	if !errors.Is(err, ErrWrongApplication) {
		t.Fatalf("error = %v, want ErrWrongApplication", err)
	}
	if databaseObjectExists(t, db, "table", "schema_migrations") {
		t.Fatal("foreign database was destructively adopted")
	}
	var value string
	if err := db.QueryRow(`SELECT value FROM foreign_data`).Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "kept" {
		t.Fatalf("foreign data = %q, want %q", value, "kept")
	}
	application, err := readPragmaInt(db, "application_id")
	if err != nil {
		t.Fatal(err)
	}
	if application != foreignApplicationID {
		t.Fatalf("foreign application_id changed to %d", application)
	}
}

func TestMigrateDatabaseRejectsMalformedMigrationStore(t *testing.T) {
	db := openMigrationTestDatabase(t)
	if _, err := db.Exec(fmt.Sprintf(`
		PRAGMA application_id = %d;
		CREATE TABLE schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			checksum TEXT NOT NULL,
			applied_at_unix_ms INTEGER NOT NULL
		) STRICT;
	`, applicationID)); err != nil {
		t.Fatal(err)
	}

	err := migrateDatabase(db, oneMigrationFS(), "migrations")
	if !errors.Is(err, ErrUnrecognizedDatabase) {
		t.Fatalf("error = %v, want ErrUnrecognizedDatabase", err)
	}
	if databaseObjectExists(t, db, "table", "widgets") {
		t.Fatal("migration ran against malformed metadata")
	}
}

func TestMigrateDatabaseRejectsObjectsBesideAnEmptyMigrationStore(t *testing.T) {
	db := openMigrationTestDatabase(t)
	if _, err := db.Exec(migrationTableDDL); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA application_id = %d", applicationID)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE unrelated_data (value TEXT) STRICT`); err != nil {
		t.Fatal(err)
	}

	err := migrateDatabase(db, oneMigrationFS(), "migrations")
	if !errors.Is(err, ErrUnrecognizedDatabase) {
		t.Fatalf("error = %v, want ErrUnrecognizedDatabase", err)
	}
	if databaseObjectExists(t, db, "table", "widgets") {
		t.Fatal("migration ran beside an unrelated version-zero object")
	}
}

func TestMigrateDatabaseRejectsNewerSchema(t *testing.T) {
	db := openMigrationTestDatabase(t)
	source := oneMigrationFS()
	if err := migrateDatabase(db, source, "migrations"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE future_data (value TEXT); INSERT INTO future_data VALUES ('kept');`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO schema_migrations (version, name, checksum, applied_at_unix_ms)
		VALUES (2, '0002_future.sql', ?, 1)
	`, strings.Repeat("a", 64)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`PRAGMA user_version = 2`); err != nil {
		t.Fatal(err)
	}

	err := migrateDatabase(db, source, "migrations")
	if !errors.Is(err, ErrDatabaseTooNew) {
		t.Fatalf("error = %v, want ErrDatabaseTooNew", err)
	}
	var value string
	if err := db.QueryRow(`SELECT value FROM future_data`).Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "kept" {
		t.Fatalf("future data = %q, want %q", value, "kept")
	}
}

func openMigrationTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	db := openMigrationTestDatabaseAt(t, filepath.Join(t.TempDir(), "migration-test.db"))
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})
	return db
}

func openMigrationTestDatabaseAt(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	return db
}

func migrationTestFS(files map[string]string) fstest.MapFS {
	source := make(fstest.MapFS, len(files))
	for name, content := range files {
		source["migrations/"+name] = &fstest.MapFile{Data: []byte(content), Mode: 0o444}
	}
	return source
}

func embeddedBaselineOnlyFS(t *testing.T) fstest.MapFS {
	t.Helper()
	content, err := schemaFS.ReadFile("schemas/0001_v2_baseline.sql")
	if err != nil {
		t.Fatal(err)
	}
	return fstest.MapFS{
		"schemas/0001_v2_baseline.sql": &fstest.MapFile{Data: content, Mode: 0o444},
	}
}

func oneMigrationFS() fstest.MapFS {
	return migrationTestFS(map[string]string{
		"0001_create_widgets.sql": `CREATE TABLE widgets (id INTEGER PRIMARY KEY) STRICT;`,
	})
}

func migrationChecksum(content string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
}

func databaseObjectExists(t *testing.T, db *sql.DB, objectType, name string) bool {
	t.Helper()
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_schema WHERE type = ? AND name = ?
	`, objectType, name).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count == 1
}
