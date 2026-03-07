package main

import (
	"fmt"
	"image/color"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// portalWindow holds references to all live UI elements so they can be
// refreshed when the gaming mode state changes from outside the window
// (e.g. tray icon left-click).
type portalWindow struct {
	win          fyne.Window
	gm           *GamingMode
	cfg          *Config
	statusLabel  *widget.Label
	toggleBtn    *widget.Button
	serviceItems []*serviceRow
}

// serviceRow is a single row in the managed-services list.
type serviceRow struct {
	dot   *canvas.Circle
	label *widget.Label
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
	pw.win.Resize(fyne.NewSize(700, 560))
	pw.win.SetFixedSize(false)

	// ── Header ────────────────────────────────────────────────────────────────
	title := widget.NewLabelWithStyle("🛡️  Warden", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.TextStyle.Bold = true
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

	gamingCard := widget.NewCard("Gaming Mode", "", container.NewVBox(desc, pw.toggleBtn))

	// ── Managed Services Card ─────────────────────────────────────────────────
	pw.serviceItems = make([]*serviceRow, 0, len(pw.cfg.Services)+len(pw.cfg.Processes))

	serviceRows := container.NewVBox()

	addSection := func(heading string) {
		lbl := widget.NewLabelWithStyle(heading, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		serviceRows.Add(lbl)
	}

	addSection("Windows Services")
	for _, svc := range pw.cfg.Services {
		row := pw.makeServiceRow(svc.Name, svc.Enabled)
		serviceRows.Add(row.container)
	}

	addSection("Processes")
	for _, proc := range pw.cfg.Processes {
		row := pw.makeProcRow(proc.Name, proc.Exes, proc.Enabled)
		serviceRows.Add(row.container)
	}

	servicesCard := widget.NewCard("Managed Services", "Stopped when gaming mode is active.", container.NewScroll(serviceRows))

	// ── Hotlinks Card ─────────────────────────────────────────────────────────
	hotlinksGrid := container.NewGridWithColumns(3)
	for _, link := range pw.cfg.Hotlinks {
		link := link // capture loop var
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
		servicesCard,
		hotlinksCard,
	)

	scroll := container.NewScroll(content)
	pw.win.SetContent(container.NewPadded(scroll))
	pw.win.Show()
}

// ── Row builders ──────────────────────────────────────────────────────────────

type namedRow struct {
	container *fyne.Container
	dot       *canvas.Circle
}

func makeDot(col color.Color) (*canvas.Circle, *fyne.Container) {
	dot := canvas.NewCircle(col)
	dot.Resize(fyne.NewSize(10, 10))
	// Wrap in a fixed-size container so the circle gets proper layout space.
	dotBox := container.NewWithoutLayout(dot)
	dotBox.Resize(fyne.NewSize(14, 14))
	dot.Move(fyne.NewPos(2, 2))
	return dot, dotBox
}

func (pw *portalWindow) makeServiceRow(name string, enabled bool) namedRow {
	dot, dotBox := makeDot(theme.SuccessColor())

	lbl := widget.NewLabel(name)
	tag := widget.NewLabel("service")
	tag.TextStyle = fyne.TextStyle{Italic: true}

	row := namedRow{
		dot:       dot,
		container: container.NewBorder(nil, nil, dotBox, tag, lbl),
	}

	sr := &serviceRow{dot: dot, label: lbl}
	if !enabled {
		sr.label.TextStyle = fyne.TextStyle{Italic: true}
	}
	pw.serviceItems = append(pw.serviceItems, sr)
	return row
}

func (pw *portalWindow) makeProcRow(name string, exes []string, enabled bool) namedRow {
	dot, dotBox := makeDot(theme.SuccessColor())

	label := name
	if len(exes) > 1 {
		label = fmt.Sprintf("%s (%d exes)", name, len(exes))
	}
	lbl := widget.NewLabel(label)
	tag := widget.NewLabel("process")
	tag.TextStyle = fyne.TextStyle{Italic: true}

	row := namedRow{
		dot:       dot,
		container: container.NewBorder(nil, nil, dotBox, tag, lbl),
	}

	sr := &serviceRow{dot: dot, label: lbl}
	if !enabled {
		sr.label.TextStyle = fyne.TextStyle{Italic: true}
	}
	pw.serviceItems = append(pw.serviceItems, sr)
	return row
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

	// Update dot colours for every service/process row
	for _, sr := range pw.serviceItems {
		if gaming {
			sr.dot.FillColor = theme.ErrorColor()
		} else {
			sr.dot.FillColor = theme.SuccessColor()
		}
		sr.dot.Refresh()
	}
}
