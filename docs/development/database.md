# Database development

Sweeters V2 uses one local SQLite file and embedded, forward-only migrations.
The database is the first executable layer in the bottom-up architecture: a
store, service, Wails handler, or page is valid only after it conforms to this
contract.

See [ADR 0009](../decisions/0009-v2-sqlite-baseline-and-enforcement.md) for the
baseline decision, [ADR 0011](../decisions/0011-recipe-output-and-archive-version-integrity.md)
for the first hardening migration, and [the data model](../architecture/data-model.md)
for table responsibilities.

## Local files

The development script sets the data directory to:

```text
%APPDATA%\saas-dev
```

and the database is `%APPDATA%\saas-dev\app.db`. Set `SAAS_DATA_DIR` before
running `scripts/dev.ps1` to isolate a different fixture or manual test. A
normal packaged run uses the platform user configuration directory unless that
environment variable is set.

Database files, WAL sidecars, exports, and preserved development copies match
the repository's ignore rules and must not be committed. The pre-baseline file
found during Phase 3 contained no business rows and has been removed from the
repository; it is not a supported import source.

## Startup contract

On open, the database package:

1. loads embedded files whose names exactly match contiguous
   `NNNN_name.sql` versions;
2. opens SQLite with connection-local foreign-key enforcement, bounded busy
   waiting, untrusted-schema protection, and `NORMAL` synchronization encoded
   in the modernc DSN;
3. accepts either a truly empty SQLite database or one with the recognized
   Sweeters identity and migration table;
4. validates `application_id`, the exact migration-table DDL, `user_version`,
   and the complete recorded migration prefix, including each filename and
   SHA-256 checksum;
5. applies missing migrations in order with an immediate transaction;
6. enables persistent WAL mode only after database identity and migrations
   have been accepted; and
7. runs SQLite quick integrity and foreign-key checks before returning the
   connection.

The pool remains limited to one physical connection, but correctness does not
depend on that first connection living forever. The DSN reapplies
`foreign_keys = ON`, the configured `busy_timeout`, `trusted_schema = OFF`, and
`synchronous = NORMAL` whenever `database/sql` replaces it. WAL mode remains a
file-persistent startup step so opening an unidentified or unsupported file
does not change its journal mode.

The identity is `application_id = 1398228308` (`0x53574554`, `SWET`). The
baseline is migration version 1 and the current schema is `user_version = 2`.
Migration metadata stores `version`, `name`, `checksum`, and
`applied_at_unix_ms` in a `STRICT` table.

Migration 2 validates existing recipe chains, active recipe outputs, and
archive/version timestamps before recording its version. It fails atomically
rather than guessing how to repair an invalid version-one file.

Startup fails closed when a file belongs to another application, contains
unrecognized legacy objects, has missing/reordered/renamed migration history,
has a checksum mismatch, is newer than the executable, or fails integrity.
Opening does not infer a schema from table names and does not stamp an existing
file as V2. A version-zero database may contain only the exact bootstrap
migration table; unrelated objects beside an empty history are rejected.

The seven old `001_` through `007_` files are intentionally incompatible. If
meaningful data is later found in that model, preserve the original file and
write a separately reviewed one-time extraction/import tool. Do not weaken the
normal startup checks or add compatibility columns to the baseline.

## Writing a migration

`0001_v2_baseline.sql` is immutable after it has been shared. A later schema
change follows this sequence:

1. accept or update the ADR and domain invariant;
2. add the next contiguous migration, for example
   `0003_add_supplier_reference.sql`;
3. use strict integer/text/blob representations, explicit checks, and foreign
   keys consistent with the accepted model;
4. update real-SQLite migration and schema tests;
5. update stores and every higher layer only after the migration passes; and
6. run the complete desktop gate.

Do not rename, reorder, reformat, or edit an applied file: its checksum is over
the exact embedded bytes. There are no down migrations. Correct a released
mistake with the next forward migration. The bootstrap `schema_migrations` DDL
is also persisted identity and must not be reformatted or edited casually.

Keep structural protection in SQLite. Put multi-row algorithms in an explicit
Go transaction rather than in migration triggers. In particular, valuation,
complete document/recipe construction, full lot allocation, FEFO selection,
projection updates, and replay reconciliation belong to application commands.

## Test loop

From the repository root:

```powershell
Push-Location app
go test -shuffle=on -count=1 ./database ./internal/infrastructure/sqlite
Pop-Location
```

Tests create real temporary SQLite files. A schema change is incomplete until
tests cover fresh application, reopen, transactional failure, history and
checksum rejection, strict storage, foreign keys, its new constraints, and
affected aggregate mappings. Use the full `scripts/check-desktop.ps1` gate
before committing.

## Export, reset, and restore

Export uses SQLite `VACUUM INTO` to create a consistent standalone snapshot.
Choose a new destination; an export does not replace or re-identify the active
database.

For a disposable development reset, stop the application first and move the
entire configured data directory to a timestamped backup location. Starting
again creates a fresh V2 file. Keeping the directory together avoids separating
an SQLite file from a possible `-wal` or `-shm` sidecar. Never reset a file whose
contents have not been inspected or backed up.

Restore/import is intentionally disabled in Phase 3. Do not copy bytes over
`app.db` while the process is running and do not reopen individual service
connections afterward. That can mix repositories connected to different
database generations.

Restore may return only when one restart-based workflow can:

1. copy the candidate to a staging path without touching the live file;
2. validate application identity, supported schema version, exact migration
   checksums, SQLite integrity, foreign keys, and ledger/projection replay;
3. create and verify an automatic safety backup of the active database;
4. close all database users and atomically replace the live file; and
5. restart the process before constructing any store or service.

Until those steps and their failure-path tests exist, export is the only
supported database lifecycle operation.

## Disposable demo database

`scripts/demo-data.ps1` creates a separate, marker-protected database for manual
UI testing. It posts fixtures through the application/Wails boundaries instead
of inserting business rows directly. The default path is the ignored
`tmp/demo-data` directory and can be launched with the same `SAAS_DATA_DIR`
override used by normal development.

The script's `Clean -Force` action removes the complete fixture directory after
the application is closed. It requires the exact demo marker and refuses the
repository root, filesystem roots, `%APPDATA%\saas-dev`, and the packaged
default data directory. It is not a reset or deletion mechanism for a real
Sweeters database.
