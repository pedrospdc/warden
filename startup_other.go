//go:build !windows

package main

func isStartupEnabled() bool      { return false }
func setStartupEnabled(bool) error { return nil }
