package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.GamingMode {
		t.Error("default config should have gaming mode OFF")
	}

	// ── Processes ──────────────────────────────────────────────────────────────
	wantProcs := map[string]bool{
		"Sonarr":         false,
		"Radarr":         false,
		"Prowlarr":       false,
		"Jellyfin":       false,
		"Deluge":         false,
		"Docker Desktop": false,
	}
	for _, p := range cfg.Processes {
		wantProcs[p.Name] = true
	}
	for name, found := range wantProcs {
		if !found {
			t.Errorf("default processes missing %q", name)
		}
	}

	// ── Services ───────────────────────────────────────────────────────────────
	wantSvcs := map[string]bool{
		"WSearch":          false,
		"SysMain":          false,
		"wuauserv":         false,
		"DiagTrack":        false,
		"dmwappushservice": false,
		"OneSyncSvc":       false,
		"com.docker.service": false,
		"LxssManager":      false,
	}
	for _, s := range cfg.Services {
		wantSvcs[s.ID] = true
	}
	for id, found := range wantSvcs {
		if !found {
			t.Errorf("default services missing service ID %q", id)
		}
	}

	// ── Hotlinks ───────────────────────────────────────────────────────────────
	if len(cfg.Hotlinks) == 0 {
		t.Error("default config should have at least one hotlink")
	}
}

func TestDelugeHasMultipleExes(t *testing.T) {
	cfg := DefaultConfig()
	for _, p := range cfg.Processes {
		if p.Name == "Deluge" {
			if len(p.Exes) < 2 {
				t.Errorf("Deluge should list multiple executables, got %v", p.Exes)
			}
			return
		}
	}
	t.Error("Deluge not found in default processes")
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	// Temporarily redirect the config path by writing and reading manually.
	original := DefaultConfig()
	original.GamingMode = true

	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read it back.
	readData, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var loaded Config
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !loaded.GamingMode {
		t.Error("loaded config should have gaming mode ON")
	}
	if len(loaded.Processes) != len(original.Processes) {
		t.Errorf("process count mismatch: got %d want %d", len(loaded.Processes), len(original.Processes))
	}
	if len(loaded.Services) != len(original.Services) {
		t.Errorf("service count mismatch: got %d want %d", len(loaded.Services), len(original.Services))
	}
}
