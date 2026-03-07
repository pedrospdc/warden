//go:build windows

package main

import (
	"os/exec"
	"strings"
)

// ServiceInfo holds the ID and display name of a Windows service.
type ServiceInfo struct {
	ID          string
	DisplayName string
}

// ListWindowsServices returns all installed Windows services via PowerShell.
func ListWindowsServices() []ServiceInfo {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-Service | ForEach-Object { $_.Name + '|' + $_.DisplayName }").Output()
	if err != nil {
		return nil
	}
	var services []ServiceInfo
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, "|", 2)
		if len(parts) == 2 && parts[0] != "" {
			services = append(services, ServiceInfo{
				ID:          strings.TrimSpace(parts[0]),
				DisplayName: strings.TrimSpace(parts[1]),
			})
		}
	}
	return services
}
