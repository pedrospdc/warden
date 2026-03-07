# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...
go build -o warden.exe .        # Windows release binary

# Test (all tests)
go test ./...

# Run a single test
go test -run TestGamingMode_IsEnabled ./...

# Lint (requires system deps on Linux: gcc libgl1-mesa-dev xorg-dev)
go vet ./...
```

Tests run on Linux via CI (ubuntu-latest); the Windows binary is only built in the release workflow triggered on `main`.

## Architecture

Warden is a single `package main` Go application — a system-tray app (Fyne) that optimizes the system for gaming by stopping background processes and services.

**Core layers:**

- **`config.go`** — `Config` struct (JSON-backed), `LoadConfig`/`SaveConfig`. Config file lives next to the binary (`config.json`). Defaults ship with opinionated process/service lists (Sonarr, Radarr, Docker, etc.).

- **`service_manager.go`** — `ServiceManager` interface + `StopAll`/`StartAll`/`KillAll` helpers that operate on `Config`. The interface is the seam for testing — no real OS calls in tests.
  - `service_manager_windows.go` — Windows impl: `net stop/start`, `taskkill`
  - `service_manager_other.go` — Linux/macOS impl: `systemctl`, `pkill` (strips `.exe` suffix)

- **`gaming_mode.go`** — `GamingMode` struct orchestrates: kill processes → stop services → `wsl --shutdown` → `powercfg` (High Performance GUID). Disable reverses: start services → Balanced power plan. State is mutex-protected; `OnChange` listeners allow the tray and portal to stay in sync.

- **`portal.go`** — Fyne window ("Portal") with three cards: Gaming Mode toggle, Managed Services status list, and Hotlinks grid. Calls `gm.OnChange` to refresh UI when state changes externally (from tray).

- **`main.go`** — Entry point. Builds the system-tray menu via `fyne/v2/driver/desktop`. Falls back to opening the portal directly if no system tray is available. Tray menu and icon update via `gm.OnChange`.

**Key design decisions:**
- `GamingMode` holds a `*Config` pointer; `Enable`/`Disable` call `SaveConfig` to persist state.
- `OnChange` callbacks fire in goroutines (`go fn(enabled)`), so UI refresh must be thread-safe via Fyne's own refresh methods.
- The mock (`mockManager` in `service_manager_test.go`) records all calls; tests use `failOn` to simulate errors.

## Platform targeting

Primary target is **Windows** (the release binary is `warden.exe`). The `//go:build windows` / `//go:build !windows` split in `service_manager_*.go` keeps the non-Windows stub functional for local dev and CI on Linux.

## Versioning

The `VERSION` file holds the base semver (`major.minor.patch`). CI appends the GitHub run number: `v{VERSION}.{run_number}`.
