# Sweeters Desktop

Sweeters is a local-first desktop application for catalog, recipes, purchases,
inventory, production, sales, and backups. It uses Wails, Go, SQLite, React,
Vite, and Tailwind CSS. Application data is stored locally on the user's
machine.

> **Current capability:** Phase 5 is in progress. The V2 schema, domain model,
> aggregate SQLite stores, application/Wails contracts, typed desktop gateway,
> and first real desktop screens are in place for settings, measurement units,
> catalog items, item packagings, and counterparties. Purchase posting has
> started at the backend/Wails/gateway layer: a purchase can create an immutable
> stock document, inbound line, inventory lot, and updated balance, and the
> desktop contract can read inventory balances, item lots, ledger entries, and
> line allocations. Legacy pages for recipe, production, sales, and reporting
> workflows are still not real V2 workflows.

The `api/` directory is parked legacy code. It is outside the desktop runtime,
development setup, verification gates, and CI, and this repository does not
deploy it.

## Prerequisites

- Go 1.26 or newer as a bootstrap for the pinned Go 1.26.5 toolchain
- Node.js 24.17.0 and npm 11.13.0
- WebView2 on Windows

The setup command downloads development dependencies once. After setup, the
desktop build and checks can run using the local dependency caches.

## Setup

From Windows PowerShell at the repository root:

```powershell
.\scripts\setup-dev.ps1
```

## Run in development

```powershell
.\scripts\dev.ps1
```

The development script keeps its database in:

```text
%APPDATA%\saas-dev\app.db
```

To use another data directory:

```powershell
$env:SAAS_DATA_DIR = "C:\temp\sweeters-data"
.\scripts\dev.ps1
```

## Verify the desktop application

```powershell
.\scripts\check-desktop.ps1
```

This command checks formatting, validates named SQL and committed generated
queries, vets and tests the Go code, builds the frontend with Vite, runs ESLint,
TypeScript, Vitest, Staticcheck, and the Go race detector when supported locally,
then compiles a desktop executable without downloading dependencies or requiring
remote runtime services. CI always enforces the race detector on Linux.

The browser smoke test and online dependency audit are separate because they
need a downloaded browser or current vulnerability data:

```powershell
Push-Location app\frontend
npm run test:e2e
Pop-Location
.\scripts\audit-desktop.ps1
```

## Project layout

```text
app/
  database/       SQLite lifecycle, migrations, and write coordinator
  internal/
    domain/       Strong values and aggregate snapshots
    infrastructure/sqlite/
                  Named SQL, generated queries, and aggregate stores
  service/        Operating-system/database lifecycle Wails services
  frontend/       React desktop interface
api/              Parked legacy code; excluded from the desktop application
docs/             Architecture, domain contract, ADRs, and historical archive
scripts/          Windows setup, development, and verification commands
```

The architecture is being migrated bottom-up. The accepted schema, strong
domain model, and aggregate persistence boundary are implemented and tested
before application use cases, Wails contracts, and frontend features.

Start with the [documentation index](docs/README.md) for the accepted V2
decisions, target data model, glossary, invariants, and use cases.
The [toolchain guide](docs/development/toolchain.md) contains exact versions and
upgrade instructions; the [testing guide](docs/development/testing.md) explains
the local and CI quality gates.

## Current development policy

- The SQLite schema is the executable persistence contract.
- Posted inventory history is immutable.
- Repository and service behavior must be tested against real temporary SQLite
  databases.
- Generated SQLite queries are committed and must match their named SQL source.
- React components call the typed desktop bridge instead of generated Wails
  modules.
- A feature is complete only when its tests and documentation are updated.
- Generated Wails files and build output are not committed.
