package database

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDatabaseReadMaintainsSnapshotAcrossExternalCommit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "read-snapshot.db")
	reader := newTransactionTestDatabase(t, path, DefaultOpenOptions())
	writer := newTransactionTestDatabase(t, path, DefaultOpenOptions())
	ctx := context.Background()

	var firstName, secondLocale string
	err := reader.Read(ctx, func(tx *ReadTx) error {
		if err := tx.QueryRowContext(ctx, `
			SELECT business_name FROM app_settings WHERE id = 1
		`).Scan(&firstName); err != nil {
			return err
		}
		if err := writer.Write(ctx, func(tx *WriteTx) error {
			_, err := tx.ExecContext(ctx, `
				UPDATE app_settings
				SET business_name = 'External', locale_code = 'en-US', updated_at_ms = 1
				WHERE id = 1
			`)
			return err
		}); err != nil {
			return err
		}
		return tx.QueryRowContext(ctx, `
			SELECT locale_code FROM app_settings WHERE id = 1
		`).Scan(&secondLocale)
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if firstName != "Sweeters" || secondLocale != "pt-BR" {
		t.Fatalf("read snapshot = %q/%q, want Sweeters/pt-BR", firstName, secondLocale)
	}

	var latestName, latestLocale string
	if err := reader.QueryRowContext(ctx, `
		SELECT business_name, locale_code FROM app_settings WHERE id = 1
	`).Scan(&latestName, &latestLocale); err != nil {
		t.Fatal(err)
	}
	if latestName != "External" || latestLocale != "en-US" {
		t.Fatalf("latest settings = %q/%q, want External/en-US", latestName, latestLocale)
	}
}

func TestDatabaseReadRejectsQueryReturningWritesAndRestoresConnection(t *testing.T) {
	db := newTransactionTestDatabase(t, filepath.Join(t.TempDir(), "read-only.db"), DefaultOpenOptions())
	ctx := context.Background()

	err := db.Read(ctx, func(tx *ReadTx) error {
		var name string
		return tx.QueryRowContext(ctx, `
			UPDATE app_settings SET business_name = 'Forbidden', updated_at_ms = 1
			WHERE id = 1
			RETURNING business_name
		`).Scan(&name)
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "readonly") {
		t.Fatalf("read transaction write error = %v, want readonly error", err)
	}

	if err := db.Write(ctx, func(tx *WriteTx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE app_settings SET business_name = 'Writable', updated_at_ms = 1
			WHERE id = 1
		`)
		return err
	}); err != nil {
		t.Fatalf("Write after Read: %v", err)
	}
	var name string
	if err := db.QueryRowContext(ctx, "SELECT business_name FROM app_settings WHERE id = 1").Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Writable" {
		t.Fatalf("business name = %q, want Writable", name)
	}
}

func TestDatabaseReadRejectsNilOperation(t *testing.T) {
	db := newTransactionTestDatabase(t, filepath.Join(t.TempDir(), "read-nil.db"), DefaultOpenOptions())
	if err := db.Read(context.Background(), nil); err == nil {
		t.Fatal("Read with nil operation returned nil error")
	}
}

func TestDatabaseReadReturnsCallbackErrorAndRestoresConnection(t *testing.T) {
	db := newTransactionTestDatabase(t, filepath.Join(t.TempDir(), "read-error.db"), DefaultOpenOptions())
	ctx := context.Background()
	wantErr := errors.New("stop read")

	err := db.Read(ctx, func(tx *ReadTx) error {
		var name string
		if err := tx.QueryRowContext(ctx, "SELECT business_name FROM app_settings WHERE id = 1").Scan(&name); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Read error = %v, want callback error", err)
	}
	assertDatabaseWritable(t, db, "After callback error")
}

func TestDatabaseReadRollsBackAndRepanics(t *testing.T) {
	db := newTransactionTestDatabase(t, filepath.Join(t.TempDir(), "read-panic.db"), DefaultOpenOptions())
	ctx := context.Background()
	const panicValue = "read panic"

	func() {
		defer func() {
			if recovered := recover(); recovered != panicValue {
				t.Fatalf("recovered value = %#v, want %q", recovered, panicValue)
			}
		}()
		_ = db.Read(ctx, func(tx *ReadTx) error {
			var name string
			if err := tx.QueryRowContext(ctx, "SELECT business_name FROM app_settings WHERE id = 1").Scan(&name); err != nil {
				t.Fatalf("read before panic: %v", err)
			}
			panic(panicValue)
		})
	}()

	assertDatabaseWritable(t, db, "After panic")
}

func TestDatabaseReadRollsBackCancelledContextBeforeCommit(t *testing.T) {
	db := newTransactionTestDatabase(t, filepath.Join(t.TempDir(), "read-cancelled.db"), DefaultOpenOptions())
	ctx, cancel := context.WithCancel(context.Background())

	err := db.Read(ctx, func(tx *ReadTx) error {
		var name string
		if err := tx.QueryRowContext(ctx, "SELECT business_name FROM app_settings WHERE id = 1").Scan(&name); err != nil {
			return err
		}
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Read error = %v, want context.Canceled", err)
	}
	assertDatabaseWritable(t, db, "After cancellation")
}

func TestDatabaseWriteCommits(t *testing.T) {
	db := newTransactionTestDatabase(t, filepath.Join(t.TempDir(), "commit.db"), DefaultOpenOptions())

	err := db.Write(context.Background(), func(tx *WriteTx) error {
		_, err := tx.ExecContext(context.Background(), `
			UPDATE app_settings
			SET business_name = 'Committed', updated_at_ms = updated_at_ms + 1
			WHERE id = 1
		`)
		return err
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	var name string
	if err := db.QueryRowContext(context.Background(), `
		SELECT business_name FROM app_settings WHERE id = 1
	`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Committed" {
		t.Fatalf("business name = %q, want Committed", name)
	}
}

func TestDatabaseWriteRollsBackCallbackError(t *testing.T) {
	db := newTransactionTestDatabase(t, filepath.Join(t.TempDir(), "callback-error.db"), DefaultOpenOptions())
	wantErr := errors.New("stop write")

	err := db.Write(context.Background(), func(tx *WriteTx) error {
		if _, err := tx.ExecContext(context.Background(), testItemInsertSQL, "Rolled back", "rolled back"); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Write error = %v, want callback error", err)
	}
	assertItemCount(t, db, "rolled back", 0)
}

func TestDatabaseWriteRollsBackAndRepanics(t *testing.T) {
	db := newTransactionTestDatabase(t, filepath.Join(t.TempDir(), "panic.db"), DefaultOpenOptions())
	const panicValue = "write panic"

	func() {
		defer func() {
			if recovered := recover(); recovered != panicValue {
				t.Fatalf("recovered value = %#v, want %q", recovered, panicValue)
			}
		}()
		_ = db.Write(context.Background(), func(tx *WriteTx) error {
			if _, err := tx.ExecContext(context.Background(), testItemInsertSQL, "Panicked", "panicked"); err != nil {
				t.Fatalf("insert before panic: %v", err)
			}
			panic(panicValue)
		})
	}()

	assertItemCount(t, db, "panicked", 0)
}

func TestDatabaseWriteRollsBackCancelledContextBeforeCommit(t *testing.T) {
	db := newTransactionTestDatabase(t, filepath.Join(t.TempDir(), "cancelled.db"), DefaultOpenOptions())
	ctx, cancel := context.WithCancel(context.Background())

	err := db.Write(ctx, func(tx *WriteTx) error {
		if _, err := tx.ExecContext(ctx, testItemInsertSQL, "Cancelled", "cancelled"); err != nil {
			return err
		}
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Write error = %v, want context.Canceled", err)
	}
	assertItemCount(t, db, "cancelled", 0)
}

func TestDatabaseWriteSeesEarlierWritesInSameTransaction(t *testing.T) {
	db := newTransactionTestDatabase(t, filepath.Join(t.TempDir(), "visibility.db"), DefaultOpenOptions())

	err := db.Write(context.Background(), func(tx *WriteTx) error {
		if _, err := tx.ExecContext(context.Background(), testItemInsertSQL, "Visible", "visible"); err != nil {
			return err
		}
		var count int
		if err := tx.QueryRowContext(context.Background(), `
			SELECT COUNT(*) FROM items WHERE normalized_name = 'visible'
		`).Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			return errors.New("insert was not visible inside its transaction")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
}

func TestDatabaseWriteReturnsExternalBusyWithoutRetry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "busy.db")
	options := DefaultOpenOptions()
	options.BusyTimeout = 25 * time.Millisecond
	first := newTransactionTestDatabase(t, path, options)
	second := newTransactionTestDatabase(t, path, options)

	entered := make(chan struct{})
	release := make(chan struct{})
	firstResult := make(chan error, 1)
	go func() {
		firstResult <- first.Write(context.Background(), func(_ *WriteTx) error {
			close(entered)
			<-release
			return nil
		})
	}()
	<-entered

	callbackCalled := false
	err := second.Write(context.Background(), func(_ *WriteTx) error {
		callbackCalled = true
		return nil
	})
	close(release)
	if firstErr := <-firstResult; firstErr != nil {
		t.Fatalf("holding Write: %v", firstErr)
	}

	if err == nil {
		t.Fatal("contending Write returned nil error")
	}
	message := strings.ToLower(err.Error())
	if !strings.Contains(message, "busy") && !strings.Contains(message, "locked") {
		t.Fatalf("contending Write error = %v, want SQLite busy/locked error", err)
	}
	if callbackCalled {
		t.Fatal("contending Write called its callback without acquiring the transaction")
	}
}

const testItemInsertSQL = `
	INSERT INTO items (
		name, normalized_name, base_unit_code,
		is_purchasable, is_producible, is_sellable,
		created_at_ms, updated_at_ms
	) VALUES (?, ?, 'g', 1, 0, 0, 1, 1)
`

func newTransactionTestDatabase(t *testing.T, path string, options OpenOptions) *Database {
	t.Helper()
	db, err := NewDatabaseWithOptions(path, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	return db
}

func assertItemCount(t *testing.T, db *Database, normalizedName string, want int) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM items WHERE normalized_name = ?
	`, normalizedName).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("item count for %q = %d, want %d", normalizedName, count, want)
	}
}

func assertDatabaseWritable(t *testing.T, db *Database, businessName string) {
	t.Helper()
	ctx := context.Background()
	if err := db.Write(ctx, func(tx *WriteTx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE app_settings
			SET business_name = ?, updated_at_ms = updated_at_ms + 1
			WHERE id = 1
		`, businessName)
		return err
	}); err != nil {
		t.Fatalf("Write after Read: %v", err)
	}
	var got string
	if err := db.QueryRowContext(ctx, "SELECT business_name FROM app_settings WHERE id = 1").Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != businessName {
		t.Fatalf("business name = %q, want %q", got, businessName)
	}
}
