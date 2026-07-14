# Testing and quality gates

The default desktop gate is intentionally offline after setup:

```powershell
.\scripts\check-desktop.ps1
```

It rejects remote API references in the desktop source, verifies frontend and
Go formatting, runs ESLint and strict TypeScript, executes Vitest, vets and
statically analyzes Go, validates GitHub Actions workflows, runs Go tests with
shuffled test order, and builds the Wails desktop executable. It also enables
the race detector when local CGO is available; the Linux CI job always enforces
a race-enabled test run.

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
go mod tidy -diff
$packages = go list ./... | Where-Object { $_ -notmatch "/frontend/" }
go vet $packages
go tool staticcheck $packages
go test -race -shuffle=on -count=1 $packages
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
- SQLite migrations, stores, and application commands use real temporary
  SQLite databases rather than SQL mocks.
- React components and the typed desktop bridge use Vitest and Testing Library.
- Playwright covers a small number of critical browser-rendered workflows.
- Desktop builds remain the final integration proof for Wails bindings and
  native packaging.

Every defect fix receives a regression test at the lowest layer that can
express the failure. Coverage is evidence, not a target by itself; tests must
assert domain invariants and observable behavior.

## CI order

CI first runs the quality and dependency-audit job plus the Chromium smoke test.
Only after both pass does it build Windows and macOS desktop artifacts. GitHub
Actions are pinned to immutable commit hashes, with the corresponding release
tag documented beside each pin.

## Transitional binding warning

The current experimental service model can make Wails print `Not found:
time.Time` and list `sql.Null*` types while generating bindings. The desktop
build still succeeds. These persistence-shaped contracts are not accepted V2
interfaces and will disappear when the bottom-up migration reaches the Wails
presentation boundary. New handlers must not copy that legacy pattern.
