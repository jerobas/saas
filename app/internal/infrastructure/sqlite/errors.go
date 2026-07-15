package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	modernsqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	errNilReadCallback  = errors.New("sqlite read callback is nil")
	errNilWriteCallback = errors.New("sqlite write callback is nil")
)

// classifyError translates persistence failures into stable domain categories
// while retaining the original error for errors.Is/errors.As and diagnostics.
func classifyError(operation string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return wrapClassifiedError(operation, domain.ErrNotFound, err)
	}

	var sqliteErr *modernsqlite.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() {
		case sqlite3.SQLITE_CONSTRAINT_UNIQUE,
			sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY,
			sqlite3.SQLITE_CONSTRAINT_ROWID:
			return wrapClassifiedError(operation, domain.ErrConflict, err)
		case sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY:
			return wrapClassifiedError(operation, domain.ErrInvalidReference, err)
		}

		switch sqliteErr.Code() & 0xff {
		case sqlite3.SQLITE_BUSY, sqlite3.SQLITE_LOCKED:
			return wrapClassifiedError(operation, domain.ErrBusy, err)
		case sqlite3.SQLITE_CONSTRAINT:
			// Domain input is validated before SQL. A remaining CHECK,
			// NOT NULL, or trigger constraint therefore represents a
			// persisted-state or concurrent business conflict, never an
			// adapter-specific error string for callers to parse.
			return wrapClassifiedError(operation, domain.ErrConflict, err)
		}
	}

	return wrapOperation(operation, err)
}

// corruptDataError marks a successful SQL read that cannot be mapped into a
// valid domain value. The mapping cause remains inspectable by callers.
func corruptDataError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return wrapClassifiedError(operation, domain.ErrCorruptData, err)
}

func wrapClassifiedError(operation string, classification, cause error) error {
	if operation == "" {
		return fmt.Errorf("%w: %w", classification, cause)
	}
	return fmt.Errorf("%s: %w: %w", operation, classification, cause)
}

func wrapOperation(operation string, err error) error {
	if operation == "" {
		return err
	}
	return fmt.Errorf("%s: %w", operation, err)
}
