# Development toolchain

Phase 2 pins the desktop toolchain so a local build and CI use the same major
and patch baselines. Application dependencies are exact in `package.json`; Go
tools are module-managed in `go.mod`. No global Wails installation is used.

## Pinned baseline

| Component | Version | Policy |
| --- | ---: | --- |
| Go language | 1.26.0 | Module language and standard-library baseline |
| Go toolchain | 1.26.5 | Exact local-script and CI compiler |
| Wails | 2.13.0 | Latest stable V2; V3 remains prerelease |
| modernc SQLite driver | 1.53.0 | Pure-Go SQLite driver |
| Node.js | 24.17.0 | Active LTS baseline from `.node-version` |
| npm | 11.13.0 | Lockfile and CI package manager |
| React | 19.2.7 | `react` and `react-dom` must match exactly |
| React Router | 8.2.0 | Desktop routing package; no DOM compatibility shim |
| Vite | 8.1.4 | Frontend build tool |
| Tailwind CSS | 4.3.2 | CSS framework and Vite plugin |
| TypeScript | 6.0.3 | Strict incremental migration baseline |

TypeScript 7 is not used yet because the current TypeScript ESLint release
supports TypeScript versions below 6.1. Existing JSX remains valid JavaScript
with `allowJs`; new boundary, configuration, and test files are strict
TypeScript. Feature files migrate vertically in later phases.

## Windows setup

With nvm-windows installed, select the pinned Node and npm versions first:

```powershell
nvm install 24.17.0
nvm use 24.17.0
npm install --global npm@11.13.0
```

Install a Go 1.26 bootstrap if `go` is not already available. The Go command
will download the exact 1.26.5 toolchain declared by the module during the
online setup step.

From the repository root, run:

```powershell
.\scripts\setup-dev.ps1
```

Setup downloads Go and npm dependencies, prepares the module-managed Wails,
Staticcheck, actionlint, and govulncheck tools, and installs Playwright's
Chromium binary. After that, normal development and the main desktop
verification can run from local caches.

## Reproducibility rules

- Change `.node-version`, `package.json` engines, and CI together for a Node
  upgrade.
- Change the Go language/toolchain directives and CI together for a Go upgrade.
- Add Go developer tools with `go get -tool path@version`; invoke them with
  `go tool`, never a GOPATH binary.
- Commit `package-lock.json`, `go.mod`, and `go.sum` with every dependency
  change.
- Runtime packages stay exact. Re-run all checks before accepting an upgrade.
- Generated Wails bindings, frontend build output, browsers, and coverage
  reports are local artifacts and are not committed.
