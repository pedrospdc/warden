package main

import (
	_ "embed"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
)

//go:embed assets/icon-normal.png
var iconNormalBytes []byte

//go:embed assets/icon-gaming.png
var iconGamingBytes []byte

func normalIcon() fyne.Resource {
	return fyne.NewStaticResource("icon-normal.png", iconNormalBytes)
}

func gamingIcon() fyne.Resource {
	return fyne.NewStaticResource("icon-gaming.png", iconGamingBytes)
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Printf("Warning: could not load config: %v (using defaults)", err)
	}

	manager := NewOSServiceManager()
	gm := NewGamingMode(&cfg, manager)

	a := app.NewWithID("com.warden.app")
	a.SetIcon(normalIcon())

	// Warden lives in the system tray — there is no main window.
	// On macOS the Dock icon is hidden automatically when there are no windows.
	if drv, ok := a.(interface{ SetDockVisible(bool) }); ok {
		drv.SetDockVisible(false)
	}

	desk, isDesktop := a.(desktop.App)
	if !isDesktop {
		// Fallback: open the portal directly if tray is unavailable.
		ShowPortal(a, gm, &cfg)
		a.Run()
		return
	}

	// ── Build tray menu ───────────────────────────────────────────────────────
	var (
		statusItem *fyne.MenuItem
		toggleItem *fyne.MenuItem
		portalItem *fyne.MenuItem
		quitItem   *fyne.MenuItem
	)

	buildMenu := func() *fyne.Menu {
		if gm.IsEnabled() {
			statusItem = fyne.NewMenuItem("🎮 Gaming Mode: ON", nil)
			toggleItem = fyne.NewMenuItem("Disable Gaming Mode", func() {
				go func() {
					gm.Disable()
					updateTray(desk, gm)
				}()
			})
		} else {
			statusItem = fyne.NewMenuItem("🖥️  Gaming Mode: OFF", nil)
			toggleItem = fyne.NewMenuItem("Enable Gaming Mode", func() {
				go func() {
					gm.Enable()
					updateTray(desk, gm)
				}()
			})
		}
		statusItem.Disabled = true

		portalItem = fyne.NewMenuItem("Open Portal…", func() {
			ShowPortal(a, gm, &cfg)
		})
		quitItem = fyne.NewMenuItem("Quit Warden", func() {
			a.Quit()
		})

		return fyne.NewMenu("Warden",
			statusItem,
			fyne.NewMenuItemSeparator(),
			toggleItem,
			fyne.NewMenuItemSeparator(),
			portalItem,
			fyne.NewMenuItemSeparator(),
			quitItem,
		)
	}

	desk.SetSystemTrayIcon(normalIcon())
	desk.SetSystemTrayMenu(buildMenu())

	// Re-register listeners so gaming mode changes triggered via the portal
	// also update the tray icon and menu.
	gm.OnChange(func(_ bool) {
		updateTray(desk, gm)
		desk.SetSystemTrayMenu(buildMenu())
	})

	a.Run()
}

// updateTray swaps the tray icon to reflect the current gaming mode state.
func updateTray(desk desktop.App, gm *GamingMode) {
	if gm.IsEnabled() {
		desk.SetSystemTrayIcon(gamingIcon())
	} else {
		desk.SetSystemTrayIcon(normalIcon())
	}
}
