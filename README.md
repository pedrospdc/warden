# Warden

Windows system-tray app that optimizes your PC for gaming by stopping heavy background processes and services with one click.

## What it does

- Kills configured background processes (Sonarr, Radarr, Jellyfin, Docker, etc.)
- Stops Windows services (Windows Search, Windows Update, SysMain, DiagTrack, WSL, etc.)
- Shuts down WSL
- Switches the power plan to High Performance

Disabling gaming mode reverses all of the above (restores services, switches back to Balanced power plan).

## Usage

Download `warden.exe` from [Releases](../../releases) and run it. Warden lives in the system tray — right-click the icon to toggle gaming mode or open the Portal.

The Portal window shows the current state of all managed services and provides quick-access links to your local *arr apps.

## Configuration

`config.json` is created next to the binary on first save. Edit it to add or remove processes, services, and hotlinks.

```json
{
  "gaming_mode": false,
  "processes": [
    { "name": "Sonarr", "exes": ["Sonarr.exe"], "enabled": true }
  ],
  "services": [
    { "name": "Windows Search", "id": "WSearch", "enabled": true }
  ],
  "hotlinks": [
    { "name": "Sonarr", "url": "http://localhost:8989", "icon": "📺" }
  ]
}
```

## Building

Requires Go 1.24+ and [mingw-w64](https://www.mingw-w64.org/) for cross-compilation from Linux/WSL.

```bash
# Windows binary (from WSL/Linux)
make build-windows

# Run tests
make test
```

Native Windows build:

```bash
go build -ldflags="-H windowsgui" -o warden.exe .
```
