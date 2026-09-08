package ui

import (
	"fmt"
	"gogallery/pkg/ai"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/pipeline"
	"gogallery/pkg/preview"
	"gogallery/pkg/ui/components"
	"gogallery/pkg/ui/pages"

	uiMonitor "gogallery/pkg/ui/monitors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func App() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	myApp := app.New()
	myApp.Settings().SetTheme(NewComfortableTheme(cfg.UI.Theme))
	myWindow := myApp.NewWindow("GoGallery")
	myWindow.SetMaster()
	installCloseShortcut(myWindow, myApp.Driver().Quit)

	monitor := uiMonitor.NewUIMonitor()
	db, err := datastore.Open(datastore.DefaultDatabasePath, monitor)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return err
	}
	defer db.Close()
	if cfg.UI.GeminiApiKey != "" {
		if _, err := ai.RegisterGeminiClient(); err != nil {
			fmt.Println("Could not configure Gemini:", err)
		}
	}
	server := preview.NewServer(db)
	defer server.Stop()

	galleryPage := pages.NewGalleryPage(db)
	settingsPage := pages.NewSettingsPage(db, server)
	collectionPage := pages.NewCollectionPage(db)
	tasksPage := pages.NewTasksPage(db, server)

	pages := map[string]pages.Page{
		"Gallery":     galleryPage,
		"Settings":    settingsPage,
		"Tasks":       tasksPage,
		"Collections": collectionPage,
	}

	var navBar *fyne.Container
	var header *components.Header

	setPage := func(page string) {
		currentPage, ok := pages[page]
		if !ok {
			fmt.Println("Page not found:", page)
			return
		}
		if header != nil {
			header.SetActive(page)
		}
		content := container.NewBorder(
			navBar, nil, nil, nil,
			container.NewStack(currentPage.Layout()),
		)
		myWindow.SetContent(content)
	}
	header = components.NewHeader(config.Config.Gallery.Name, server, setPage)
	navBar = header.Layout()
	setPage("Gallery")

	// Start the scan after pages are ready so completion can refresh the views.
	go backgroundTask(db, cfg, func() {
		fyne.Do(func() {
			galleryPage.Refresh()
			collectionPage.Refresh()
		})
	})

	myWindow.Resize(fyne.NewSize(1200, 800))
	appStopped := make(chan struct{})
	stopSignalHandler := installShutdownHandler(myApp.Driver().Quit, appStopped)
	defer func() {
		close(appStopped)
		stopSignalHandler()
	}()
	myWindow.ShowAndRun()
	return nil
}

func backgroundTask(db *datastore.DataStore, cfg *config.Configuration, onComplete func()) {
	fmt.Println("Starting background task to scan path:", cfg.Gallery.Basepath)
	if err := db.ScanPath(cfg.Gallery.Basepath); err != nil {
		fmt.Println("Initial gallery scan skipped:", err)
		return
	}
	if onComplete != nil {
		defer onComplete()
	}
	pipeline := pipeline.NewRenderPipeline(&cfg.Gallery, db)
	if err := pipeline.GenerateThumbnails(); err != nil {
		fmt.Println("Thumbnail generation failed:", err)
		return
	}
	fmt.Println("Background task completed")
}
