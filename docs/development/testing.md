# Testing and quality gates

The default desktop gate is intentionally offline after setup:

```powershell
.\scripts\check-desktop.ps1
```

It rejects known forbidden remote-runtime patterns in the desktop frontend
source, verifies frontend and Go formatting, validates SQL and rejects stale
generated query code, runs
ESLint and strict TypeScript, executes Vitest, vets and statically analyzes Go,
validates GitHub Actions workflows, runs Go tests with shuffled test order, and
builds the Wails desktop executable. It also enables the race detector when
local CGO is available; the Linux CI job always enforces a race-enabled test
run.

## Focused commands

Run frontend checks from `app/frontend`:

```powershell
npm run format:check
npm run lint
npm run typecheck
npm test
npm run test:coverage
npm run build
npm run test:e2e
```

Run Go checks from `app` with `GOTOOLCHAIN=go1.26.5`:

```powershell
go tool sqlc compile
go tool sqlc diff
go mod tidy -diff
$packages = go list ./... | Where-Object { $_ -notmatch "/frontend/" }
go vet $packages
go tool staticcheck $packages
go test -race -shuffle=on -count=1 $packages
```

During migration and schema work, the focused database loop is:

```powershell
Push-Location app
go test -shuffle=on -count=1 ./database ./internal/infrastructure/sqlite
Pop-Location
```

The package filter keeps npm dependencies that happen to contain Go examples
outside the desktop module's quality scope while preserving Wails' required
`go:embed` relationship with `frontend/dist`.

The online audit is separate because vulnerability databases and the npm
registry must be current:

```powershell
.\scripts\audit-desktop.ps1
```

It runs `govulncheck` for reachable Go vulnerabilities and `npm audit` at high
severity or above.

## Test placement

- Domain algorithms use table, property, and fuzz tests close to their Go
  packages.
- SQLite migrations, stores, and application commands use real temporary file
  databases rather than SQL mocks. Migration tests cover fresh initialization,
  reopen, exact history and checksum validation, unsupported/legacy rejection,
  and rollback after a failing migration.
- Schema tests exercise representative strict-type, foreign-key, uniqueness,
  immutable-history, reversal, lot-allocation, and projection constraints. Go
  transaction tests later add aggregate completeness, valuation, FEFO, and
  replay guarantees that cannot be proven by one row constraint.
- React components and the typed desktop bridge use Vitest and Testing Library.
- Playwright covers a small number of critical browser-rendered workflows.
- Desktop builds prove native packaging plus compilation and binding generation
  for services actually bound by Go. They do not prove that dynamically resolved
  route and bridge method names exist; those require Wails-runtime integration
  coverage.

Every defect fix receives a regression test at the lowest layer that can
express the failure. Coverage is evidence, not a target by itself; tests must
assert domain invariants and observable behavior.

## CI order

CI first runs the quality and dependency-audit job, the Chromium smoke test, and
the Linux race job. Only after all three pass does it build Windows and macOS
desktop artifacts. GitHub Actions are pinned to immutable commit hashes, with
the corresponding release tag documented beside each pin.

## Phase 4 persistence boundary

The experimental models, repositories, and domain services targeted tables
that the V2 baseline deliberately removed, so Phase 4 deletes them instead of
keeping a parallel compatibility model. Their replacements are strong domain
snapshots, committed generated queries, and aggregate stores tested against
temporary V2 databases. Inventory is read-only at this boundary. Phase 5 adds
posting/application commands; later layers add presentation DTOs and replace
the still-unavailable legacy page workflows.

Restore/import is also intentionally disabled in Phase 3. Tests must prove that
a restore request cannot replace the live file. Re-enabling it requires staged
identity, schema, checksum, integrity, foreign-key, and replay validation plus
a safety backup, atomic replacement, and process restart.
