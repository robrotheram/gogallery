package ui

import (
	"fmt"
	"testingFyne/pkg/config"
	"testingFyne/pkg/datastore"
	"testingFyne/pkg/pipeline"
	"testingFyne/pkg/preview"
	"testingFyne/pkg/ui/components"
	"testingFyne/pkg/ui/pages"

	uiMonitor "testingFyne/pkg/ui/monitors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func App() error {
	myApp := app.New()
	myWindow := myApp.NewWindow("GoGallery")

	cfg := config.LoadConfig()

	monitor := uiMonitor.NewUIMonitor()
	db, err := datastore.Open("gogallery.sql.db", monitor)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return err
	}
	server := preview.NewServer(db)
	// Start background task to scan path and generate thumbnails
	go backgroundTask(db, cfg)

	galleryPage := pages.NewGalleryPage(db)
	settingsPage := pages.NewSettingsPage(db)
	tasksPage := pages.NewTasksPage(db, server)

	pages := map[string]pages.Page{
		"Gallery":  galleryPage,
		"Settings": settingsPage,
		"Tasks":    tasksPage,
	}

	var navBar *fyne.Container

	setPage := func(page string) {
		currentPage, ok := pages[page]
		if !ok {
			fmt.Println("Page not found:", page)
			return
		}
		content := container.NewBorder(
			navBar, nil, nil, nil,
			container.NewStack(currentPage.Layout()),
		)
		myWindow.SetContent(content)
	}
	navBar = (components.NewHeader("Gallery", db, server, setPage)).Layout()
	setPage("Gallery")
	myApp.Settings().SetTheme(NewComfortableTheme(cfg.UI.Theme))
	myWindow.Resize(fyne.NewSize(1200, 800))
	myWindow.ShowAndRun()
	return nil
}

func backgroundTask(db *datastore.DataStore, cfg *config.Configuration) {
	fmt.Println("Starting background task to scan path:", cfg.Gallery.Basepath)
	db.ScanPath(cfg.Gallery.Basepath)
	pipeline := pipeline.NewRenderPipeline(&cfg.Gallery, db)
	pipeline.GenTumbnails()
	fmt.Println("Background task completed")
}
