# Warden

Windows system-tray app that kills background processes and stops services when you game, then restores everything when you're done.

## What it does

- **Enable** — kills configured processes, stops Windows services, shuts down WSL, switches to High Performance power plan
- **Disable** — starts services back up, returns to Balanced power plan
- **Tray-first** — left-click to open the portal, right-click for the menu (toggle gaming mode, Start with Windows, quit)
- **Manage Services** — toggle, add, edit, or delete entries from the UI; fuzzy-search picker auto-fills from your live Windows service list
- **Start with Windows** — toggleable from the tray menu; the MSI installer enables it by default

## Installation

Download `warden.msi` from [Releases](../../releases) for a guided install (sets up Program Files, Start Menu shortcut, and autostart). Or grab the standalone `warden.exe` and run it directly — configuration is saved to `config.json` next to the binary.

Default managed items: Sonarr, Radarr, Prowlarr, Jellyfin, Deluge, Docker Desktop (processes) and Windows Search, SysMain, Windows Update, DiagTrack, OneDrive Sync, Docker Service, WSL (services).

## Building

Requires Go 1.24+, [mingw-w64](https://www.mingw-w64.org/), and `msitools` (`sudo apt install msitools`) for cross-compilation from Linux/WSL.

```bash
make build-windows   # produces warden.exe
make build-msi       # produces warden-<version>.msi
make deploy          # build + copy warden.exe to %USERPROFILE%\Documents
make test
```
