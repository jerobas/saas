# ADR 0009: V2 SQLite baseline and enforcement boundary

- Status: Accepted
- Date: 2026-07-14

## Context

ADRs 0001 through 0008 define the V2 domain, but the repository still contains
seven experimental migrations based on floating-point quantities, duplicated
commercial and inventory records, mutable workflow rows, and trigger-owned
posting. Preserving those tables would make the old implementation, rather
than the accepted bottom-up contract, authoritative.

The application is pre-production. The discovered development database has no
business rows and has been preserved outside version control. This is the last
safe point to establish an unambiguous SQLite identity and migration history
without maintaining a compatibility model that the product has rejected.

## Decision

### Baseline and identity

The seven experimental migrations are replaced by one reviewed migration,
`0001_v2_baseline.sql`. This is a pre-production reset, not an in-place data
upgrade. A database containing legacy user objects is rejected; it is never
silently adopted by `CREATE TABLE IF NOT EXISTS` statements.

Every Sweeters V2 database has:

- SQLite `application_id` `1398228308` (`0x53574554`, ASCII `SWET`);
- `user_version` equal to the highest applied migration, initially `1`;
- a `STRICT` `schema_migrations` table containing the integer version, exact
  filename, SHA-256 checksum of the embedded bytes, and UTC application time in
  Unix milliseconds.

Migration filenames are contiguous `NNNN_name.sql` values. Migrations are
embedded, applied in order inside `BEGIN IMMEDIATE` transactions, forward-only,
and immutable after release. Startup validates the exact applied prefix,
including names and checksums. A foreign application, unrecognized legacy
database, missing or reordered history, modified migration, or schema newer
than the executable is rejected rather than repaired or guessed.

All domain tables are SQLite `STRICT` tables with foreign keys enabled. Domain
entities use local integer primary keys. Controlled measurement units use their
immutable code, counterparty roles use their natural composite key, and
one-to-one detail/projection rows reuse the owning row's integer key. A stock
document additionally stores a unique positive `posting_sequence`. The
application assigns it strictly above the current maximum inside the serialized
posting transaction; gaps are valid and values are never reused.

### Storage representation

Business quantities use signed 64-bit integers in atomic units, commercial
money uses integer minor units, and inventory valuation and estimated/direct
production cost use integer microcurrency units. Floating-point business
columns are forbidden. Storage names make scales visible with suffixes such as
`_atomic`, `_minor`, and `_micro`.

UTC instants use integer Unix milliseconds with `_at_ms`; business and expiry
dates use ISO `YYYY-MM-DD` text with `_on`. Unit and packaging conversions are
positive rational numerator/denominator pairs directly to an item's atomic
quantity. Mass, volume, and count units are seeded; `g`, `ml`, and `each` are
the item base units.

SQLite's built-in `NOCASE` collation is ASCII-only and is not the V2 identity
rule. Values with case-insensitive uniqueness store a separate key produced by
the Go application using Unicode whitespace trim, NFC normalization, full
Unicode case folding, and a final NFC normalization, in that order. SQLite
enforces uniqueness on that key, including archived rows. This applies to item
and recipe names, item-specific packaging names, and optional item SKUs; display
values remain separate.

Document reasons are closed values: purchase permits `FREE_STOCK`; sale
permits `PROMOTION` or `SAMPLE`; adjustment requires `OPENING_BALANCE`,
`FREE_STOCK`, `PHYSICAL_COUNT`, `WASTE`, `EXPIRY`, `DAMAGE`, `SAMPLE`, or
`DOCUMENTED_CORRECTION`; production has no reason; and reversal requires
`EXACT_REVERSAL`.

The baseline implements the accepted unified item catalog, recipe revisions,
one immutable posted document ledger, lots and allocation effects, and the
rebuildable inventory balance projection. It does not provide legacy views,
columns, or compatibility triggers.

### Enforcement boundary

SQLite owns structural truth: strict types, checks, foreign keys, uniqueness,
row-shape constraints, immutable-history guards, valid cross-row references,
and nonnegative projection limits. Those rules protect the file even if a
future caller bypasses a repository.

The Go posting transaction owns algorithms and complete aggregate validity:
checked arithmetic, normalization and reduced/exact conversions, document and
revision completeness, idempotent retry behavior, weighted-average valuation,
line-value reconciliation, full lot allocation, FEFO and expiry eligibility,
projection updates, and replay reconciliation. No general-purpose CRUD API may
mutate ledger history or inventory balances.

Legacy repositories and services are not evidence that the V2 baseline is
supported. They remain outside the bound desktop surface until their aggregate
stores and use cases are rebuilt against this schema in bottom-up order.

### Backup and restore

The ignored development database is preserved as
`app/database/app.pre-v2-20260714.db` before the reset. It is a diagnostic
backup, not a supported V2 import fixture.

Consistent export may remain available. Restore/import is disabled during this
phase because replacing the file beneath live `*sql.DB` users can leave closed
or obsolete connections in services. Restore may be re-enabled only as a
restart-based workflow that stages and validates application identity, schema
support, migration checksums, SQLite integrity, foreign keys, and ledger versus
projection reconciliation; creates a safety backup; closes the active
database; atomically replaces it; and restarts before repositories reopen.

## Consequences

- Phase 3 makes the migration the executable lower-layer authority; Phase 4
  stores must conform to it instead of preserving experimental interfaces.
- An old development database does not open automatically. Recovery is an
  explicit migration project if meaningful legacy data is later discovered.
- Editing `0001_v2_baseline.sql` after it has been shared causes a checksum
  failure; every subsequent change requires a new migration.
- SQLite tests use real temporary files and cover identity, checksums,
  transactional rollback, strict storage, foreign keys, immutability, and
  representative ledger/lot constraints.
- The database alone does not make a posting valid. Application transaction and
  replay tests remain required before operational workflows are enabled.
