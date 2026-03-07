//go:build !windows

package main

import (
	"fmt"
	"os/exec"
)

// OSServiceManager is the real implementation for non-Windows platforms
// (Linux / macOS) using systemctl and pkill.
type OSServiceManager struct{}

// NewOSServiceManager returns the platform service manager.
func NewOSServiceManager() ServiceManager { return &OSServiceManager{} }

func (m *OSServiceManager) StopService(id string) error {
	return m.RunCommand("systemctl", "stop", id)
}

func (m *OSServiceManager) StartService(id string) error {
	return m.RunCommand("systemctl", "start", id)
}

func (m *OSServiceManager) KillProcess(exe string) error {
	// pkill matches against process name; strip .exe suffix if present
	name := exe
	if len(name) > 4 && name[len(name)-4:] == ".exe" {
		name = name[:len(name)-4]
	}
	cmd := exec.Command("pkill", "-f", name)
	err := cmd.Run()
	if err == nil {
		return nil
	}
	// pkill exits with code 1 when no matching process was found — that is
	// not an error for our use-case (the process may already be stopped).
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return nil
	}
	return fmt.Errorf("pkill %q: %w", name, err)
}

func (m *OSServiceManager) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}
