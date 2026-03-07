package main

import (
	"errors"
	"testing"
)

// mockManager records calls so we can assert on them in tests.
type mockManager struct {
	stopped  []string
	started  []string
	killed   []string
	commands []string
	failOn   string // if non-empty, calls that match this string return an error
}

func (m *mockManager) StopService(id string) error {
	if m.failOn == id {
		return errors.New("mock stop error")
	}
	m.stopped = append(m.stopped, id)
	return nil
}

func (m *mockManager) StartService(id string) error {
	if m.failOn == id {
		return errors.New("mock start error")
	}
	m.started = append(m.started, id)
	return nil
}

func (m *mockManager) KillProcess(exe string) error {
	if m.failOn == exe {
		return errors.New("mock kill error")
	}
	m.killed = append(m.killed, exe)
	return nil
}

func (m *mockManager) RunCommand(name string, args ...string) error {
	key := name
	for _, a := range args {
		key += " " + a
	}
	if m.failOn == key {
		return errors.New("mock run error")
	}
	m.commands = append(m.commands, key)
	return nil
}

// ── StopAll ───────────────────────────────────────────────────────────────────

func TestStopAll_CallsEnabledServices(t *testing.T) {
	cfg := Config{
		Services: []ServiceEntry{
			{Name: "Sonarr", ID: "Sonarr", Enabled: true},
			{Name: "Radarr", ID: "Radarr", Enabled: true},
			{Name: "Disabled", ID: "Disabled", Enabled: false},
		},
	}
	m := &mockManager{}
	results := StopAll(m, cfg)

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if len(m.stopped) != 2 {
		t.Errorf("expected 2 StopService calls, got %d", len(m.stopped))
	}
}

func TestStopAll_ReportsFailure(t *testing.T) {
	cfg := Config{
		Services: []ServiceEntry{
			{Name: "WSearch", ID: "WSearch", Enabled: true},
		},
	}
	m := &mockManager{failOn: "WSearch"}
	results := StopAll(m, cfg)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OK() {
		t.Error("expected failure result, got OK")
	}
}

// ── StartAll ──────────────────────────────────────────────────────────────────

func TestStartAll_CallsEnabledServices(t *testing.T) {
	cfg := Config{
		Services: []ServiceEntry{
			{Name: "SysMain", ID: "SysMain", Enabled: true},
			{Name: "Off", ID: "Off", Enabled: false},
		},
	}
	m := &mockManager{}
	results := StartAll(m, cfg)

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK() {
		t.Error("expected success result")
	}
}

// ── KillAll ───────────────────────────────────────────────────────────────────

func TestKillAll_KillsAllExes(t *testing.T) {
	cfg := Config{
		Processes: []ProcessEntry{
			{Name: "Deluge", Exes: []string{"deluged.exe", "deluge-web.exe", "deluge.exe"}, Enabled: true},
			{Name: "Disabled", Exes: []string{"disabled.exe"}, Enabled: false},
		},
	}
	m := &mockManager{}
	results := KillAll(m, cfg)

	// Deluge has 3 exes, disabled has 0
	if len(results) != 3 {
		t.Errorf("expected 3 results for Deluge exes, got %d", len(results))
	}
	if len(m.killed) != 3 {
		t.Errorf("expected 3 KillProcess calls, got %d", len(m.killed))
	}
}

func TestKillAll_SkipsDisabled(t *testing.T) {
	cfg := Config{
		Processes: []ProcessEntry{
			{Name: "Disabled", Exes: []string{"disabled.exe"}, Enabled: false},
		},
	}
	m := &mockManager{}
	KillAll(m, cfg)
	if len(m.killed) != 0 {
		t.Errorf("expected no KillProcess calls for disabled processes, got %d", len(m.killed))
	}
}

// ── ActionResult ──────────────────────────────────────────────────────────────

func TestActionResult_OK(t *testing.T) {
	ok := ActionResult{Name: "test", Err: nil}
	fail := ActionResult{Name: "test", Err: errors.New("oops")}

	if !ok.OK() {
		t.Error("nil-error result should be OK")
	}
	if fail.OK() {
		t.Error("error result should not be OK")
	}
}
