package main

// ServiceManager abstracts OS-level process and service control so that
// the core logic can be tested without executing real system commands.
type ServiceManager interface {
	// StopService stops a named OS service (Windows service / systemd unit).
	StopService(id string) error
	// StartService starts a named OS service.
	StartService(id string) error
	// KillProcess terminates all running instances of the given executable name.
	KillProcess(exe string) error
	// RunCommand executes an arbitrary command with the given arguments.
	RunCommand(name string, args ...string) error
}

// StopAll stops every enabled service in cfg.Services using the provided manager.
func StopAll(m ServiceManager, cfg Config) []ActionResult {
	results := make([]ActionResult, 0)
	for _, svc := range cfg.Services {
		if !svc.Enabled {
			continue
		}
		err := m.StopService(svc.ID)
		results = append(results, ActionResult{Name: svc.Name, Err: err})
	}
	return results
}

// StartAll starts every enabled service in cfg.Services.
func StartAll(m ServiceManager, cfg Config) []ActionResult {
	results := make([]ActionResult, 0)
	for _, svc := range cfg.Services {
		if !svc.Enabled {
			continue
		}
		err := m.StartService(svc.ID)
		results = append(results, ActionResult{Name: svc.Name, Err: err})
	}
	return results
}

// KillAll kills every executable in every enabled ProcessEntry.
func KillAll(m ServiceManager, cfg Config) []ActionResult {
	results := make([]ActionResult, 0)
	for _, proc := range cfg.Processes {
		if !proc.Enabled {
			continue
		}
		for _, exe := range proc.Exes {
			err := m.KillProcess(exe)
			results = append(results, ActionResult{Name: proc.Name + " (" + exe + ")", Err: err})
		}
	}
	return results
}

// ActionResult captures the outcome of a single stop/start/kill operation.
type ActionResult struct {
	Name string
	Err  error
}

// OK returns true when the action succeeded.
func (r ActionResult) OK() bool { return r.Err == nil }
