# Architectural decision records

An ADR records a decision that changes the product model, database contract, or
dependency rules. Accepted ADRs are authoritative until replaced by a later ADR.

| ADR | Status | Decision |
|---|---|---|
| [0001](0001-local-desktop-and-bottom-up-authority.md) | Accepted | Local desktop boundary and bottom-up authority |
| [0002](0002-exact-quantities-money-and-units.md) | Accepted | Exact quantities, commercial money, valuation, and units |
| [0003](0003-document-ledger-valuation-and-projections.md) | Accepted | One document ledger, weighted average, and projections |
| [0004](0004-lots-fefo-and-expiry.md) | Accepted | First-class lots, FEFO, and expiry policy |
| [0005](0005-posting-lifecycle-and-reversals.md) | Accepted | Posted-only lifecycle, idempotency, and exact reversals |
| [0006](0006-catalog-counterparties-and-archive.md) | Accepted | Unified catalog, counterparties, and archival |
| [0007](0007-recipe-revisions-and-production.md) | Accepted | Immutable recipe revisions and production value transfer |
| [0008](0008-settings-time-and-data-lifecycle.md) | Accepted | Settings, time, identifiers, migrations, and restore |

## Lifecycle

- `Proposed`: under discussion and not authoritative.
- `Accepted`: must be reflected by new implementation work.
- `Superseded`: retained for history and linked to its replacement.
- `Rejected`: considered but not selected.

An ADR change is made before its schema migration, tests, repository changes,
application behavior, Wails contract, or UI.
