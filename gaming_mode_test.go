package main

import (
	"testing"
)

func newTestGamingMode(gamingOn bool) (*GamingMode, *mockManager) {
	cfg := DefaultConfig()
	cfg.GamingMode = gamingOn
	m := &mockManager{}
	gm := NewGamingMode(&cfg, m)
	return gm, m
}

func TestNewGamingMode_IsEnabledMatchesConfig(t *testing.T) {
	gmOff, _ := newTestGamingMode(false)
	if gmOff.IsEnabled() {
		t.Error("expected gaming mode OFF when config says false")
	}

	gmOn, _ := newTestGamingMode(true)
	if !gmOn.IsEnabled() {
		t.Error("expected gaming mode ON when config says true")
	}
}

func TestEnable_SetsEnabledTrue(t *testing.T) {
	gm, _ := newTestGamingMode(false)
	gm.Enable()
	if !gm.IsEnabled() {
		t.Error("Enable() should set gaming mode to true")
	}
}

func TestDisable_SetsEnabledFalse(t *testing.T) {
	gm, _ := newTestGamingMode(true)
	gm.Disable()
	if gm.IsEnabled() {
		t.Error("Disable() should set gaming mode to false")
	}
}

func TestToggle_FlipsFromOffToOn(t *testing.T) {
	gm, _ := newTestGamingMode(false)
	newState, _ := gm.Toggle()
	if !newState {
		t.Error("Toggle from OFF should return true (ON)")
	}
	if !gm.IsEnabled() {
		t.Error("Toggle from OFF should leave gaming mode ON")
	}
}

func TestToggle_FlipsFromOnToOff(t *testing.T) {
	gm, _ := newTestGamingMode(true)
	newState, _ := gm.Toggle()
	if newState {
		t.Error("Toggle from ON should return false (OFF)")
	}
	if gm.IsEnabled() {
		t.Error("Toggle from ON should leave gaming mode OFF")
	}
}

func TestEnable_RunsWslShutdown(t *testing.T) {
	gm, m := newTestGamingMode(false)
	gm.Enable()

	found := false
	for _, cmd := range m.commands {
		if cmd == "wsl --shutdown" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Enable() should run 'wsl --shutdown', got commands: %v", m.commands)
	}
}

func TestEnable_SetsHighPerfPowerPlan(t *testing.T) {
	gm, m := newTestGamingMode(false)
	gm.Enable()

	found := false
	for _, cmd := range m.commands {
		if cmd == "powercfg /setactive "+powerPlanHighPerf {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Enable() should set High Performance power plan, got commands: %v", m.commands)
	}
}

func TestDisable_SetsBalancedPowerPlan(t *testing.T) {
	gm, m := newTestGamingMode(true)
	gm.Disable()

	found := false
	for _, cmd := range m.commands {
		if cmd == "powercfg /setactive "+powerPlanBalanced {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Disable() should set Balanced power plan, got commands: %v", m.commands)
	}
}

func TestEnable_StopsAllEnabledServices(t *testing.T) {
	gm, m := newTestGamingMode(false)
	gm.Enable()

	cfg := DefaultConfig()
	var enabledCount int
	for _, svc := range cfg.Services {
		if svc.Enabled {
			enabledCount++
		}
	}
	if len(m.stopped) != enabledCount {
		t.Errorf("Enable() should stop %d services, stopped %d", enabledCount, len(m.stopped))
	}
}

func TestEnable_KillsAllEnabledProcesses(t *testing.T) {
	gm, m := newTestGamingMode(false)
	gm.Enable()

	cfg := DefaultConfig()
	var exeCount int
	for _, p := range cfg.Processes {
		if p.Enabled {
			exeCount += len(p.Exes)
		}
	}
	if len(m.killed) != exeCount {
		t.Errorf("Enable() should kill %d executables, killed %d", exeCount, len(m.killed))
	}
}

func TestOnChange_CallbackInvokedOnEnable(t *testing.T) {
	gm, _ := newTestGamingMode(false)
	done := make(chan bool, 1)
	gm.OnChange(func(enabled bool) {
		done <- enabled
	})
	gm.Enable()
	got := <-done
	if !got {
		t.Error("OnChange callback should receive true when gaming mode enabled")
	}
}

func TestOnChange_CallbackInvokedOnDisable(t *testing.T) {
	gm, _ := newTestGamingMode(true)
	done := make(chan bool, 1)
	gm.OnChange(func(enabled bool) {
		done <- enabled
	})
	gm.Disable()
	got := <-done
	if got {
		t.Error("OnChange callback should receive false when gaming mode disabled")
	}
}
