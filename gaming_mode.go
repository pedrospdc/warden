package main

import (
	"sync"
)

// powerPlanHighPerf is the Windows GUID for the High Performance power plan.
const powerPlanHighPerf = "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c"

// powerPlanBalanced is the Windows GUID for the Balanced power plan.
const powerPlanBalanced = "381b4222-f694-41f0-9685-ff5bb260df2e"

// GamingMode manages the enabled/disabled state of gaming mode and orchestrates
// service/process control through a ServiceManager.
type GamingMode struct {
	mu        sync.Mutex
	enabled   bool
	cfg       *Config
	manager   ServiceManager
	listeners []func(enabled bool)
}

// NewGamingMode constructs a GamingMode with the given config and manager.
func NewGamingMode(cfg *Config, manager ServiceManager) *GamingMode {
	return &GamingMode{
		enabled: cfg.GamingMode,
		cfg:     cfg,
		manager: manager,
	}
}

// IsEnabled returns the current gaming mode state (safe for concurrent use).
func (g *GamingMode) IsEnabled() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.enabled
}

// OnChange registers a callback that is called after every state change.
func (g *GamingMode) OnChange(fn func(bool)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.listeners = append(g.listeners, fn)
}

// Enable switches gaming mode on: kills processes, stops services, and runs
// the power and WSL commands.  It persists the new state to disk.
func (g *GamingMode) Enable() []ActionResult {
	g.mu.Lock()
	defer g.mu.Unlock()

	var results []ActionResult

	results = append(results, KillAll(g.manager, *g.cfg)...)
	results = append(results, StopAll(g.manager, *g.cfg)...)

	// Shut down WSL
	if err := g.manager.RunCommand("wsl", "--shutdown"); err != nil {
		results = append(results, ActionResult{Name: "wsl --shutdown", Err: err})
	} else {
		results = append(results, ActionResult{Name: "wsl --shutdown"})
	}

	// Switch to High Performance power plan
	if err := g.manager.RunCommand("powercfg", "/setactive", powerPlanHighPerf); err != nil {
		results = append(results, ActionResult{Name: "powercfg High Performance", Err: err})
	} else {
		results = append(results, ActionResult{Name: "powercfg High Performance"})
	}

	g.enabled = true
	g.cfg.GamingMode = true
	_ = SaveConfig(*g.cfg)
	g.notify(true)
	return results
}

// Disable switches gaming mode off: starts services back, restores balanced
// power plan.  It persists the new state to disk.
func (g *GamingMode) Disable() []ActionResult {
	g.mu.Lock()
	defer g.mu.Unlock()

	var results []ActionResult

	results = append(results, StartAll(g.manager, *g.cfg)...)

	// Restore Balanced power plan
	if err := g.manager.RunCommand("powercfg", "/setactive", powerPlanBalanced); err != nil {
		results = append(results, ActionResult{Name: "powercfg Balanced", Err: err})
	} else {
		results = append(results, ActionResult{Name: "powercfg Balanced"})
	}

	g.enabled = false
	g.cfg.GamingMode = false
	_ = SaveConfig(*g.cfg)
	g.notify(false)
	return results
}

// Toggle flips the current gaming mode state and returns the new state.
func (g *GamingMode) Toggle() (bool, []ActionResult) {
	if g.IsEnabled() {
		results := g.Disable()
		return false, results
	}
	results := g.Enable()
	return true, results
}

// notify calls all registered listeners (must be called with mu held).
func (g *GamingMode) notify(enabled bool) {
	for _, fn := range g.listeners {
		go fn(enabled)
	}
}
