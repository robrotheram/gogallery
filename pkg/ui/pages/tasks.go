package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/deploy"
	"gogallery/pkg/monitor"
	"gogallery/pkg/pipeline"
	"gogallery/pkg/preview"
	"gogallery/pkg/ui/components"
	uiMonitor "gogallery/pkg/ui/monitors"
	"sort"
	"sync/atomic"
	"time"
)

type TasksPage struct {
	*datastore.DataStore
	Title            string
	monitor          *uiMonitor.UIMonitor
	OnRescanComplete func()
	table            *fyne.Container
	server           *preview.Server
	content          fyne.CanvasObject
	actions          []*widget.Button
	summary          *widget.Label
	previewButton    *widget.Button
	busy             bool
	active           bool
	updatePending    atomic.Bool
}

func NewTasksPage(db *datastore.DataStore, server *preview.Server) *TasksPage {
	m, ok := db.Monitor.(*uiMonitor.UIMonitor)
	if !ok {
		m = uiMonitor.NewUIMonitor()
		db.Monitor = m
	}
	p := &TasksPage{DataStore: db, Title: "Tasks", monitor: m, server: server}
	p.init()
	m.RegisterListener(func() {
		if !p.updatePending.CompareAndSwap(false, true) {
			return
		}
		fyne.Do(func() {
			p.updatePending.Store(false)
			if p.active {
				p.Refresh()
			}
		})
	})
	return p
}
func (t *TasksPage) action(label string, icon fyne.Resource, run func()) fyne.CanvasObject {
	button := widget.NewButtonWithIcon(label, icon, run)
	t.actions = append(t.actions, button)
	return container.NewThemeOverride(button, taskActionTheme{fyne.CurrentApp().Settings().Theme()})
}
func (t *TasksPage) init() {
	scan := t.action("Scan library", theme.ViewRefreshIcon(), func() {
		t.run("Scan library", t.rescanGallery, func() {
			if t.OnRescanComplete != nil {
				t.OnRescanComplete()
			}
		})
	})
	build := t.action("Build website", theme.ComputerIcon(), func() {
		cfg := config.Config.Gallery
		t.run("Build website", func() error { return pipeline.NewRenderPipeline(&cfg, t.DataStore).BuildSite() }, nil)
	})
	previewButton := t.action("Start preview", theme.VisibilityIcon(), func() {
		if t.server == nil {
			return
		}
		running, _ := t.server.Status()
		if running {
			t.run("Stop preview", t.server.Stop, nil)
		} else {
			t.run("Start preview", t.server.Start, func() { components.NewHeader("", t.server, nil).Preview() })
		}
	})
	t.previewButton = t.actions[len(t.actions)-1]
	publish := t.action("Publish to Netlify", theme.UploadIcon(), func() {
		cfg := *config.Config
		t.run("Publish website", func() error { return deploy.DeploySite(cfg, t.NewTask("Netlify deployment", 1)) }, nil)
	})
	cleanup := t.action("Delete output", theme.DeleteIcon(), func() {
		app := fyne.CurrentApp()
		if app == nil || len(app.Driver().AllWindows()) == 0 {
			return
		}
		dialog.ShowConfirm("Delete generated website?", "This removes the generated output directory. Your source photos will not be deleted.", func(confirmed bool) {
			if confirmed {
				cfg := config.Config.Gallery
				t.run("Delete generated website", func() error { return pipeline.NewRenderPipeline(&cfg, t.DataStore).DeleteSite() }, nil)
			}
		}, app.Driver().AllWindows()[0])
	})
	t.actions[len(t.actions)-1].Importance = widget.DangerImportance
	maintenanceOptions := widget.NewCard("", "", container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("Generated website cleanup", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		components.Hint("Delete the generated output folder to start with a clean build. Your original photos stay in your library."),
		container.NewHBox(cleanup))))
	maintenanceOptions.Hide()
	maintenanceButton := widget.NewButtonWithIcon("Maintenance options", theme.MenuDropDownIcon(), nil)
	maintenanceButton.IconPlacement = widget.ButtonIconTrailingText
	maintenanceButton.OnTapped = func() {
		if maintenanceOptions.Visible() {
			maintenanceOptions.Hide()
			maintenanceButton.SetText("Maintenance options")
			maintenanceButton.SetIcon(theme.MenuDropDownIcon())
		} else {
			maintenanceOptions.Show()
			maintenanceButton.SetText("Hide maintenance options")
			maintenanceButton.SetIcon(theme.MenuDropUpIcon())
		}
		t.content.Refresh()
	}
	t.summary = widget.NewLabel("No tasks yet")
	t.table = container.NewVBox()
	grid := &taskButtonsLayout{}
	actions := container.New(grid, scan, build, previewButton, publish)
	body := container.New(layout.NewCustomPaddedVBoxLayout(16), actions, container.NewHBox(maintenanceButton), maintenanceOptions,
		container.NewVBox(components.NewTextEntry("Activity", 20), t.summary, t.table))
	heading := container.NewVBox(components.NewTextEntry("Tasks", 28), components.Hint("Scan your library, prepare your website, and follow its progress."))
	scroll := container.NewVScroll(container.New(layout.NewCustomPaddedLayout(0, 12, 0, 12), body))
	t.content = components.WorkspaceFrame(container.New(&tasksPageLayout{buttons: grid}, heading, scroll), 1120)
	t.Refresh()
}

// One user-initiated operation at a time prevents duplicate builds or a
// cleanup racing a publish. Every failure is visible in the activity list.
func (t *TasksPage) run(name string, work func() error, done func()) {
	if t.busy {
		return
	}
	t.busy = true
	for _, button := range t.actions {
		button.Disable()
	}
	stat := t.monitor.NewTask(name, 0)
	go func() {
		stat.Start()
		err := work()
		if err != nil {
			stat.Fail(err.Error())
		} else {
			stat.Complete()
		}
		fyne.Do(func() {
			t.busy = false
			for _, button := range t.actions {
				button.Enable()
			}
			if err == nil && done != nil {
				done()
			}
			t.Refresh()
		})
	}()
}

func (t *TasksPage) rescanGallery() error {
	stat := t.monitor.NewTask("Rescan gallery", 2)
	stat.Start()
	if err := t.ScanPath(config.Config.Gallery.Basepath); err != nil {
		stat.Fail(err.Error())
		return err
	}
	stat.Update()
	if err := pipeline.NewRenderPipeline(&config.Config.Gallery, t.DataStore).GenerateThumbnails(); err != nil {
		err = fmt.Errorf("generate thumbnails: %w", err)
		stat.Fail(err.Error())
		return err
	}
	stat.Complete()
	return nil
}

func (t *TasksPage) Refresh() {
	tasks := t.monitor.GetTasks()
	snapshots := make([]monitor.ProgressStats, 0, len(tasks))
	running, completed, failed := 0, 0, 0
	for _, task := range tasks {
		if ps, ok := task.(*monitor.ProgressStats); ok {
			snapshot := ps.Snapshot()
			snapshots = append(snapshots, snapshot)
			if snapshot.State == monitor.RUNNING {
				running++
			}
			if snapshot.State == monitor.COMPLETE {
				completed++
			}
			if snapshot.State == monitor.ERROR {
				failed++
			}
		}
	}
	sort.SliceStable(snapshots, func(i, j int) bool {
		a, b := snapshots[i], snapshots[j]
		if (a.State == monitor.RUNNING) != (b.State == monitor.RUNNING) {
			return a.State == monitor.RUNNING
		}
		return a.StartTime.After(b.StartTime)
	})
	rows := make([]fyne.CanvasObject, 0, len(snapshots))
	for _, snapshot := range snapshots {
		rows = append(rows, t.taskRow(snapshot))
	}
	if len(rows) == 0 {
		rows = append(rows, components.Hint("Activity appears here when you scan, build, preview, or publish."))
	}
	t.table.Objects = rows
	t.summary.SetText(fmt.Sprintf("%d running · %d completed · %d failed", running, completed, failed))
	if len(snapshots) == 0 {
		t.summary.SetText("Ready when you are")
	}
	if t.server != nil {
		running, _ := t.server.Status()
		if running {
			t.previewButton.SetText("Stop preview")
		} else {
			t.previewButton.SetText("Start preview")
		}
	}
	t.content.Refresh()
}
func (t *TasksPage) taskRow(snapshot monitor.ProgressStats) fyne.CanvasObject {
	name := widget.NewLabelWithStyle(snapshot.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	name.Truncation = fyne.TextTruncateEllipsis
	state := taskStatusBadge(t.getTaskStatus(snapshot.State), snapshot.State)
	head := container.NewBorder(nil, nil, nil, state, name)
	body := container.NewVBox(head)
	if snapshot.State == monitor.RUNNING && snapshot.Total > 0 {
		progress := widget.NewProgressBar()
		value := min(1.0, max(0.0, float64(snapshot.Processed)/float64(snapshot.Total)))
		progress.TextFormatter = func() string { return "" }
		progress.SetValue(value)
		// Keep the percentage on the page background so it remains readable
		// across both the filled and unfilled portions of the progress bar.
		body.Add(container.NewBorder(nil, nil, nil, widget.NewLabel(fmt.Sprintf("%.0f%%", value*100)), progress))
	} else if snapshot.State == monitor.RUNNING {
		body.Add(components.Hint("Working…"))
	}
	if snapshot.Message != "" {
		message := components.Hint(snapshot.Message)
		message.Importance = widget.DangerImportance
		body.Add(message)
	}
	if !snapshot.StartTime.IsZero() {
		body.Add(components.Hint(snapshot.StartTime.Format("2 Jan, 15:04") + " · " + t.getTaskTimeTaken(&snapshot)))
	}
	return container.NewVBox(body, widget.NewSeparator())
}
func (t *TasksPage) getTaskStatus(state monitor.ProcessState) string {
	switch state {
	case monitor.COMPLETE:
		return "Complete"
	case monitor.RUNNING:
		return "Running"
	case monitor.ERROR:
		return "Failed"
	case monitor.INACTIVE:
		return "Waiting"
	case monitor.STOPPED:
		return "Stopped"
	}
	return string(state)
}
func (t *TasksPage) getTaskTimeTaken(ps *monitor.ProgressStats) string {
	if !ps.EndTime.IsZero() {
		return ps.Duration.Round(time.Second).String()
	}
	if ps.State == monitor.RUNNING && !ps.StartTime.IsZero() {
		return time.Since(ps.StartTime).Truncate(time.Second).String()
	}
	return "Waiting"
}
func (t *TasksPage) Layout() fyne.CanvasObject { t.active = true; t.Refresh(); return t.content }

func (t *TasksPage) SetActive(active bool) { t.active = active }
