package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
	modernsqlite "modernc.org/sqlite"
)

func TestWithWriteQueriesCommits(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "commit.db"), database.DefaultOpenOptions())
	ctx := context.Background()

	err := store.withWriteQueries(ctx, "create item", func(queries *sqlcgen.Queries) error {
		_, err := queries.InsertItem(ctx, testItemParams("Flour", "flour"))
		return err
	})
	if err != nil {
		t.Fatalf("withWriteQueries: %v", err)
	}

	item, err := store.queries.GetItem(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "Flour" {
		t.Fatalf("item name = %q, want Flour", item.Name)
	}
}

func TestWithWriteQueriesRollsBackCallbackError(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "rollback.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	wantErr := errors.New("stop aggregate")

	err := store.withWriteQueries(ctx, "create item", func(queries *sqlcgen.Queries) error {
		if _, err := queries.InsertItem(ctx, testItemParams("Sugar", "sugar")); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("withWriteQueries error = %v, want callback error", err)
	}

	_, err = store.queries.GetItem(ctx, 1)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetItem error = %v, want sql.ErrNoRows after rollback", err)
	}
}

func TestClassifyErrorMapsNoRowsAndPreservesCause(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "not-found.db"), database.DefaultOpenOptions())
	_, cause := store.queries.GetItem(context.Background(), 999)
	mapped := classifyError("get item", cause)

	if !errors.Is(mapped, domain.ErrNotFound) {
		t.Fatalf("mapped error = %v, want domain.ErrNotFound", mapped)
	}
	if !errors.Is(mapped, sql.ErrNoRows) {
		t.Fatalf("mapped error lost sql.ErrNoRows: %v", mapped)
	}
}

func TestClassifyErrorMapsUniqueAndPrimaryKeyConflicts(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "conflicts.db"), database.DefaultOpenOptions())
	ctx := context.Background()

	if err := store.withWriteQueries(ctx, "seed conflicts", func(queries *sqlcgen.Queries) error {
		if _, err := queries.InsertItem(ctx, testItemParams("Butter", "butter")); err != nil {
			return err
		}
		counterpartyID, err := queries.InsertCounterparty(ctx, sqlcgen.InsertCounterpartyParams{
			Name:        "Supplier",
			CreatedAtMs: 1,
			UpdatedAtMs: 1,
		})
		if err != nil {
			return err
		}
		return queries.InsertCounterpartyRole(ctx, sqlcgen.InsertCounterpartyRoleParams{
			CounterpartyID: counterpartyID,
			Role:           "SUPPLIER",
			CreatedAtMs:    1,
		})
	}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		operation func(*sqlcgen.Queries) error
	}{
		{
			name: "unique",
			operation: func(queries *sqlcgen.Queries) error {
				_, err := queries.InsertItem(ctx, testItemParams("BUTTER", "butter"))
				return err
			},
		},
		{
			name: "primary key",
			operation: func(queries *sqlcgen.Queries) error {
				return queries.InsertCounterpartyRole(ctx, sqlcgen.InsertCounterpartyRoleParams{
					CounterpartyID: 1,
					Role:           "SUPPLIER",
					CreatedAtMs:    2,
				})
			},
		},
		{
			name: "check constraint",
			operation: func(queries *sqlcgen.Queries) error {
				return queries.InsertCounterpartyRole(ctx, sqlcgen.InsertCounterpartyRoleParams{
					CounterpartyID: 1,
					Role:           "INVALID",
					CreatedAtMs:    2,
				})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := store.withWriteQueries(ctx, test.name, test.operation)
			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("error = %v, want domain.ErrConflict", err)
			}
			var sqliteErr *modernsqlite.Error
			if !errors.As(err, &sqliteErr) {
				t.Fatalf("mapped error lost modernc SQLite cause: %v", err)
			}
		})
	}
}

func TestClassifyErrorMapsForeignKeyViolation(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "foreign-key.db"), database.DefaultOpenOptions())
	ctx := context.Background()

	err := store.withWriteQueries(ctx, "add counterparty role", func(queries *sqlcgen.Queries) error {
		return queries.InsertCounterpartyRole(ctx, sqlcgen.InsertCounterpartyRoleParams{
			CounterpartyID: 999,
			Role:           "SUPPLIER",
			CreatedAtMs:    1,
		})
	})
	if !errors.Is(err, domain.ErrInvalidReference) {
		t.Fatalf("error = %v, want domain.ErrInvalidReference", err)
	}
	var sqliteErr *modernsqlite.Error
	if !errors.As(err, &sqliteErr) {
		t.Fatalf("mapped error lost modernc SQLite cause: %v", err)
	}
}

func TestClassifyErrorMapsExternalBusy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "busy.db")
	options := database.DefaultOpenOptions()
	options.BusyTimeout = 25 * time.Millisecond
	first := newAdapterTestStore(t, path, options)
	second := newAdapterTestStore(t, path, options)
	ctx := context.Background()

	entered := make(chan struct{})
	release := make(chan struct{})
	firstResult := make(chan error, 1)
	go func() {
		firstResult <- first.withWriteQueries(ctx, "hold write", func(_ *sqlcgen.Queries) error {
			close(entered)
			<-release
			return nil
		})
	}()
	<-entered

	err := second.withWriteQueries(ctx, "contending write", func(_ *sqlcgen.Queries) error {
		return nil
	})
	close(release)
	if firstErr := <-firstResult; firstErr != nil {
		t.Fatalf("holding write: %v", firstErr)
	}

	if !errors.Is(err, domain.ErrBusy) {
		t.Fatalf("error = %v, want domain.ErrBusy", err)
	}
	var sqliteErr *modernsqlite.Error
	if !errors.As(err, &sqliteErr) {
		t.Fatalf("mapped error lost modernc SQLite cause: %v", err)
	}
}

func TestCorruptDataErrorPreservesMappingCause(t *testing.T) {
	cause := errors.New("invalid persisted enum")
	err := corruptDataError("map item", cause)
	if !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("error = %v, want domain.ErrCorruptData", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("corrupt-data error lost mapping cause: %v", err)
	}
}

func newAdapterTestStore(t *testing.T, path string, options database.OpenOptions) *Store {
	t.Helper()
	db, err := database.NewDatabaseWithOptions(path, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	return NewStore(db)
}

func testItemParams(name, normalizedName string) sqlcgen.InsertItemParams {
	return sqlcgen.InsertItemParams{
		Name:           name,
		NormalizedName: normalizedName,
		BaseUnitCode:   "g",
		IsPurchasable:  1,
		CreatedAtMs:    1,
		UpdatedAtMs:    1,
	}
}
