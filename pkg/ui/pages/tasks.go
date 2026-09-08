package pages

import (
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/deploy"
	"gogallery/pkg/monitor"
	"gogallery/pkg/pipeline"
	"gogallery/pkg/preview"
	"log"
	"time"

	uiMonitor "gogallery/pkg/ui/monitors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

func NewTasksPage(db *datastore.DataStore, server *preview.Server) *TasksPage {
	uiTaskMonitor, ok := db.Monitor.(*uiMonitor.UIMonitor)
	if !ok {
		uiTaskMonitor = uiMonitor.NewUIMonitor()
		db.Monitor = uiTaskMonitor
	}
	page := &TasksPage{
		Title:     "Tasks",
		DataStore: db,
		server:    server,
		monitor:   uiTaskMonitor,
		table:     nil, // Will be initialized in Layout
	}
	page.init() // Initialize buttons and table

	uiTaskMonitor.RegisterListener(func() {
		fyne.Do(func() {
			page.Refresh() // Refresh the page when tasks are updated
		})
	})
	return page
}

func (t *TasksPage) init() {
	// --- Action Buttons ---
	rescanBtn := widget.NewButton("Rescan", func() {
		go func() {
			if err := t.ScanPath(config.Config.Gallery.Basepath); err != nil {
				log.Printf("Gallery scan skipped: %v", err)
			}
		}()
	})
	deleteSite := func() {
		stat := t.monitor.NewTask("Delete Site", 0)
		go func() {
			stat.Start()
			if err := pipeline.NewRenderPipeline(&config.Config.Gallery, t.DataStore).DeleteSite(); err != nil {
				stat.Fail(err.Error())
				log.Printf("Could not delete generated site: %v", err)
				return
			}
			stat.Complete()
		}()
	}
	deleteBtn := widget.NewButton("Delete Site", func() {
		app := fyne.CurrentApp()
		if app == nil || len(app.Driver().AllWindows()) == 0 {
			log.Print("Could not show delete confirmation: application window unavailable")
			return
		}
		dialog.ShowConfirm(
			"Delete generated site?",
			"This removes the generated output directory. Your source photos will not be deleted.",
			func(confirmed bool) {
				if confirmed {
					deleteSite()
				}
			},
			app.Driver().AllWindows()[0],
		)
	})
	buildBtn := widget.NewButton("Build Site", func() {
		go func() {
			if err := pipeline.NewRenderPipeline(&config.Config.Gallery, t.DataStore).BuildSite(); err != nil {
				log.Printf("Could not build site: %v", err)
			}
		}()
	})
	deployBtn := widget.NewButton("Deploy Site", func() {
		go func() {
			if err := deploy.DeploySite(*config.Config, t.NewTask("Netlify deployment", 1)); err != nil {
				log.Printf("Could not deploy site: %v", err)
			}
		}()
	})

	startServerBtn := widget.NewButton("Start Preview Server", func() {
		go func() {
			if err := t.server.Start(); err != nil {
				log.Printf("Could not start preview server: %v", err)
			}
		}()
	})
	stopServerBtn := widget.NewButton("Stop Preview Server", func() {
		go func() {
			if err := t.server.Stop(); err != nil {
				log.Printf("Could not stop preview server: %v", err)
			}
		}()
	})

	// Button grid in a centered, fixed-width box
	t.btnGrid = container.NewGridWithColumns(2,
		rescanBtn, deleteBtn,
		buildBtn, deployBtn,
		startServerBtn, stopServerBtn,
	)

	t.table = createTable(nil)
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

func createTable(r []fyne.CanvasObject) *fyne.Container {
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
		snapshot := ps.Snapshot()
		name = snapshot.Name
		status = t.getTaskStatus(snapshot.State)
		startedAt = t.getTaskStartTime(&snapshot)
		timeTaken = t.getTaskTimeTaken(&snapshot)
		if snapshot.Total == 0 {
			percent = 1
		} else {
			percent = float64(snapshot.Processed) / float64(snapshot.Total)
		}
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

func (t *TasksPage) getTaskStatus(state monitor.ProcessState) string {
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
