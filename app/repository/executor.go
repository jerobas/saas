package repository

import (
	"context"
	"database/sql"
)

// Executor is implemented by both *sql.DB and *sql.Tx. It lets application
// services compose several repository operations in one transaction.
type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
