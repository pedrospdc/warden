package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ProcessEntry describes a process to kill/restore when gaming mode toggles.
type ProcessEntry struct {
	Name    string   `json:"name"`
	Exes    []string `json:"exes"`    // one or more executables (e.g. deluged.exe, deluge-web.exe)
	Enabled bool     `json:"enabled"`
}

// ServiceEntry describes a Windows service to stop/start.
type ServiceEntry struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

// Hotlink is a quick-access link shown in the portal.
type Hotlink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon"`
}

// Config holds the full application configuration.
type Config struct {
	GamingMode bool           `json:"gaming_mode"`
	Processes  []ProcessEntry `json:"processes"`
	Services   []ServiceEntry `json:"services"`
	Hotlinks   []Hotlink      `json:"hotlinks"`
}

// DefaultConfig returns the factory-default configuration.
func DefaultConfig() Config {
	return Config{
		GamingMode: false,
		Processes: []ProcessEntry{
			{Name: "Sonarr", Exes: []string{"Sonarr.exe"}, Enabled: true},
			{Name: "Radarr", Exes: []string{"Radarr.exe"}, Enabled: true},
			{Name: "Prowlarr", Exes: []string{"Prowlarr.exe"}, Enabled: true},
			{Name: "Jellyfin", Exes: []string{"jellyfin.exe"}, Enabled: true},
			{Name: "Deluge", Exes: []string{"deluged.exe", "deluge-web.exe", "deluge.exe"}, Enabled: true},
			{Name: "Docker Desktop", Exes: []string{"Docker Desktop.exe"}, Enabled: true},
		},
		Services: []ServiceEntry{
			{Name: "Windows Search", ID: "WSearch", Enabled: true},
			{Name: "Superfetch (SysMain)", ID: "SysMain", Enabled: true},
			{Name: "Windows Update", ID: "wuauserv", Enabled: true},
			{Name: "Telemetry (DiagTrack)", ID: "DiagTrack", Enabled: true},
			{Name: "WAP Push (dmwappushservice)", ID: "dmwappushservice", Enabled: true},
			{Name: "OneDrive Sync (OneSyncSvc)", ID: "OneSyncSvc", Enabled: true},
			{Name: "Docker Service", ID: "com.docker.service", Enabled: true},
			{Name: "WSL (LxssManager)", ID: "LxssManager", Enabled: true},
		},
		Hotlinks: []Hotlink{
			{Name: "Sonarr", URL: "http://localhost:8989", Icon: "📺"},
			{Name: "Radarr", URL: "http://localhost:7878", Icon: "🎬"},
			{Name: "Lidarr", URL: "http://localhost:8686", Icon: "🎵"},
			{Name: "Readarr", URL: "http://localhost:8787", Icon: "📚"},
			{Name: "Prowlarr", URL: "http://localhost:9696", Icon: "🔍"},
			{Name: "Jellyfin", URL: "http://localhost:8096", Icon: "🎞️"},
		},
	}
}

// configFilePath returns the path to the user config file next to the binary.
func configFilePath() string {
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

// LoadConfig reads the config file and merges it over the defaults.
// If the file does not exist the defaults are returned.
func LoadConfig() (Config, error) {
	cfg := DefaultConfig()
	path := configFilePath()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// SaveConfig writes the current config to disk.
func SaveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath(), data, 0o644)
}
