# ADR 0008: Settings, time, identifiers, migrations, and restore

- Status: Accepted
- Date: 2026-07-13

## Context

Locale and BRL formatting are hard-coded, business timezone is implicit, and
the current settings page contains fake planning fields. Migration records do
not have checksums and restore can replace a live database without a complete
compatibility contract.

## Decision

The database has strongly typed singleton settings rather than arbitrary
key/value strings. Initial defaults are locale `pt-BR`, currency `BRL`, and
timezone `America/Sao_Paulo`, but first-run setup may change them before data is
posted.

Currency and its minor digits become immutable after the first stock document,
because every inventory valuation is denominated in that currency even when a
particular line has zero value. Multi-currency and exchange rates are excluded.
Locale and timezone remain mutable display/report settings. Planning defaults
may include hourly labor cost and default margin in integer basis points, but
they do not automatically alter inventory value.

Posted instants are stored as UTC Unix milliseconds and cross Wails as RFC3339.
Business and expiry dates are stored as ISO `YYYY-MM-DD`. Report day boundaries
are calculated in Go using the persisted IANA timezone. Business dates may be
backdated but cannot reorder the ledger.

Local integer primary keys are canonical inside the single database. Posted
commands additionally use a globally unique text idempotency key. V2 does not
pay the complexity cost of UUIDs for every local row.

Because the application is pre-production, the experimental migrations will be
replaced by one reviewed baseline after preserving the discovered development
database. From that baseline onward migrations are embedded, ordered,
transactional, forward-only, and checksummed. An applied migration is never
edited.

Restore validates application identity, supported schema version, migration
checksums, SQLite integrity, foreign keys, and ledger/projection reconciliation
before activation. The application makes a safety backup, atomically replaces
the database, and restarts so no repository retains an obsolete connection.

## Consequences

- Time and expiry semantics do not depend on a Windows machine's current locale.
- Reports can change timezone presentation without rewriting posting history.
- Fake monthly-profit/expense settings remain absent until a real use case gives
  them semantics.
- Migration fixtures and backup/restore tests are mandatory before schema V1 is
  considered stable.
- Import of a newer unsupported schema is rejected rather than guessed.
