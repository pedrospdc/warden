//go:build !windows

package main

// ServiceInfo holds the ID and display name of a Windows service.
type ServiceInfo struct {
	ID          string
	DisplayName string
}

// ListWindowsServices is a no-op stub for non-Windows platforms.
func ListWindowsServices() []ServiceInfo { return nil }
