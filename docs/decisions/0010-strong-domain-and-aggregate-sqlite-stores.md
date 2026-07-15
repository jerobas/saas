# ADR 0010: Strong domain values and aggregate SQLite stores

- Status: Accepted
- Date: 2026-07-14

## Context

The V2 schema is intentionally incompatible with the experimental table-by-table
models and repositories. Recreating that layer above the new schema would leak
storage representations into business code and would make invalid quantities,
currencies, dates, identifiers, and aggregate states easy to construct.

Phase 4 needs a typed persistence boundary that follows the accepted bottom-up
authority without prematurely defining Wails handlers or the stock-posting
application layer.

## Decision

### Domain values

Domain identifiers are distinct private-field types. Quantities, money,
valuation, percentages, conversion ratios, business dates, UTC instants,
timezones, locales, capabilities, roles, and controlled enum values are created
through validating constructors rather than primitive aliases.

Optional values use an explicit generic option type. Zero values do not
silently mean "missing". Domain collections returned by stores are non-nil,
including when empty.

User-visible names and optional identifying text must be valid UTF-8. The
stored display value is Unicode NFC after trimming surrounding Unicode
whitespace. A uniqueness key is produced from that display value by full
Unicode case folding followed by NFC normalization. SQLite `NOCASE` and
`lower()` are not substitutes because they do not implement this contract.

Positive rational conversion factors are reduced at construction and use
checked integer arithmetic. Applying a factor succeeds only when the result is
an exact atomic quantity; it never rounds. The direct conversion is:

```text
canonical atomic quantity = entered quantity * numerator / denominator
```

Commercial money stores a currency-code and minor-digit snapshot. Creating a
new currency selection validates the current ISO registry, while reading a
persisted historical snapshot validates its stored shape without changing its
digits to match a future registry. Inventory valuation remains a distinct
microcurrency type.

An ISO business date and a UTC instant are different types. IANA timezone names
are validated before persistence. Conversion between them must be explicit at
an application boundary.

### Persistence boundary

SQL is authored as named sqlc queries and generated with the module-pinned sqlc
version. Generated rows and SQLite null types remain private to the adapter.
Stores return domain aggregates and map malformed persisted rows to a corruption
error.

Persistence is aggregate-oriented:

- settings owns the singleton settings aggregate and controlled units;
- catalog owns an item together with its packaging definitions;
- counterparties own identity and the complete role set;
- recipes own their immutable ordered revision history and components; and
- inventory exposes balance, lot-availability, and ledger read models only.

There is no generic table repository, mutable recipe-component repository, or
public inventory projection writer.

Every multi-row write uses one database-owned `BEGIN IMMEDIATE` transaction on
a dedicated connection. A public aggregate mutation owns its complete
transaction; its private adapter callback receives generated queries bound to
that connection. Higher layers do not compose root-store calls inside a
transaction. The Phase 5 posting store will likewise expose complete posting
operations rather than a transaction callback. Callbacks and transaction
handles cannot be retained or nested. Panics and errors roll back; context
cancellation is checked before commit.

The adapter maps missing rows, stale optimistic versions, uniqueness conflicts,
invalid foreign references, busy/locked databases, corrupt persisted data, and
invalid domain input to stable typed errors. Callers do not branch on SQLite
error strings.

### Phase boundary

Phase 4 may mutate master data: settings, items and packaging, counterparties,
and recipe identities/revisions. It may only read inventory balances, lots,
allocations, and ledger history.

Phase 5 owns posting sequence allocation, idempotency orchestration, weighted
average valuation, FEFO selection, lot allocation/restoration, production cost
transfer, inventory projection mutation, reversals, and replay. Phase 4 stores
must not provide shortcuts around that transaction.

Counterparty names are not unique in the V2 schema and have no normalized-key
column. Unicode-insensitive counterparty search is therefore deferred until a
forward migration can support it correctly; the adapter will not pretend that
SQLite ASCII collation satisfies the domain policy.

## Consequences

- Invalid primitive combinations are rejected before SQL execution, while the
  schema remains the final structural authority.
- Aggregate operations have more explicit mapping code than generic CRUD, but
  their transactional and archival rules are visible and testable.
- Generated query changes are reproducible and checked in the desktop gate.
- The experimental model/repository/service compatibility layer is deleted
  instead of being adapted to the incompatible V2 contract.
- Application commands can define task-specific ports in Phase 5 without
  depending on generated SQL types.
