package sqlite

import (
	"context"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

// Store is the SQLite adapter root. Generated query and row types remain
// private to this package; aggregate stores added alongside it map those rows
// to domain values before returning them.
type Store struct {
	database *database.Database
	queries  *sqlcgen.Queries

	// catalogReadHook is a package-private synchronization seam used by
	// deterministic snapshot tests. Production stores leave it nil.
	catalogReadHook func(catalogReadStage) error
	recipeReadHook  func(recipeReadStage) error
}

func NewStore(db *database.Database) *Store {
	if db == nil {
		panic("sqlite store requires a database")
	}
	return &Store{
		database: db,
		queries:  sqlcgen.New(db),
	}
}

// withReadQueries binds generated queries to the read-only snapshot owned by
// Database.Read. Keeping the callback private prevents the transaction or
// generated rows from escaping the adapter boundary.
func (s *Store) withReadQueries(
	ctx context.Context,
	operation string,
	callback func(*sqlcgen.Queries) error,
) error {
	if callback == nil {
		return classifyError(operation, errNilReadCallback)
	}
	err := s.database.Read(ctx, func(tx *database.ReadTx) error {
		return callback(sqlcgen.New(tx))
	})
	return classifyError(operation, err)
}

// withWriteQueries binds generated queries to the dedicated connection owned
// by Database.Write. Keeping this callback private prevents incomplete SQL
// aggregates and generated rows from becoming an application API.
func (s *Store) withWriteQueries(
	ctx context.Context,
	operation string,
	callback func(*sqlcgen.Queries) error,
) error {
	if callback == nil {
		return classifyError(operation, errNilWriteCallback)
	}
	err := s.database.Write(ctx, func(tx *database.WriteTx) error {
		return callback(sqlcgen.New(tx))
	})
	return classifyError(operation, err)
}
