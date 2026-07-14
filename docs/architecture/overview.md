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

## Target source layout

```text
app/
  cmd/desktop/
  internal/
    domain/
      catalog/
      inventory/
      recipes/
    application/
      commands/
      queries/
    infrastructure/sqlite/
      migrations/
      queries/
      stores/
      backup/
    presentation/wails/
  frontend/src/
    desktop/
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

The exact package split may evolve, but dependencies may only point downward.

## Layer responsibilities

### SQLite migrations

Define types, keys, checks, relationships, indexes, immutability guards, and the
representable states of the product. Applied migrations are embedded,
checksummed, ordered, and immutable.

### Generated queries and stores

Own SQL and persistence mapping. Stores are aggregate-oriented rather than one
generic CRUD repository per table. No persistence type is exposed to Wails.

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
- Stock posting uses a serialized write transaction so two commands cannot
  consume the same final stock.
- Every posting command carries a unique idempotency key. Retrying the same
  command returns the original result rather than posting twice.
