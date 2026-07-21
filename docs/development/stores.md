# Domain and store development

Phase 4 implements the first Go layer above the V2 SQLite contract. It contains
strong domain values, aggregate snapshots, generated named queries, and
SQLite-backed stores. It deliberately does not contain application commands or
Wails DTOs.

## Dependency boundary

```text
application (Phase 5)
    -> domain aggregates and errors
    -> SQLite store facade
        -> private domain/row mapping
        -> generated sqlc query package
            -> database lifecycle and BEGIN IMMEDIATE coordinator
                -> embedded migrations
```

Only the adapter imports the generated query package. Domain packages do not
import `database/sql`, SQLite, JSON, Wails, or generated rows. Generated rows
must be converted while still inside the adapter.

## Changing SQL

Named queries live in `app/internal/infrastructure/sqlite/queries`. Generated
Go lives in the sibling `sqlcgen` directory and is committed so a build does
not require the generator.

From `app`:

```powershell
go tool sqlc compile
go tool sqlc generate
go tool sqlc diff
```

Edit the SQL source, not generated `.go` files. The desktop check runs
`compile` and `diff`, so invalid SQL and stale output fail before Go tests.
Changing a table still requires a forward migration and an ADR first; sqlc is
not a migration tool.

## Transactions

Single-statement reads use the database-owned context query surface. Every
multi-query aggregate read owns one guarded read transaction on a dedicated
connection and binds its private generated queries to that single SQLite
snapshot. The connection is query-only for the callback and is restored before
it returns to the pool.

Each public multi-row mutation owns one complete `BEGIN IMMEDIATE` transaction
and binds its private generated queries to that dedicated connection. There is
no public transaction-composition facade. Application code calls a complete
aggregate operation; it must not start `Database.Write` and then call a root
store, or nest another write, because that would wait on the single SQLite
connection.

The callback and its values must not escape. An error or panic rolls back, and
cancellation is checked again before commit. The coordinator deliberately does
not retry; an application command decides whether a whole operation is safe to
retry.

## Store rules

- Load and save complete aggregate boundaries, not arbitrary table rows.
- Validate input with domain constructors before opening a write transaction.
- Use the caller's expected `updated_at` value for mutable master data. Zero
  affected rows become either not-found, archived/state conflict, or stale
  data after a same-transaction existence check.
- Advance `updated_at` monotonically, even when two actions occur in the same
  wall-clock millisecond.
- Archive master data; never expose hard deletion of domain or history rows.
  The archive instant is the same value as the `updated_at` version advanced by
  that mutation.
- Return non-nil empty slices.
- Wrap invalid persisted values as corrupt-data errors rather than returning a
  partially valid aggregate.
- Validate a recipe's complete contiguous revision chain on aggregate,
  individual-revision, list, and publish reads before returning or mutating it.
- Map SQLite constraints, foreign references, busy/locked state, and missing
  rows to stable domain errors. Higher layers must not inspect driver strings.

Settings, catalog, counterparties, and recipe definitions may be changed in
this layer. Inventory stores are queries only. No query or store in Phase 4 may
insert ledger facts, allocate lots, or update balances; those operations need
the complete posting transaction introduced in Phase 5.

The lot-fact, FEFO-candidate, line-allocation, and complete recipe-revision
readers are intentionally internal Phase 4 fact readers, not presentation
pagination contracts. Phase 5 must add quantity-bounded FEFO selection and
keyset-bounded history queries before any of these histories are exposed to a
user-facing route or allowed to grow without a practical bound.

## Tests

Pure domain behavior belongs beside the domain package, including property and
fuzz coverage for normalization, checked arithmetic, and exact conversions.
Store tests use a fresh file-backed database under `t.TempDir()` and the real
embedded migration. They should cover round trips, aggregate atomicity,
archival, optimistic conflicts, keyset order, constraint-to-domain error
mapping, corrupt-row handling where representable, and concurrent write
serialization.

Counterparty names have no normalized-key column in the baseline. Until a
forward migration adds one, counterparty filtering is display-text-sensitive;
do not use SQLite `NOCASE` or `lower()` as a Unicode approximation.
