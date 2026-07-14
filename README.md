# Sweeters Desktop

Sweeters is a local-first desktop application for catalog, recipes, purchases,
inventory, production, sales, and backups. It uses Wails, Go, SQLite, React,
Vite, and Tailwind CSS. Application data is stored locally on the user's
machine.

## Prerequisites

- Go in the version declared by `app/go.mod`
- Node.js 24 and npm
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

This command checks formatting, vets and tests the Go code, regenerates the
Wails bindings, builds the frontend, and compiles a desktop executable without
downloading dependencies or requiring remote runtime services.

## Project layout

```text
app/
  database/       SQLite bootstrap and migrations
  model/          Current persistence models
  repository/     SQLite repositories
  service/        Application services exposed through Wails
  frontend/       React desktop interface
docs/app/         Current design and implementation notes
scripts/          Windows setup, development, and verification commands
```

The architecture is being migrated bottom-up. Accepted schema invariants are
implemented and tested first, followed by repositories, application use cases,
Wails contracts, and frontend features.

## Current development policy

- The SQLite schema is the executable persistence contract.
- Posted inventory history is immutable.
- Repository and service behavior must be tested against real temporary SQLite
  databases.
- A feature is complete only when its tests and documentation are updated.
- Generated Wails files and build output are not committed.
