//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const startupRegKey  = `Software\Microsoft\Windows\CurrentVersion\Run`
const startupRegName = "Warden"

// isStartupEnabled reports whether the Warden run key exists in HKCU.
func isStartupEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, startupRegKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(startupRegName)
	return err == nil
}

// setStartupEnabled adds or removes the HKCU run key for Warden.
func setStartupEnabled(enabled bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, startupRegKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if enabled {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		return k.SetStringValue(startupRegName, `"`+exe+`"`)
	}
	return k.DeleteValue(startupRegName)
}
