# Architecture and file placement conventions

Sweeters V2 uses a bottom-up, local-first architecture. New work should keep the
same dependency direction unless an ADR explicitly changes it.

```text
SQLite schema / queries
-> infrastructure adapters
-> domain values and aggregates
-> application use cases
-> presentation handlers and DTOs
-> frontend feature pages
```

Higher layers conform to lower layers. UI code never bypasses application
services, application code never inspects SQLite driver details, and domain code
never imports Wails, SQL, generated rows, or frontend concepts.

## Go backend

### `internal/domain`

Use for business vocabulary, value objects, invariants, aggregate snapshots, and
stable domain errors.

- Put broad primitives in the package root only when they are truly shared:
  money, quantities, timestamps, IDs, text normalization, options, and errors.
- Put vertical aggregates in subpackages when they have their own invariants:
  `catalog`, `counterparty`, `inventory`, `recipe`, etc.
- Do not add generic `utils` or `libs` packages. Prefer a precise domain name
  or keep the helper private to the package that uses it.

### `internal/application`

Use for use cases, application services, ports, clocks, and adapters that turn
infrastructure stores into application ports.

- Name files by vertical capability: `purchase_service.go`,
  `reporting_service.go`, `*_sqlite_adapter.go`.
- Keep command validation and transaction orchestration decisions here when
  they are use-case decisions rather than pure domain invariants.
- Keep application tests beside the implementation.

### `internal/infrastructure/sqlite`

Use for the SQLite-backed store facade, SQL row mapping, generated-query calls,
transactions, and database error mapping.

- Keep one SQLite package for now. A Go subfolder is a new package, not just a
  visual grouping, so split only when the package boundary is worth the import
  and transaction-composition cost.
- Keep store implementation and tests together:
  `adjustment_store.go` with `adjustment_store_test.go`,
  `purchase_store.go` with `purchase_store_test.go`, and so on.
- Name shared files by responsibility, not by vagueness:
  `errors.go`, `store.go`, `*_mapping.go`, `queries/`, `sqlcgen/`.
- Edit SQL in `queries/`; commit generated code in `sqlcgen/`.

### `internal/presentation/wails`

Use for Wails handlers, request/response DTO mapping, Wails runtime glue, and
desktop presentation services.

- Handlers translate DTOs to application inputs and application outputs to DTOs.
- Keep DTOs in `internal/presentation/wails/dto`.
- Wails-specific platform glue such as save dialogs and runtime events belongs
  here, not in a root `service` package.

### Root `app`

Keep root `app` focused on composition:

- Wails startup;
- database initialization;
- dependency wiring;
- build/runtime mode switches.

Avoid putting business, infrastructure, or feature implementation in the root
package.

## Frontend

Feature pages live under `app/frontend/src/features`.

```text
features/catalog/
features/dashboard/
features/database/
features/inventory/
features/production/
features/purchases/
features/recipes/
features/sales/
features/settings/
```

- Keep feature page tests beside their pages.
- Shared layout/components stay in `components/`.
- Shared runtime context stays in `context/`.
- The current desktop bridge stays in `gateways/desktopBridge.ts` until we
  explicitly split it by feature.
- New visual work should prefer TypeScript files. Existing JSX can migrate
  vertically when that feature is already being touched.

## Documentation

- Product/architecture changes go through ADRs in `docs/decisions`.
- Business vocabulary and invariants live in `docs/domain`.
- Development workflow and placement conventions live in `docs/development`.
- Update docs in the same slice as the code when a rule changes.

## Tests

- Domain tests live beside domain packages.
- Application tests live beside application services.
- SQLite store tests live beside SQLite store implementations.
- Wails handler/surface tests live beside Wails handlers.
- Frontend component tests live beside the feature page or component.
- Browser smoke tests live in `app/frontend/e2e`.

Tests should describe observable behavior at the same boundary as the code they
exercise. Avoid a global test directory unless a future test suite genuinely
crosses many boundaries.
