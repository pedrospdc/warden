package main

import (
	"fmt"
	"image/color"
	"net/url"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// portalWindow holds references to all live UI elements so they can be
// refreshed when the gaming mode state changes from outside the window
// (e.g. tray icon left-click).
type portalWindow struct {
	app            fyne.App
	win            fyne.Window
	servicesWin    fyne.Window
	gm             *GamingMode
	cfg            *Config
	statusLabel    *widget.Label
	toggleBtn      *widget.Button
	servicesBox    *fyne.Container
	serviceItems   []*serviceRow
	cachedServices []ServiceInfo
	servicesMu     sync.Mutex
}

// serviceRow tracks the status dot for a single managed entry so refresh()
// can update its colour without rebuilding the whole list.
type serviceRow struct {
	dot *canvas.Circle
}

// NewPortalWindow creates the portal window without showing it.
// Call pw.win.Show() to display it.
func NewPortalWindow(a fyne.App, gm *GamingMode, cfg *Config) *portalWindow {
	pw := &portalWindow{app: a, gm: gm, cfg: cfg}
	pw.build()
	pw.refresh()

	gm.OnChange(func(_ bool) {
		pw.refresh()
	})
	return pw
}

func (pw *portalWindow) build() {
	pw.win = pw.app.NewWindow("Warden – Portal")
	pw.win.Resize(fyne.NewSize(700, 600))
	pw.win.SetFixedSize(false)
	// Hide instead of close so the tray can bring it back.
	pw.win.SetCloseIntercept(func() { pw.win.Hide() })

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
		pw.showManageServicesWindow()
	})
	gamingCard := widget.NewCard("Gaming Mode", "", container.NewVBox(desc, pw.toggleBtn, manageBtn))

	// ── Hotlinks Card ─────────────────────────────────────────────────────────
	hotlinksGrid := container.NewGridWithColumns(3)
	for _, link := range pw.cfg.Hotlinks {
		link := link
		btn := widget.NewButton(fmt.Sprintf("%s  %s", link.Icon, link.Name), func() {
			parsed, err := url.Parse(link.URL)
			if err == nil {
				pw.app.OpenURL(parsed)
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
}

// showManageServicesWindow opens a dedicated window for managing services and processes.
// The Windows services list is fetched once in the background on first open and
// cached for all subsequent add/edit dialogs.
func (pw *portalWindow) showManageServicesWindow() {
	if pw.servicesWin != nil {
		pw.servicesWin.RequestFocus()
		return
	}

	// Load the Windows services list in the background the first time only.
	pw.servicesMu.Lock()
	needsLoad := pw.cachedServices == nil
	pw.servicesMu.Unlock()
	if needsLoad {
		go func() {
			svcs := ListWindowsServices()
			pw.servicesMu.Lock()
			pw.cachedServices = svcs
			pw.servicesMu.Unlock()
		}()
	}

	pw.servicesBox = container.NewVBox()
	pw.rebuildServiceList()

	pw.servicesWin = pw.app.NewWindow("Warden – Manage Services")
	pw.servicesWin.Resize(fyne.NewSize(540, 600))
	pw.servicesWin.SetContent(container.NewPadded(container.NewScroll(pw.servicesBox)))
	pw.servicesWin.SetOnClosed(func() { pw.servicesWin = nil })
	pw.servicesWin.Show()
}

// rebuildServiceList clears and regenerates the services/processes VBox.
// Call this after any mutation (add, delete, toggle, edit) to keep the list in sync.
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
	addSvcBtn := widget.NewButton("+ Add Service", func() {
		pw.showServiceDialog("Add Windows Service", "Add", "", "", func(name, id string) {
			pw.cfg.Services = append(pw.cfg.Services, ServiceEntry{Name: name, ID: id, Enabled: true})
			_ = SaveConfig(*pw.cfg)
			pw.rebuildServiceList()
		})
	})
	addSvcBtn.Importance = widget.LowImportance
	pw.servicesBox.Add(addSvcBtn)

	// ── Processes ─────────────────────────────────────────────────────────────
	pw.servicesBox.Add(widget.NewSeparator())
	pw.servicesBox.Add(widget.NewLabelWithStyle("Processes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	for i := range pw.cfg.Processes {
		pw.servicesBox.Add(pw.makeProcRow(i))
	}
	addProcBtn := widget.NewButton("+ Add Process", func() {
		pw.showProcessDialog("Add Process", "Add", "", nil, func(name string, exes []string) {
			pw.cfg.Processes = append(pw.cfg.Processes, ProcessEntry{Name: name, Exes: exes, Enabled: true})
			_ = SaveConfig(*pw.cfg)
			pw.rebuildServiceList()
		})
	})
	addProcBtn.Importance = widget.LowImportance
	pw.servicesBox.Add(addProcBtn)

	pw.servicesBox.Refresh()
}

func (pw *portalWindow) makeServiceRow(idx int) fyne.CanvasObject {
	svc := &pw.cfg.Services[idx]

	dot, dotBox := makeDot(theme.SuccessColor())
	pw.serviceItems = append(pw.serviceItems, &serviceRow{dot: dot})

	check := widget.NewCheck("", nil)
	check.SetChecked(svc.Enabled)
	check.OnChanged = func(checked bool) {
		pw.cfg.Services[idx].Enabled = checked
		_ = SaveConfig(*pw.cfg)
	}

	lbl := widget.NewLabel(svc.Name)

	tag := widget.NewLabel("service")
	tag.TextStyle = fyne.TextStyle{Italic: true}

	editBtn := widget.NewButton("✎", func() {
		pw.showServiceDialog("Edit Windows Service", "Save", pw.cfg.Services[idx].Name, pw.cfg.Services[idx].ID, func(name, id string) {
			pw.cfg.Services[idx].Name = name
			pw.cfg.Services[idx].ID = id
			_ = SaveConfig(*pw.cfg)
			pw.rebuildServiceList()
		})
	})
	editBtn.Importance = widget.LowImportance

	delBtn := widget.NewButton("✕", func() {
		pw.cfg.Services = append(pw.cfg.Services[:idx], pw.cfg.Services[idx+1:]...)
		_ = SaveConfig(*pw.cfg)
		pw.rebuildServiceList()
	})
	delBtn.Importance = widget.LowImportance

	return container.NewBorder(nil, nil,
		container.NewHBox(dotBox, check),
		container.NewHBox(tag, editBtn, delBtn),
		lbl,
	)
}

func (pw *portalWindow) makeProcRow(idx int) fyne.CanvasObject {
	proc := &pw.cfg.Processes[idx]

	dot, dotBox := makeDot(theme.SuccessColor())
	pw.serviceItems = append(pw.serviceItems, &serviceRow{dot: dot})

	check := widget.NewCheck("", nil)
	check.SetChecked(proc.Enabled)
	check.OnChanged = func(checked bool) {
		pw.cfg.Processes[idx].Enabled = checked
		_ = SaveConfig(*pw.cfg)
	}

	label := proc.Name
	if len(proc.Exes) > 1 {
		label = fmt.Sprintf("%s (%d exes)", proc.Name, len(proc.Exes))
	}
	lbl := widget.NewLabel(label)

	tag := widget.NewLabel("process")
	tag.TextStyle = fyne.TextStyle{Italic: true}

	editBtn := widget.NewButton("✎", func() {
		pw.showProcessDialog("Edit Process", "Save", pw.cfg.Processes[idx].Name, pw.cfg.Processes[idx].Exes, func(name string, exes []string) {
			pw.cfg.Processes[idx].Name = name
			pw.cfg.Processes[idx].Exes = exes
			_ = SaveConfig(*pw.cfg)
			pw.rebuildServiceList()
		})
	})
	editBtn.Importance = widget.LowImportance

	delBtn := widget.NewButton("✕", func() {
		pw.cfg.Processes = append(pw.cfg.Processes[:idx], pw.cfg.Processes[idx+1:]...)
		_ = SaveConfig(*pw.cfg)
		pw.rebuildServiceList()
	})
	delBtn.Importance = widget.LowImportance

	return container.NewBorder(nil, nil,
		container.NewHBox(dotBox, check),
		container.NewHBox(tag, editBtn, delBtn),
		lbl,
	)
}

// ── Dialogs ───────────────────────────────────────────────────────────────────

// showServiceDialog opens a full window for adding/editing a Windows service.
// When the cached services list is available a fuzzy-search picker (Entry +
// List) occupies the top portion, filling all available space.
func (pw *portalWindow) showServiceDialog(title, confirm, initName, initID string, onSave func(name, id string)) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(initName)
	nameEntry.SetPlaceHolder("e.g. My Service")

	idEntry := widget.NewEntry()
	idEntry.SetText(initID)
	idEntry.SetPlaceHolder("e.g. MyServiceID")

	cancelBtn := widget.NewButton("Cancel", nil)
	saveBtn := widget.NewButton(confirm, nil)
	saveBtn.Importance = widget.HighImportance

	pw.servicesMu.Lock()
	svcs := pw.cachedServices
	pw.servicesMu.Unlock()

	win := pw.app.NewWindow(title)

	cancelBtn.OnTapped = func() { win.Close() }
	saveBtn.OnTapped = func() {
		name := strings.TrimSpace(nameEntry.Text)
		id := strings.TrimSpace(idEntry.Text)
		if name == "" || id == "" {
			return
		}
		onSave(name, id)
		win.Close()
	}

	form := widget.NewForm(
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Service ID", idEntry),
	)
	buttons := container.NewGridWithColumns(2, cancelBtn, saveBtn)
	bottom := container.NewVBox(widget.NewSeparator(), form, buttons)

	if len(svcs) > 0 {
		filtered := make([]ServiceInfo, len(svcs))
		copy(filtered, svcs)

		searchEntry := widget.NewEntry()
		searchEntry.SetPlaceHolder("Search services…")

		list := widget.NewList(
			func() int { return len(filtered) },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				s := filtered[id]
				obj.(*widget.Label).SetText(s.DisplayName + " (" + s.ID + ")")
			},
		)
		list.OnSelected = func(id widget.ListItemID) {
			nameEntry.SetText(filtered[id].DisplayName)
			idEntry.SetText(filtered[id].ID)
		}
		searchEntry.OnChanged = func(query string) {
			filtered = filtered[:0]
			for _, s := range svcs {
				if fuzzyMatch(query, s.DisplayName) || fuzzyMatch(query, s.ID) {
					filtered = append(filtered, s)
				}
			}
			list.Refresh()
		}

		win.Resize(fyne.NewSize(500, 540))
		win.SetContent(container.NewPadded(
			container.NewBorder(searchEntry, bottom, nil, nil, list),
		))
	} else {
		win.Resize(fyne.NewSize(420, 220))
		win.SetContent(container.NewPadded(container.NewVBox(form, buttons)))
	}
	win.Show()
}

// showProcessDialog opens an add/edit dialog for a process entry.
// A Browse button lets the user navigate to an .exe via a file picker;
// the selected filename is appended to the executables field.
func (pw *portalWindow) showProcessDialog(title, confirm, initName string, initExes []string, onSave func(name string, exes []string)) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(initName)
	nameEntry.SetPlaceHolder("e.g. My App")

	exesEntry := widget.NewEntry()
	exesEntry.SetText(strings.Join(initExes, ", "))
	exesEntry.SetPlaceHolder("e.g. myapp.exe, helper.exe")

	browseBtn := widget.NewButton("Browse…", func() {
		fd := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			defer r.Close()
			filename := r.URI().Name()
			current := strings.TrimSpace(exesEntry.Text)
			if current == "" {
				exesEntry.SetText(filename)
			} else {
				exesEntry.SetText(current + ", " + filename)
			}
		}, pw.servicesWin)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
		fd.Show()
	})

	exesRow := container.NewBorder(nil, nil, nil, browseBtn, exesEntry)

	items := []*widget.FormItem{
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Executables", exesRow),
	}

	d := dialog.NewForm(title, confirm, "Cancel", items, func(ok bool) {
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
		onSave(name, exes)
	}, pw.servicesWin)
	d.Resize(fyne.NewSize(460, 210))
	d.Show()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// fuzzyMatch returns true if every rune of pattern appears in text in order,
// case-insensitively — the same subsequence algorithm VS Code uses.
func fuzzyMatch(pattern, text string) bool {
	if pattern == "" {
		return true
	}
	text = strings.ToLower(text)
	pi := 0
	pr := []rune(strings.ToLower(pattern))
	for _, c := range text {
		if c == pr[pi] {
			pi++
			if pi == len(pr) {
				return true
			}
		}
	}
	return false
}

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
