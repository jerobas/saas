# Frontend desktop gateway split plan

`app/frontend/src/gateways/desktopBridge.ts` is intentionally still a single
file, but it has grown past the point where it is pleasant to scan. The next
refactor should split it by feature without changing Wails handler names or JSON
contracts.

## Goals

- Keep one small Wails invocation helper instead of duplicating `window.go`
  lookup logic.
- Move request/response types next to the gateway that owns them.
- Preserve current feature imports during the migration through a temporary
  barrel export.
- Avoid changing UI behavior while splitting files.

## Proposed target structure

```text
src/gateways/
  runtime.ts              # invoke<T>(), Window.go typing, namespace fallback
  sharedTypes.ts          # tiny cross-feature primitives only
  settingsGateway.ts
  referenceDataGateway.ts
  catalogGateway.ts
  counterpartyGateway.ts
  purchaseGateway.ts
  adjustmentGateway.ts
  reversalGateway.ts
  productionGateway.ts
  saleGateway.ts
  recipeGateway.ts
  inventoryGateway.ts
  reportingGateway.ts
  databaseGateway.ts
  desktopBridge.ts        # temporary barrel re-export during migration
```

`desktopBridge.ts` should eventually become either a barrel file or disappear
after all imports move to feature-specific gateway modules.

## Ownership map

| Current export | Target module |
| --- | --- |
| `settingsGateway`, settings DTOs | `settingsGateway.ts` |
| `referenceDataGateway`, measurement unit DTOs | `referenceDataGateway.ts` |
| `catalogGateway`, item/package DTOs, archive/version types | `catalogGateway.ts` |
| `counterpartyGateway`, role/counterparty DTOs | `counterpartyGateway.ts` |
| `purchaseGateway`, purchase DTOs | `purchaseGateway.ts` |
| `adjustmentGateway`, adjustment reason/direction DTOs | `adjustmentGateway.ts` |
| `reversalGateway`, reversal DTOs | `reversalGateway.ts` |
| `productionGateway`, production DTOs | `productionGateway.ts` |
| `saleGateway`, sale DTOs | `saleGateway.ts` |
| `recipeGateway`, recipe DTOs | `recipeGateway.ts` |
| `inventoryGateway`, balance/lot/ledger/allocation DTOs | `inventoryGateway.ts` |
| `reportingGateway`, reporting DTOs | `reportingGateway.ts` |
| `ExportDatabase` | `databaseGateway.ts` as `databaseGateway.exportDatabase()` |

Shared types should stay small. Good candidates:

- `ArchiveFilter`;
- `VersionedRequest`;
- `MeasurementUnitResponse`, if catalog and reference-data both need the same
  shape;
- `StockDirection`, if adjustment/production/sales/reversal continue sharing it.

Do not create a broad `types.ts` dump. If a type is owned by one feature, keep it
in that feature's gateway module and import it explicitly where needed.

## Migration sequence

1. Extract `runtime.ts` with `invoke<T>()` and `Window.go` typing.
2. Extract `sharedTypes.ts` with only the genuinely shared DTO primitives.
3. Move one feature at a time, starting with a small module:
   `settingsGateway.ts` or `databaseGateway.ts`.
4. Keep `desktopBridge.ts` re-exporting moved modules so existing feature pages
   continue compiling.
5. After all gateway modules exist, update feature imports to point at their
   owning gateway module.
6. Remove the barrel only if imports are fully direct and tests remain clear.

## Validation

After each slice:

```powershell
cd app/frontend
npm run typecheck
npm test -- --run src/gateways
npm run build
```

For backend/frontend contract confidence, also keep Wails surface tests green:

```powershell
cd app
go test ./internal/presentation/wails
```

## Non-goals

- Do not rename Wails handler methods during this split.
- Do not change JSON field names during this split.
- Do not move feature page files as part of this split; that already happened.
- Do not introduce a generated-client system unless the manual gateway becomes
  a repeated source of contract bugs.
