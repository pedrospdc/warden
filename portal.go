package main

import (
	"fmt"
	"image/color"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// portalWindow holds references to all live UI elements so they can be
// refreshed when the gaming mode state changes from outside the window
// (e.g. tray icon left-click).
type portalWindow struct {
	win          fyne.Window
	servicesWin  fyne.Window
	gm           *GamingMode
	cfg          *Config
	statusLabel  *widget.Label
	toggleBtn    *widget.Button
	servicesBox  *fyne.Container
	serviceItems []*serviceRow
}

// serviceRow tracks the status dot for a single managed entry so refresh()
// can update its colour without rebuilding the whole list.
type serviceRow struct {
	dot *canvas.Circle
}

// ShowPortal creates (or focuses) the portal window.
func ShowPortal(a fyne.App, gm *GamingMode, cfg *Config) {
	pw := &portalWindow{gm: gm, cfg: cfg}
	pw.build(a)
	pw.refresh()

	// Re-render whenever gaming mode changes (from tray or another portal).
	gm.OnChange(func(_ bool) {
		pw.refresh()
	})
}

func (pw *portalWindow) build(a fyne.App) {
	pw.win = a.NewWindow("Warden – Portal")
	pw.win.Resize(fyne.NewSize(700, 600))
	pw.win.SetFixedSize(false)

	// ── Header ────────────────────────────────────────────────────────────────
	title := widget.NewLabelWithStyle("🛡️  Warden", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	pw.statusLabel = widget.NewLabel("Gaming Mode: OFF")
	header := container.NewBorder(nil, nil, title, pw.statusLabel)

	// ── Gaming Mode Card ──────────────────────────────────────────────────────
	desc := widget.NewLabel("Stop heavy background services and switch to High Performance power plan.")
	desc.Wrapping = fyne.TextWrapWord

	pw.toggleBtn = widget.NewButton("⚔️  Enable Gaming Mode", func() {
		pw.toggleBtn.Disable()
		go func() {
			pw.gm.Toggle()
			pw.refresh()
			pw.toggleBtn.Enable()
		}()
	})
	pw.toggleBtn.Importance = widget.HighImportance

	manageBtn := widget.NewButton("⚙️  Manage Services…", func() {
		pw.showManageServicesWindow(a)
	})
	gamingCard := widget.NewCard("Gaming Mode", "", container.NewVBox(desc, pw.toggleBtn, manageBtn))

	// ── Hotlinks Card ─────────────────────────────────────────────────────────
	hotlinksGrid := container.NewGridWithColumns(3)
	for _, link := range pw.cfg.Hotlinks {
		link := link
		btn := widget.NewButton(fmt.Sprintf("%s  %s", link.Icon, link.Name), func() {
			parsed, err := url.Parse(link.URL)
			if err == nil {
				a.OpenURL(parsed)
			}
		})
		hotlinksGrid.Add(btn)
	}
	hotlinksCard := widget.NewCard("Quick Access", "Open your *arr apps and media services.", hotlinksGrid)

	// ── Layout ────────────────────────────────────────────────────────────────
	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		gamingCard,
		hotlinksCard,
	)
	pw.win.SetContent(container.NewPadded(container.NewScroll(content)))
	pw.win.Show()
}

// showManageServicesWindow opens a dedicated window for managing services and processes.
func (pw *portalWindow) showManageServicesWindow(a fyne.App) {
	if pw.servicesWin != nil {
		pw.servicesWin.RequestFocus()
		return
	}
	pw.servicesBox = container.NewVBox()
	pw.rebuildServiceList()

	pw.servicesWin = a.NewWindow("Warden – Manage Services")
	pw.servicesWin.Resize(fyne.NewSize(540, 600))
	pw.servicesWin.SetContent(container.NewPadded(container.NewScroll(pw.servicesBox)))
	pw.servicesWin.SetOnClosed(func() { pw.servicesWin = nil })
	pw.servicesWin.Show()
}

// rebuildServiceList clears and regenerates the services/processes VBox.
// Call this after any mutation (add, delete, toggle) to keep the list in sync.
func (pw *portalWindow) rebuildServiceList() {
	if pw.servicesBox == nil {
		return
	}
	pw.servicesBox.Objects = nil
	pw.serviceItems = nil

	// ── Windows Services ──────────────────────────────────────────────────────
	pw.servicesBox.Add(widget.NewLabelWithStyle("Windows Services", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	for i := range pw.cfg.Services {
		pw.servicesBox.Add(pw.makeServiceRow(i))
	}
	addSvcBtn := widget.NewButton("+ Add Service", func() { pw.showAddServiceDialog() })
	addSvcBtn.Importance = widget.LowImportance
	pw.servicesBox.Add(addSvcBtn)

	// ── Processes ─────────────────────────────────────────────────────────────
	pw.servicesBox.Add(widget.NewSeparator())
	pw.servicesBox.Add(widget.NewLabelWithStyle("Processes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	for i := range pw.cfg.Processes {
		pw.servicesBox.Add(pw.makeProcRow(i))
	}
	addProcBtn := widget.NewButton("+ Add Process", func() { pw.showAddProcessDialog() })
	addProcBtn.Importance = widget.LowImportance
	pw.servicesBox.Add(addProcBtn)

	pw.servicesBox.Refresh()
}

func (pw *portalWindow) makeServiceRow(idx int) fyne.CanvasObject {
	svc := &pw.cfg.Services[idx]

	dot, dotBox := makeDot(theme.SuccessColor())
	pw.serviceItems = append(pw.serviceItems, &serviceRow{dot: dot})

	check := widget.NewCheck("", func(checked bool) {
		pw.cfg.Services[idx].Enabled = checked
		_ = SaveConfig(*pw.cfg)
	})
	check.Checked = svc.Enabled

	lbl := widget.NewLabel(svc.Name)

	tag := widget.NewLabel("service")
	tag.TextStyle = fyne.TextStyle{Italic: true}

	delBtn := widget.NewButton("✕", func() {
		pw.cfg.Services = append(pw.cfg.Services[:idx], pw.cfg.Services[idx+1:]...)
		_ = SaveConfig(*pw.cfg)
		pw.rebuildServiceList()
	})
	delBtn.Importance = widget.LowImportance

	return container.NewBorder(nil, nil,
		container.NewHBox(dotBox, check),
		container.NewHBox(tag, delBtn),
		lbl,
	)
}

func (pw *portalWindow) makeProcRow(idx int) fyne.CanvasObject {
	proc := &pw.cfg.Processes[idx]

	dot, dotBox := makeDot(theme.SuccessColor())
	pw.serviceItems = append(pw.serviceItems, &serviceRow{dot: dot})

	check := widget.NewCheck("", func(checked bool) {
		pw.cfg.Processes[idx].Enabled = checked
		_ = SaveConfig(*pw.cfg)
	})
	check.Checked = proc.Enabled

	label := proc.Name
	if len(proc.Exes) > 1 {
		label = fmt.Sprintf("%s (%d exes)", proc.Name, len(proc.Exes))
	}
	lbl := widget.NewLabel(label)

	tag := widget.NewLabel("process")
	tag.TextStyle = fyne.TextStyle{Italic: true}

	delBtn := widget.NewButton("✕", func() {
		pw.cfg.Processes = append(pw.cfg.Processes[:idx], pw.cfg.Processes[idx+1:]...)
		_ = SaveConfig(*pw.cfg)
		pw.rebuildServiceList()
	})
	delBtn.Importance = widget.LowImportance

	return container.NewBorder(nil, nil,
		container.NewHBox(dotBox, check),
		container.NewHBox(tag, delBtn),
		lbl,
	)
}

// ── Add dialogs ───────────────────────────────────────────────────────────────

func (pw *portalWindow) showAddServiceDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("e.g. My Service")
	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("e.g. MyServiceID")

	items := []*widget.FormItem{
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Service ID", idEntry),
	}
	dialog.ShowForm("Add Windows Service", "Add", "Cancel", items, func(ok bool) {
		if !ok {
			return
		}
		name := strings.TrimSpace(nameEntry.Text)
		id := strings.TrimSpace(idEntry.Text)
		if name == "" || id == "" {
			return
		}
		pw.cfg.Services = append(pw.cfg.Services, ServiceEntry{Name: name, ID: id, Enabled: true})
		_ = SaveConfig(*pw.cfg)
		pw.rebuildServiceList()
	}, pw.servicesWin)
}

func (pw *portalWindow) showAddProcessDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("e.g. My App")
	exesEntry := widget.NewEntry()
	exesEntry.SetPlaceHolder("e.g. myapp.exe, helper.exe")

	items := []*widget.FormItem{
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Executables", exesEntry),
	}
	dialog.ShowForm("Add Process", "Add", "Cancel", items, func(ok bool) {
		if !ok {
			return
		}
		name := strings.TrimSpace(nameEntry.Text)
		if name == "" {
			return
		}
		var exes []string
		for _, e := range strings.Split(exesEntry.Text, ",") {
			if e = strings.TrimSpace(e); e != "" {
				exes = append(exes, e)
			}
		}
		if len(exes) == 0 {
			return
		}
		pw.cfg.Processes = append(pw.cfg.Processes, ProcessEntry{Name: name, Exes: exes, Enabled: true})
		_ = SaveConfig(*pw.cfg)
		pw.rebuildServiceList()
	}, pw.servicesWin)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func makeDot(col color.Color) (*canvas.Circle, *fyne.Container) {
	dot := canvas.NewCircle(col)
	dot.Resize(fyne.NewSize(10, 10))
	dotBox := container.NewWithoutLayout(dot)
	dotBox.Resize(fyne.NewSize(14, 14))
	dot.Move(fyne.NewPos(2, 2))
	return dot, dotBox
}

// ── State refresh ─────────────────────────────────────────────────────────────

func (pw *portalWindow) refresh() {
	gaming := pw.gm.IsEnabled()

	if gaming {
		pw.statusLabel.SetText("🎮 Gaming Mode: ON")
		pw.toggleBtn.SetText("🛡️  Disable Gaming Mode")
		pw.toggleBtn.Importance = widget.DangerImportance
	} else {
		pw.statusLabel.SetText("🖥️  Gaming Mode: OFF")
		pw.toggleBtn.SetText("⚔️  Enable Gaming Mode")
		pw.toggleBtn.Importance = widget.HighImportance
	}
	pw.toggleBtn.Refresh()

	for _, sr := range pw.serviceItems {
		if gaming {
			sr.dot.FillColor = theme.ErrorColor()
		} else {
			sr.dot.FillColor = theme.SuccessColor()
		}
		sr.dot.Refresh()
	}
}
