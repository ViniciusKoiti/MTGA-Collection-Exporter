# Windows packaging spike

OpenSpec `introduce-agentic-go-companion`, task 1.4. Measured on
Windows 11 Home (AMD Ryzen 7 5700U), Go 1.26.6, 2026-08-16.

## Wails (stable v2)

- **Version**: `github.com/wailsapp/wails/v2 v2.14.0`, CLI installed
  via `go install` (`GOPROXY=direct` needed on networks that block the
  Go proxy CDN, storage.googleapis.com).
- **Binary**: `wails build` produces a single `desktop.exe` of
  **14.7 MB** (frontend embedded via `go:embed`), built in 18–32 s on
  this machine including `npm` install, binding generation and Vite.
- **WebView2**: present on this machine (runtime 151.0.4129.86,
  detected by `wails doctor`). Wails v2.14 uses the Go
  WebView2Loader, so no `WebView2Loader.dll` ships next to the exe.
  Windows 11 bundles the WebView2 runtime; Windows 10 machines may
  need the Evergreen runtime — the NSIS installer template Wails
  generates (`build/windows/installer`) can carry that bootstrap.
- **CI**: the package builds only on Windows; the desktop Go package
  is `//go:build windows` with a `!windows` stub, so the Linux CI
  runners build and vet everything without GTK/WebKit.

## SQLite driver (`database/sql`)

- **Chosen**: `modernc.org/sqlite v1.56.0` — pure Go (no cgo), which
  keeps `go test` runnable on every platform and CI runner and avoids
  shipping a C toolchain. This is the driver the whole run-store and
  collection store already use.
- **Migrations**: forward-only, embedded, each in one transaction —
  proven idempotent by `TestMigracoesSaoIdempotentes`, with crash
  atomicity (`TestCheckpointComEfeitosEhAtomico`), dual-worker leases
  and retention all green in the sqlitestore suite on every push.
- **cgo alternative** (`mattn/go-sqlite3`) was not pursued: it would
  demand a C compiler on Windows CI and for every contributor, for a
  marginal performance gain irrelevant at collection sizes (~10k
  cards).

## Open items

- **Clean-machine run**: this machine has the dev toolchain and
  WebView2 already installed. A true clean-machine check (fresh
  Windows VM: offline startup, WebView2 bootstrap path, first-run DB
  creation under `%APPDATA%\MTGA-Companion`) still needs a VM and is
  tracked by task 7.4 together with the no-dev-MCP packaging proof.
