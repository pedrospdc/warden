# Warden

Windows system-tray app that kills background processes and stops services when you game, then restores everything when you're done.

## What it does

- **Enable** — kills configured processes, stops Windows services, shuts down WSL, switches to High Performance power plan
- **Disable** — starts services back up, returns to Balanced power plan
- **Tray-first** — left-click the icon to open the portal, right-click for the menu
- **Manage Services** — toggle, add, edit, or delete entries from the UI; fuzzy-search picker auto-fills from your live Windows service list

## Usage

Download `warden.exe` from [Releases](../../releases) and run it. No installer needed — configuration is saved to `config.json` next to the binary.

Default managed items: Sonarr, Radarr, Prowlarr, Jellyfin, Deluge, Docker Desktop (processes) and Windows Search, SysMain, Windows Update, DiagTrack, OneDrive Sync, Docker Service, WSL (services).

## Building

Requires Go 1.24+ and [mingw-w64](https://www.mingw-w64.org/) for cross-compilation from Linux/WSL.

```bash
make build-windows   # produces warden.exe
make deploy          # build + copy to %USERPROFILE%\Documents
make test
```
