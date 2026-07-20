# Next-session handoff

This document is the shortest reliable path for a new Codex session to continue
Sweeters V2 without losing the current direction.

## Current project posture

- The app is local-first desktop software. Do not depend on any remote API for
  development or runtime.
- The active architecture is bottom-up:

  ```text
  SQLite schema / queries
  -> infrastructure adapters
  -> domain values and aggregates
  -> application use cases
  -> Wails presentation handlers and DTOs
  -> frontend feature pages
  ```

- `todo.md` is now reserved for backup/restore work.
- `UI.md` owns UI/dashboard polish work.
- The API folder was intentionally removed from this branch. Do not restore or
  redesign API code unless the user explicitly reopens that product decision.
- Wails v2 remains the stable target. Do not migrate to Wails v3 just for
  updater support while v3 is still prerelease.

## First files to read in a fresh session

Read these before making structural decisions:

1. `README.md`
2. `docs/README.md`
3. `docs/development/architecture-conventions.md`
4. `docs/development/frontend-gateway-split.md`
5. `UI.md`
6. `todo.md`

For domain-sensitive work, also read:

- `docs/domain/invariants.md`
- `docs/domain/use-cases.md`
- `docs/domain/inventory-ledger.md`
- the relevant ADR in `docs/decisions/`

## Current recommended order

The latest agreed order is:

1. UI work from `UI.md`, with the user being opinionated before/while visual
   decisions are made.
2. Backup/restore from `todo.md`.

The last non-UI cleanup tasks were completed before this handoff:

- reporting monetary field names were standardized;
- reporting reversal policy was documented;
- architecture/file-placement conventions were documented;
- the future `desktopBridge.ts` split was planned but not implemented.

## Working conventions

- Keep changes small and commit per coherent slice.
- Prefer read-only inspection before broad refactors.
- Use `rg`/`rg --files` for code search.
- Keep tests beside implementation counterparts.
- Preserve user changes in a dirty worktree; never reset them away.
- Use `apply_patch` for file edits.

## Frontend conventions

- Feature pages live in `app/frontend/src/features/*`.
- Shared components stay in `components/`.
- Shared runtime context stays in `context/`.
- The desktop bridge is still one file for now:
  `app/frontend/src/gateways/desktopBridge.ts`.
- If splitting the bridge later, follow
  `docs/development/frontend-gateway-split.md` and keep a temporary barrel so
  feature imports do not break all at once.
- New UI decisions should be discussed with the user first when the task affects
  layout, visual priority, or workflow shape.

## Backend conventions

- Domain packages must not import SQL, Wails, generated rows, or frontend
  concepts.
- Application services own use-case orchestration and ports.
- SQLite infrastructure owns generated-query calls, transactions, row mapping,
  and database error mapping.
- Wails presentation owns handlers, DTO mapping, save dialogs, runtime events,
  and desktop-specific presentation glue.
- Keep the SQLite stores in one Go package for now; split only if the package
  boundary becomes worth the extra complexity.

## Useful validation commands

From `app`:

```powershell
go test ./...
```

From `app/frontend`:

```powershell
npm run typecheck
npm test
npm run build
```

For a focused Wails/backend surface check:

```powershell
cd app
go test ./internal/presentation/wails
```

For a focused desktop bridge check:

```powershell
cd app/frontend
npm test -- --run src/gateways
```

## Good first prompt for a new session

```text
Please continue Sweeters V2. First read:
- docs/development/session-handoff.md
- docs/development/architecture-conventions.md
- docs/development/frontend-gateway-split.md
- UI.md
- todo.md

Do not restore the removed API. Keep the app local-first. We are currently doing
UI work from UI.md before backup/restore. Ask me for visual/product opinions
before making significant UI decisions.
```
