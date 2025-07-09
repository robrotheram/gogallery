package pages

import (
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/deploy"
	"gogallery/pkg/monitor"
	"gogallery/pkg/pipeline"
	"gogallery/pkg/preview"
	"time"

	uiMonitor "gogallery/pkg/ui/monitors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type TasksPage struct {
	*datastore.DataStore
	Title   string
	monitor *uiMonitor.UIMonitor
	table   *fyne.Container
	btnGrid *fyne.Container // Action buttons grid
	server  *preview.Server
}

var cfg = config.Config

func NewTasksPage(db *datastore.DataStore, server *preview.Server) *TasksPage {
	uiMonitor, ok := db.Monitor.(*uiMonitor.UIMonitor)
	if !ok {
		panic("db.Monitor is not of type UIMonitor")
	}
	page := &TasksPage{
		Title:     "Tasks",
		DataStore: db,
		server:    server,
		monitor:   uiMonitor, // Ensure db.Monitor is of type UIMonitor
		table:     nil,       // Will be initialized in Layout
	}
	page.init() // Initialize buttons and table

	uiMonitor.RegisterListener(func() {
		fyne.Do(func() {
			page.Refresh() // Refresh the page when tasks are updated
		})
	})
	return page
}

func (t *TasksPage) init() {
	// --- Action Buttons ---
	rescanBtn := widget.NewButton("Rescan", func() {
		go t.ScanPath(cfg.Gallery.Basepath)
	})
	deleteBtn := widget.NewButton("Delete Site", func() {
		stat := t.monitor.NewTask("Delete Site", 0)
		go func() {
			stat.Start()
			defer stat.Complete()
			pipeline.NewRenderPipeline(&cfg.Gallery, t.DataStore).DeleteSite()
			t.DataStore.Reset() // Reset datastore after deletion
		}()
	})
	buildBtn := widget.NewButton("Build Site", func() {
		go pipeline.NewRenderPipeline(&cfg.Gallery, t.DataStore).BuildSite()
	})
	deployBtn := widget.NewButton("Deploy Site", func() {
		go deploy.DeploySite(*cfg, t.NewTask("netify deploy", 1))
	})

	startServerBtn := widget.NewButton("Start Preview Server", func() {
		go t.server.Start()
	})
	stopServerBtn := widget.NewButton("Stop Preview Server", func() {
		go t.server.Stop()
	})

	// Button grid in a centered, fixed-width box
	t.btnGrid = container.NewGridWithColumns(2,
		rescanBtn, deleteBtn,
		buildBtn, deployBtn,
		startServerBtn, stopServerBtn,
	)

	t.table = craeteTable([]fyne.CanvasObject{})
}

func tableHeader() fyne.CanvasObject {
	return container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("Task Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Started At", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Time Taken", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Progress", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
}

func craeteTable(r []fyne.CanvasObject) *fyne.Container {
	rows := []fyne.CanvasObject{
		tableHeader(),
	}
	rows = append(rows, r...)
	return container.NewVBox(rows...)
}

func (t *TasksPage) Refresh() {
	tasks := t.monitor.GetTasks()
	rows := []fyne.CanvasObject{
		tableHeader(),
	}
	for _, task := range tasks {
		rows = append(rows, t.createTaskRow(task))
	}
	t.table.Objects = rows
	t.table.Refresh() // Refresh the table to show new rows
}

func (t *TasksPage) createTaskRow(task interface{}) fyne.CanvasObject {
	var (
		name, status, startedAt, timeTaken string
		percent                            float64
	)
	if ps, ok := task.(*monitor.ProgressStats); ok {
		name = ps.Name
		status = t.getTaskStatus(ps.State)
		startedAt = t.getTaskStartTime(ps)
		timeTaken = t.getTaskTimeTaken(ps)
		percent = ps.Percent() / 100.0
	} else {
		name = "Unknown"
		status = "-"
		startedAt = "-"
		timeTaken = "-"
		percent = 0
	}
	progress := widget.NewProgressBar()
	progress.SetValue(percent)
	return container.NewGridWithColumns(5,
		widget.NewLabel(name),
		widget.NewLabel(status),
		widget.NewLabel(startedAt),
		widget.NewLabel(timeTaken),
		progress,
	)
}

func (t *TasksPage) getTaskStatus(state monitor.ProssesState) string {
	switch state {
	case monitor.COMPLETE:
		return "Complete"
	case monitor.RUNNING:
		return "In Progress"
	case monitor.ERROR:
		return "Error"
	default:
		return string(state)
	}
}

func (t *TasksPage) getTaskStartTime(ps *monitor.ProgressStats) string {
	if !ps.StartTime.IsZero() {
		return ps.StartTime.Format("15:04:05")
	}
	return "-"
}

func (t *TasksPage) getTaskTimeTaken(ps *monitor.ProgressStats) string {
	if ps.State == monitor.COMPLETE && !ps.EndTime.IsZero() {
		return ps.Duration.String()
	} else if ps.State == monitor.RUNNING && !ps.StartTime.IsZero() {
		return time.Since(ps.StartTime).Truncate(time.Second).String()
	}
	return "-"
}

func (t *TasksPage) Layout() fyne.CanvasObject {

	content := container.NewBorder(
		container.NewPadded(t.btnGrid),
		nil,
		nil, nil,
		container.NewPadded(container.NewVScroll(t.table)),
	)
	return content
}
