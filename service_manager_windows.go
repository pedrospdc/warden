//go:build windows

package main

import (
	"fmt"
	"os/exec"
)

// OSServiceManager is the real implementation for Windows.
type OSServiceManager struct{}

// NewOSServiceManager returns the platform service manager.
func NewOSServiceManager() ServiceManager { return &OSServiceManager{} }

func (m *OSServiceManager) StopService(id string) error {
	return m.RunCommand("net", "stop", id)
}

func (m *OSServiceManager) StartService(id string) error {
	return m.RunCommand("net", "start", id)
}

func (m *OSServiceManager) KillProcess(exe string) error {
	return m.RunCommand("taskkill", "/IM", exe, "/F")
}

func (m *OSServiceManager) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}
