# Architecture overview

## Product boundary

Sweeters is a local-first desktop application for a single confectionery or
small food-production business operating one stock location.

The supported product areas are:

- business settings and units;
- a unified stock-item catalog;
- suppliers and customers;
- purchases and inbound lots;
- recipes and immutable recipe revisions;
- production and output lots;
- sales and cost of goods sold;
- stock adjustments, expiry, and corrections;
- inventory, lot, and profitability queries;
- local backup and restore.

V2 is operational inventory software. It is not fiscal or general-ledger
accounting software. Multi-location stock, multi-user collaboration,
multi-currency transactions, tax documents, partial returns, automatic demand
planning, and remote synchronization are not part of the initial model.

## Runtime shape

The application remains a modular monolith:

```text
React feature
    -> typed desktop gateway
        -> Wails handler
            -> application command/query
                -> aggregate store
                    -> generated SQLite query
                        -> embedded migration contract
```

SQLite is the only persistence engine. A stock-posting command owns one
database transaction containing the document, ledger lines, lots, lot
allocations, and balance projections. It either commits completely or leaves no
record.

## Source layout and next layers

```text
app/
  database/                    # lifecycle, migration, BEGIN IMMEDIATE
  internal/
    domain/
      catalog/
      counterparty/
      inventory/
      recipe/
      settings/
    infrastructure/sqlite/
      queries/                 # named SQL source
      sqlcgen/                 # committed generated code
      *_store.go               # domain mapping and aggregate operations
    application/               # Phase 5 commands and queries
      commands/
      queries/
    presentation/wails/        # later task-oriented desktop handlers
  service/                     # current OS/database lifecycle handlers
  frontend/src/
    gateways/
    features/
      settings/
      catalog/
      counterparties/
      purchases/
      inventory/
      recipes/
      production/
      sales/
      backup/
```

The comments marked Phase 5 or later are planned boundaries, not empty
compatibility packages. The exact split may evolve, but dependencies may only
point downward. In particular, domain packages import neither SQLite nor Wails,
and generated query rows do not leave the SQLite adapter.

## Layer responsibilities

### SQLite migrations

Define types, keys, checks, relationships, indexes, immutability guards, and the
representable states of the product. Applied migrations are embedded,
checksummed, ordered, and immutable.

### Generated queries and stores

Own SQL and persistence mapping. Stores are aggregate-oriented rather than one
generic CRUD repository per table. Named queries are generated reproducibly by
the pinned sqlc tool. No persistence type is exposed to Wails.

### Domain and application

Use typed identifiers, quantities, money, document kinds, and errors.
Application commands validate cross-row rules and orchestrate transactions.
Complex costing and allocation algorithms live here and are verified against
the SQLite contract with integration and property tests.

### Wails presentation boundary

Exposes task-oriented handlers and explicit request/response DTOs. Time values
are RFC3339 strings, date-only values are ISO dates, and persistence null types
never cross the boundary. Exact `int64` quantities, money, valuation, and
conversion values cross as base-10 strings to avoid JavaScript precision loss.

### React frontend

Calls only the typed desktop gateway. Components do not import generated Wails
bindings directly and do not coordinate multi-step posting workflows.

## Write and read models

The immutable document ledger is the historical write model. Inventory
balances and lot availability are rebuildable projections. Query-specific
views may be introduced for the dashboard and reports, but they never become a
second source of historical truth.

## Error and concurrency policy

- Business commands return stable typed errors suitable for user-facing
  messages.
- SQLite constraint errors are mapped at the infrastructure boundary.
- Mutable master data uses an expected `updated_at` snapshot; stale writes are
  rejected instead of silently overwriting a newer edit.
- Stock posting uses a serialized write transaction so two commands cannot
  consume the same final stock.
- Every posting command carries a unique idempotency key. Retrying the same
  command returns the original result rather than posting twice.
