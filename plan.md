My recommendation is to continue in this repository and keep the stack, but treat the next stage as a controlled V2 rebuild of the desktop core. We should not continue adding services on top of the current incompatible contracts.

No files were changed while preparing this plan.

## Scope

The active product becomes a completely local desktop application:

- The former `api/` experiment is deleted from the active repository.
- No HTTP client, authentication, licensing, payments, remote synchronization, or cloud deployment.
- Desktop development, testing, building, and releases must not require network services.
- Existing remote-dependent frontend functionality will be removed from the active desktop build, not mocked.
- Desktop CI and releases will be independent from every non-desktop workflow.

## Authority order

This will be the development contract:

```text
Product invariants and ADRs
        ↓
SQLite schema and immutable migrations
        ↓
Generated SQL queries and persistence types
        ↓
Aggregate repositories
        ↓
Application use cases
        ↓
Wails DTOs and handlers
        ↓
React features
```

The existing database is not automatically considered perfect. We correct its design first. Once approved, it becomes the executable authority: repository or service incompatibilities are fixed upward, not by weakening the database.

A later database change requires an ADR, migration, database test, and then corresponding changes through every higher layer.

## Architectural decisions

I recommend locking in these decisions before implementation:

- Keep the local-first Wails modular monolith with Go, SQLite, React, Vite, and Tailwind.
- Keep one unified item catalog. “Ingredients” and “Products” are filtered views based on item capabilities.
- Replace generic `events`, duplicated purchase/sale lines, and independently created movements with:

  - `stock_documents`
  - `stock_document_lines`
  - `inventory_lots`
  - `lot_allocations`
  - rebuildable `inventory_balances`

- A document line is the canonical ledger entry. Commercial totals and inventory valuation remain separate fields.
- Use exact fixed-point integers for quantities and currency minor units. Do not use SQLite `REAL` for business arithmetic.
- Lots are first-class. Outbound stock can consume multiple lots, preferably using FEFO for physical allocation.
- Use weighted-average inventory valuation initially, with the outbound cost stamped when posting.
- Posted documents are immutable. Corrections use reversal documents.
- Keep drafts in frontend state initially. Persist drafts only when autosave or multi-session workflows justify them.
- Version recipes using immutable recipe revisions. Production records exactly which revision was used.
- Production explicitly consumes input lots and creates an output lot. It never silently creates catalog items.
- Replace the current generic conversion workflow with units and packaging conversions. Physical transformation belongs to production.
- Archive referenced items, recipes, units, and counterparties instead of deleting them.
- Keep structural invariants in SQLite: foreign keys, checks, uniqueness, immutability guards, and indexes.
- Put posting, costing, allocation, and projection updates in one explicit Go transaction. Avoid complex business logic hidden in triggers.
- Adopt `sqlc` for schema-first, checked Go queries.
- Use aggregate-oriented repositories, not one CRUD repository per table.
- Isolate Wails-generated code behind typed handlers and a frontend desktop gateway.
- Convert the frontend to strict TypeScript and feature-based folders.
- Remove fake dashboard, sales, product, and batch behavior until the corresponding real query model exists.

## Execution plan

### Phase 0 — Safe starting point

- Create a dedicated V2/refactor branch and retain the existing backup reference.
- Back up any local development database.
- Verify the desktop starts, builds, and executes its minimal checks offline.
- Detach desktop CI/release behavior from unrelated workflows.
- Remove remote-dependent modules from the active desktop dependency graph.
- Establish one local command for all fast checks.

Exit condition: a clean desktop baseline with no network requirement.

### Phase 1 — Product and domain contract

Write short ADRs covering:

- Quantity precision and rounding.
- Money and currency policy.
- Base units and packaging conversions.
- Inventory valuation and FEFO lot allocation.
- Negative-stock policy.
- Document lifecycle and reversal rules.
- Recipe revision policy.
- Archive policy.
- Application timezone and timestamps.
- Backup, restore, and schema compatibility.
- Single-user/local-first scope.

Produce the glossary, use-case list, and proposed ERD before writing services.

Exit condition: every important database behavior has a documented, testable invariant.

### Phase 2 — Toolchain and quality foundation

Upgrade packages in isolated groups, keeping the build green after each group.

| Component | Target |
|---|---:|
| Go | 1.26.5; Go has no LTS channel, so use latest supported stable. [Official downloads](https://go.dev/dl/) |
| Wails | 2.13.0; already latest stable. Do not adopt Wails 3 while it remains prerelease. [Wails v2.13](https://github.com/wailsapp/wails/releases/tag/v2.13.0) |
| Node | 24.17.0 LTS, pinned in the repository. [Node release status](https://nodejs.org/en/about/previous-releases) |
| Vite / React plugin | 8.1.4 / 6.0.3. [Vite 8 changes](https://vite.dev/blog/announcing-vite8) |
| React / React DOM | Matching 19.2.7 versions |
| React Router | 7.18.1 using `react-router`, removing the compatibility re-export. [Migration guidance](https://reactrouter.com/upgrading/v7) |
| Tailwind | 4.3.2 for both Tailwind and its Vite plugin. [Tailwind 4.3](https://tailwindcss.com/blog/tailwindcss-v4-3) |
| SQLite driver | `modernc.org/sqlite` 1.53.0. [Driver releases](https://gitlab.com/cznic/sqlite/-/tags) |
| TypeScript | 6.0.3 initially; TS 7 tooling support is not mature enough yet. [TypeScript 7 caveats](https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/) |

Also:

- Add Vitest, Testing Library, Playwright, ESLint, Prettier, and coverage tooling.
- Move `framer-motion` to `motion`.
- Standardize on `@phosphor-icons/react`.
- Remove Axios and unused frontend dependencies.
- Pin the Wails CLI instead of installing `@latest`.
- Add an independent desktop CI workflow.

Exit condition: formatting, linting, Go tests, frontend tests, production frontend build, and Wails build all use pinned versions.

### Phase 3 — Canonical database V1

Because this is pre-production and the seven existing migrations describe an experimental model, I recommend rebaselining them into a clean canonical schema after making a safety backup.

If valuable local data is discovered, we add a one-time converter rather than contaminating the permanent schema with compatibility structures.

Implement:

- Settings, units, item units, items, and counterparties.
- Recipe definitions, immutable revisions, and revision components.
- Stock documents and canonical document lines.
- Lots and many-to-many outbound lot allocations.
- Rebuildable quantity and inventory-value projections.
- Reversal relationships and immutable posted records.
- Archive timestamps, complete foreign-key actions, and required indexes.
- Embedded, checksummed, forward-only migrations.

Exit condition: the schema can be created fresh, reopened, upgraded from fixtures, integrity-checked, and completely validated without any repository or service code.

### Phase 4 — Go foundation

Restructure the desktop code approximately as:

```text
app/
  cmd/desktop/
  internal/
    domain/
    application/
    infrastructure/sqlite/
    presentation/wails/
  frontend/src/features/
```

Implement:

- `Money`, `Quantity`, `Unit`, IDs, statuses, document types, and typed errors.
- Generated SQL queries.
- Aggregate repositories for catalog, recipes, documents, inventory, and reporting.
- A transaction runner that supplies transaction-scoped repositories.
- A central posting engine for validation, costing, allocation, ledger writes, and balance updates.
- Explicit Wails request/response DTOs; no `sql.Null*` or generated persistence types cross that boundary.

Exit condition: repositories and use cases operate against real temporary SQLite databases with no compatibility adapters.

### Phase 5 — Vertical capabilities

Build each capability fully before moving to the next:

1. Settings, units, catalog, and counterparties.
2. Purchases, inbound lots, and inventory balances.
3. Adjustments and reversals.
4. Recipes and recipe revisions.
5. Production consumption, yield, costing, and output lots.
6. Sales, outbound allocation, cost of goods, and stock updates.
7. Reporting and dashboard queries.
8. Validated backup and restore.

Phase 5.7 is a read-model vertical, not another operational store. It should
derive a dashboard from posted documents, document lines, lot allocations, and
`inventory_balances` without writing any business state. Its first useful slice
is a real empty/loading/error dashboard plus period totals for sales, revenue,
COGS, gross margin, inventory value, low stock, and top products. Exact
reversals are correction/audit events: operational dashboard aggregates exclude
documents that have been exactly reversed, and reversal documents themselves are
reserved for audit reporting.

Every slice follows the same order:

```text
Schema/query
→ repository integration test
→ application use-case test
→ Wails contract
→ TypeScript feature
→ component/workflow tests
→ documentation
```

Old mocks, fake data, pages, and adapters are deleted as their real replacements land.

### Phase 6 — Frontend boundary

- Convert incrementally to strict TypeScript.
- Organize by feature instead of technical file type.
- Introduce a desktop gateway as the only consumer of generated Wails bindings.
- Use TanStack Query for Wails-backed state and cache invalidation; keep local form/UI state in React.
- Standardize validation, loading, empty, and error behavior.
- Add route-level code splitting and basic accessibility checks.
- Components never know about SQL, repository structures, or transport-generated types.

### Phase 7 — Durability and release readiness

- Validate candidate backup databases before import.
- Automatically create a safety backup before replacement.
- Use atomic replacement and controlled restart after restore.
- Test projection reconstruction from the immutable ledger.
- Test WAL, foreign-key enforcement, rollback, retries, double posting, and concurrent commands.
- Produce reproducible, tag-triggered desktop builds.
- Run packaged Windows smoke tests; add macOS when it becomes a supported target.

## Testing policy

The current baseline has only five Go tests and no frontend test setup, so tests are part of architecture work, not final cleanup.

Required layers:

- Database constraint and migration tests.
- Ledger-versus-projection property and fuzz tests.
- Fixed-point arithmetic and rounding unit tests.
- Repository integration tests using real temporary SQLite.
- Application tests using real repositories, with a fixed clock.
- Rollback, idempotency, reversal, and insufficient-stock tests.
- Vitest and Testing Library component tests.
- Playwright critical workflows through a deterministic fake desktop gateway.
- Packaged Wails startup and clean-profile smoke tests.

Mocks should be limited to clocks, OS dialogs, notifications, and the frontend Wails gateway—not used between the database, repositories, and application services.

## Documentation deliverables

- Desktop-focused README and quick start.
- Architecture overview and dependency rules.
- ERD and database invariant catalog.
- Domain glossary and inventory-ledger specification.
- ADR collection.
- Migration and data compatibility policy.
- Testing and fixture guide.
- “How to add a feature bottom-up” development guide.
- Backup/recovery manual.
- Release and contribution guides.
- Archived historical proposal/chat documents after extracting valid decisions.

## Final completion criteria

Development is considered caught up when:

- The application develops, tests, builds, and runs entirely offline.
- All accepted database invariants are executable tests.
- Each public use case has success, validation, and rollback coverage.
- Ledger replay always equals the stored inventory projection.
- No legacy compatibility adapters or fake operational data remain.
- A clean machine can build the packaged application using pinned tooling.
- A previous supported database fixture can be upgraded or safely converted.
- Documentation describes the implemented system rather than intended future behavior.
