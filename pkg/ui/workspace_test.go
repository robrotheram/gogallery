package ui

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/preview"
	"gogallery/pkg/ui/components"
	uiMonitor "gogallery/pkg/ui/monitors"
	"gogallery/pkg/ui/pages"
)

// The production driver serializes fyne.Do calls. Reproduce that contract in
// these interaction tests instead of letting the test driver execute UI work
// directly on thumbnail/database goroutines.
type queuedDriver struct {
	fyne.Driver
	pending chan func()
}

func (d *queuedDriver) DoFromGoroutine(fn func(), wait bool) {
	if !wait {
		d.pending <- fn
		return
	}
	done := make(chan struct{})
	d.pending <- func() { fn(); close(done) }
	<-done
}

type queuedApp struct {
	fyne.App
	driver *queuedDriver
}

func (a *queuedApp) Driver() fyne.Driver { return a.driver }
func (d *queuedDriver) pump() {
	for {
		select {
		case f := <-d.pending:
			f()
		default:
			return
		}
	}
}
func (d *queuedDriver) until(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		d.pump()
		if ready() {
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
	t.Fatal("UI operation did not finish")
}

func descendants(root fyne.CanvasObject) []fyne.CanvasObject {
	objects := []fyne.CanvasObject{}
	seen := map[fyne.CanvasObject]bool{}
	var visit func(fyne.CanvasObject)
	visit = func(o fyne.CanvasObject) {
		if o == nil || seen[o] || !o.Visible() {
			return
		}
		seen[o] = true
		objects = append(objects, o)
		if c, ok := o.(*fyne.Container); ok {
			for _, child := range c.Objects {
				visit(child)
			}
		} else if w, ok := o.(fyne.Widget); ok {
			for _, child := range test.WidgetRenderer(w).Objects() {
				visit(child)
			}
		}
	}
	visit(root)
	return objects
}
func buttonNamed(t *testing.T, root fyne.CanvasObject, name string) *widget.Button {
	t.Helper()
	for _, o := range descendants(root) {
		if b, ok := o.(*widget.Button); ok && b.Text == name {
			return b
		}
	}
	t.Fatalf("button %q not found", name)
	return nil
}
func entryPlaceholder(t *testing.T, root fyne.CanvasObject, placeholder string) *widget.Entry {
	t.Helper()
	for _, o := range descendants(root) {
		if e, ok := o.(*widget.Entry); ok && e.PlaceHolder == placeholder {
			return e
		}
	}
	t.Fatalf("entry %q not found", placeholder)
	return nil
}
func hasLabel(root fyne.CanvasObject, text string) bool {
	for _, o := range descendants(root) {
		if l, ok := o.(*widget.Label); ok && l.Text == text {
			return true
		}
	}
	return false
}
func firstPhoto(t *testing.T, root fyne.CanvasObject) *components.Image {
	t.Helper()
	for _, o := range descendants(root) {
		if p, ok := o.(*components.Image); ok {
			return p
		}
	}
	t.Fatal("no photo tile found")
	return nil
}
func albumCoverNamed(t *testing.T, root fyne.CanvasObject, name string) *components.Image {
	t.Helper()
	for _, object := range descendants(root) {
		card, ok := object.(*fyne.Container)
		if !ok {
			continue
		}
		for _, child := range card.Objects {
			if image, ok := child.(*components.Image); ok && hasLabel(card, name) {
				return image
			}
		}
	}
	t.Fatalf("album card %q not found", name)
	return nil
}
func lastPhoto(t *testing.T, root fyne.CanvasObject) *components.Image {
	t.Helper()
	var last *components.Image
	for _, object := range descendants(root) {
		if photo, ok := object.(*components.Image); ok {
			last = photo
		}
	}
	if last == nil {
		t.Fatal("no photo found")
	}
	return last
}
func captureWorkspace(t *testing.T, window fyne.Window, name string) {
	t.Helper()
	dir := os.Getenv("GOGALLERY_UI_SCREENSHOTS")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err = png.Encode(file, window.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}

func TestPhotoWorkspaceInteractions(t *testing.T) {
	base := test.NewApp()
	driver := &queuedDriver{Driver: base.Driver(), pending: make(chan func(), 2048)}
	app := &queuedApp{App: base, driver: driver}
	fyne.SetCurrentApp(app)
	t.Cleanup(base.Quit)
	app.Settings().SetTheme(NewComfortableTheme("dark"))
	previous := config.Config
	root := t.TempDir()
	config.Config = &config.Configuration{Gallery: config.GalleryConfiguration{Name: "GoGallery", Basepath: root, Theme: "EmeraldNoir", ImagesPerPage: 24}, UI: config.UIConfiguration{Theme: "dark", ImagesPerPage: 12}}
	t.Cleanup(func() { config.Config = previous })
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	db, err := datastore.Open(filepath.Join(t.TempDir(), "gallery.db"), uiMonitor.NewUIMonitor())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	albums := []datastore.Album{
		{Id: "coast", Name: "Coastal walks", Path: filepath.Join(root, "Coastal walks")},
		{Id: "year25", Name: "2025", Path: filepath.Join(root, "2025")},
		{Id: "year26", Name: "2026", Path: filepath.Join(root, "2026")},
		{Id: "macro", Name: "Close-ups", Path: filepath.Join(root, "2026", "Garden", "Macro")},
		{Id: "garden", Name: "Garden", Path: filepath.Join(root, "2025", "Garden")},
		{Id: "garden2", Name: "Garden", Path: filepath.Join(root, "2026", "Garden")},
		{Id: "empty", Name: "An empty album", Path: filepath.Join(root, "Empty")},
	}
	for _, album := range albums {
		if err := db.Albums.Save(album); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 18; i++ {
		path := filepath.Join(root, fmt.Sprintf("photo-%02d.jpg", i))
		photo := image.NewRGBA(image.Rect(0, 0, 300, 200))
		for y := 0; y < 200; y++ {
			for x := 0; x < 300; x++ {
				photo.Set(x, y, color.NRGBA{uint8(30 + i*10), uint8(60 + y/2), uint8(90 + x/3), 255})
			}
		}
		file, _ := os.Create(path)
		if err := jpeg.Encode(file, photo, nil); err != nil {
			t.Fatal(err)
		}
		file.Close()
		// Optional real photos are read only for local visual review; tests remain
		// self-contained and never require them.
		if source := os.Getenv("GOGALLERY_UI_PHOTOS"); source != "" {
			name := "flowers.png"
			if i%2 == 0 {
				name = "branches.jpg"
			}
			data, err := os.ReadFile(filepath.Join(source, name))
			if err == nil {
				if err = os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
		album := "coast"
		if i >= 12 {
			album = "garden"
		}
		if i == 16 {
			album = "macro"
		}
		if i == 17 {
			album = "garden2"
		}
		pic := datastore.Picture{Id: fmt.Sprintf("photo%02d", i), Name: fmt.Sprintf("Photo %02d", i), Path: path, Album: album, AlbumName: "Garden", Tags: "summer, landscape", Visibility: "PUBLIC", Camera: "Example camera", DateTaken: time.Date(2026, 7, i+1, 12, 0, 0, 0, time.UTC)}
		if err := db.Pictures.Save(pic); err != nil {
			t.Fatal(err)
		}
	}
	server := preview.NewServer(db)
	defer server.Stop()
	gallery := pages.NewGalleryPage(db)
	albumsPage := pages.NewCollectionPage(db)
	settings := pages.NewSettingsPage(db, server)
	tasks := pages.NewTasksPage(db, server)
	views := map[string]pages.Page{"Photos": gallery, "Albums": albumsPage, "Settings": settings, "Tasks": tasks}
	content := container.NewStack()
	var nav *components.Header
	setPage := func(name string) {
		tasks.SetActive(name == "Tasks")
		nav.SetActive(name)
		content.Objects = []fyne.CanvasObject{views[name].Layout()}
		content.Refresh()
	}
	nav = components.NewHeader("GoGallery", server, setPage)
	window := test.NewWindow(nav.Shell(content))
	defer window.Close()
	window.SetPadded(false)
	window.Resize(fyne.NewSize(1280, 800))
	window.Show()
	setPage("Photos")
	driver.until(t, func() bool { return hasLabel(content, "18 photos") })
	// Allow queued thumbnails to finish before captures and user interactions.
	driver.until(t, func() bool {
		for _, o := range descendants(content) {
			if l, ok := o.(*widget.Label); ok && l.Text == "Loading thumbnail…" {
				return false
			}
		}
		return true
	})
	for _, object := range descendants(content) {
		if label, ok := object.(*widget.Label); ok && label.Text == "1 / 2" {
			pos := app.Driver().AbsolutePositionForObject(label)
			if pos.Y+label.Size().Height > 800 {
				t.Fatal("pagination is clipped below window")
			}
		}
	}
	for _, object := range descendants(content) {
		if check, ok := object.(*widget.Check); ok && check.Text == "Titles" {
			t.Fatal("title toggle should be removed")
		}
	}
	tile := firstPhoto(t, content)
	tile.MouseIn(&desktop.MouseEvent{})
	captureWorkspace(t, window, "photo-hover-title")
	tile.MouseOut()
	captureWorkspace(t, window, "photos-desktop")
	search := entryPlaceholder(t, content, "Search titles, filenames, or tags")
	search.SetText("Photo 17")
	if !hasLabel(content, "1 of 18 photos") {
		t.Fatal("photo search did not narrow the library")
	}
	test.Tap(firstPhoto(t, content))
	if !hasLabel(content, "Photo details") {
		t.Fatal("photo selection did not open inspector")
	}
	driver.pump()
	photoTitle := entryPlaceholder(t, content, "Photo title")
	photoTitle.SetText("Photo 17 edited")
	test.Tap(buttonNamed(t, content, "Save changes"))
	driver.until(t, func() bool { return hasLabel(content, "Changes saved") })
	saved, err := db.Pictures.FindByID("photo17")
	if err != nil || saved.Name != "Photo 17 edited" {
		t.Fatalf("photo edit did not persist: %v, %v", saved.Name, err)
	}
	captureWorkspace(t, window, "photo-inspector-desktop")
	window.Resize(fyne.NewSize(600, 640))
	captureWorkspace(t, window, "photo-inspector-narrow")
	close := buttonNamed(t, content, "Close")
	test.Tap(close)
	if !hasLabel(content, "1 of 18 photos") {
		t.Fatal("closing inspector lost filtered grid")
	}
	window.Resize(fyne.NewSize(1280, 800))
	search.SetText("")
	windowsBefore := len(app.Driver().AllWindows())
	test.DoubleTap(firstPhoto(t, content))
	if len(app.Driver().AllWindows()) != windowsBefore+1 {
		t.Fatal("double click did not open a viewer")
	}
	viewer := app.Driver().AllWindows()[windowsBefore]
	if viewer.Title() != "Photo 17 edited — GoGallery" {
		t.Fatalf("viewer title: %s", viewer.Title())
	}
	viewer.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyRight})
	if viewer.Title() != "Photo 16 — GoGallery" {
		t.Fatal("right arrow did not navigate to next photo")
	}
	viewer.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyLeft})
	if viewer.Title() != "Photo 17 edited — GoGallery" {
		t.Fatal("left arrow did not navigate to previous photo")
	}
	driver.until(t, func() bool { return !hasLabel(viewer.Content(), "Loading preview…") })
	captureWorkspace(t, viewer, "photo-viewer")
	viewer.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if len(app.Driver().AllWindows()) != windowsBefore {
		t.Fatal("Escape did not close viewer")
	}
	setPage("Albums")
	driver.until(t, func() bool { return hasLabel(content, "4 albums · 7 including subalbums") })
	driver.pump()
	captureWorkspace(t, window, "albums-desktop")
	albumSearch := entryPlaceholder(t, content, "Find an album")
	if hasLabel(content, "Garden") || hasLabel(content, "Close-ups") {
		t.Fatal("subalbums leaked into the root album list")
	}
	test.Tap(albumCoverNamed(t, content, "2026"))
	driver.until(t, func() bool { return hasLabel(content, "1 subalbum") })
	if hasLabel(content, "This album is empty") || !hasLabel(content, "Garden") || hasLabel(content, "Close-ups") {
		t.Fatal("parent should show only immediate subalbums")
	}
	captureWorkspace(t, window, "subalbums-parent")
	window.Resize(fyne.NewSize(600, 640))
	captureWorkspace(t, window, "subalbums-parent-narrow")
	window.Resize(fyne.NewSize(1280, 800))
	test.Tap(albumCoverNamed(t, content, "Garden"))
	driver.until(t, func() bool { return hasLabel(content, "1 subalbum · 1 photo in this album") })
	captureWorkspace(t, window, "subalbums-and-photos")
	test.Tap(albumCoverNamed(t, content, "Close-ups"))
	driver.until(t, func() bool { return hasLabel(content, "1 photo") })
	if !hasLabel(content, "Close-ups") {
		t.Fatal("grandchild album did not open")
	}
	test.Tap(buttonNamed(t, content, "Back"))
	driver.until(t, func() bool { return hasLabel(content, "1 subalbum · 1 photo in this album") })
	test.Tap(buttonNamed(t, content, "2026"))
	driver.until(t, func() bool { return hasLabel(content, "1 subalbum") })
	// Rename the parent without changing its folder path; its child must stay attached.
	test.Tap(buttonNamed(t, content, "Album details"))
	parentTitle := entryPlaceholder(t, content, "Album title")
	parentTitle.SetText("This year")
	test.Tap(buttonNamed(t, content, "Save album"))
	driver.until(t, func() bool { return hasLabel(content, "This year") })
	test.Tap(buttonNamed(t, content, "Close"))
	test.Tap(albumCoverNamed(t, content, "Garden"))
	driver.until(t, func() bool { return hasLabel(content, "1 subalbum · 1 photo in this album") })
	if buttonNamed(t, content, "This year") == nil {
		t.Fatal("renamed parent missing from breadcrumbs")
	}
	// The Garden in the other branch contains four different photos.
	test.Tap(lastPhoto(t, content))
	test.Tap(buttonNamed(t, content, "Use as album cover"))
	album, err := db.Albums.FindById("garden2")
	if err != nil || album.ProfileId != "photo17" {
		t.Fatalf("cover = %s, err = %v", album.ProfileId, err)
	}
	test.Tap(buttonNamed(t, content, "Close"))
	test.Tap(buttonNamed(t, content, "Album details"))
	title := entryPlaceholder(t, content, "Album title")
	title.SetText("Summer garden")
	test.Tap(buttonNamed(t, content, "Save album"))
	driver.until(t, func() bool { return hasLabel(content, "Summer garden") })
	album, _ = db.Albums.FindById("garden2")
	if album.Name != "Summer garden" || album.ProfileId != "photo17" {
		t.Fatal("rename lost album identity or cover")
	}
	test.Tap(buttonNamed(t, content, "Close"))
	test.Tap(buttonNamed(t, content, "Back"))
	driver.until(t, func() bool { return hasLabel(content, "1 subalbum") })
	if !hasLabel(content, "Summer garden") {
		t.Fatal("renamed child vanished from its parent")
	}
	test.Tap(buttonNamed(t, content, "All albums"))
	albumSearch.SetText("Summer garden")
	if !hasLabel(content, "1 matching album") {
		t.Fatal("search must find nested albums")
	}
	test.Tap(albumCoverNamed(t, content, "Summer garden"))
	driver.until(t, func() bool { return hasLabel(content, "1 subalbum · 1 photo in this album") })
	test.Tap(buttonNamed(t, content, "All albums"))
	albumSearch.SetText("empty")
	test.Tap(firstPhoto(t, content))
	driver.until(t, func() bool { return hasLabel(content, "This album is empty") })
	setPage("Settings")
	captureWorkspace(t, window, "settings-desktop")
	window.Resize(fyne.NewSize(600, 640))
	captureWorkspace(t, window, "settings-narrow")
	button := buttonNamed(t, content, "Save changes")
	position := app.Driver().AbsolutePositionForObject(button)
	if position.Y+button.Size().Height > 640 {
		t.Fatalf("save button outside narrow window: %v", position)
	}
	for _, o := range descendants(content) {
		if tabs, ok := o.(*container.AppTabs); ok {
			tabs.SelectIndex(1)
			break
		}
	}
	captureWorkspace(t, window, "settings-author-narrow")
	for _, size := range []fyne.Size{fyne.NewSize(1280, 800), fyne.NewSize(1920, 1080), fyne.NewSize(2560, 1440)} {
		window.Resize(size)
		captureWorkspace(t, window, fmt.Sprintf("settings-author-%d", int(size.Width)))
	}
	window.Resize(fyne.NewSize(600, 640))
	setPage("Tasks")
	captureWorkspace(t, window, "tasks-narrow")
	test.Tap(buttonNamed(t, content, "Maintenance options"))
	if !hasLabel(content, "Generated website cleanup") || buttonNamed(t, content, "Delete output").Disabled() {
		t.Fatal("Maintenance options did not reveal the available cleanup action")
	}
	captureWorkspace(t, window, "tasks-maintenance-narrow")
	test.Tap(buttonNamed(t, content, "Hide maintenance options"))
	if hasLabel(content, "Generated website cleanup") {
		t.Fatal("maintenance panel did not close")
	}
	window.Resize(fyne.NewSize(1280, 800))
	test.Tap(buttonNamed(t, content, "Build website"))
	driver.until(t, func() bool { return hasLabel(content, "Failed") })
	if !hasLabel(content, "gallery destination path is not configured") {
		t.Fatal("build error is not visible in Activity")
	}
	if hasLabel(content, "Build website…") || hasLabel(content, "Build website failed: gallery destination path is not configured") {
		t.Fatal("task status is duplicated outside Activity")
	}
	completed := db.NewTask("Refresh thumbnails", 4)
	completed.Start()
	completed.Complete()
	running := db.NewTask("Prepare website assets", 8)
	running.Start()
	running.Update()
	running.Update()
	imageFailure := "1 item failed:\n/photos/Trips/Florence/panorama.jpg: write image size xlarge: no space left on device"
	failed := db.NewTask("Building images", 1)
	failed.Start()
	failed.Fail(imageFailure)
	driver.until(t, func() bool { return hasLabel(content, "Complete") && hasLabel(content, "25%") })
	driver.until(t, func() bool { return hasLabel(content, imageFailure) })
	captureWorkspace(t, window, "tasks-desktop")
	window.Resize(fyne.NewSize(600, 640))
	captureWorkspace(t, window, "tasks-status-narrow")
	tasks.SetActive(false)
	app.Settings().SetTheme(NewComfortableTheme("light"))
	// Create new panels as the production app does on theme changes at restart.
	nav.SetActive("Settings")
	light := pages.NewSettingsPage(db, server)
	content.Objects = []fyne.CanvasObject{light.Layout()}
	content.Refresh()
	lightNav := components.NewHeader("GoGallery", server, nil)
	window.SetContent(lightNav.Shell(content))
	lightNav.SetActive("Settings")
	window.Resize(fyne.NewSize(1920, 1080))
	for _, object := range descendants(content) {
		if tabs, ok := object.(*container.AppTabs); ok {
			for i, name := range []string{"library", "author", "application", "publishing"} {
				tabs.SelectIndex(i)
				captureWorkspace(t, window, "settings-light-"+name+"-wide")
			}
			tabs.SelectIndex(2)
			window.Resize(fyne.NewSize(2560, 1440))
			captureWorkspace(t, window, "settings-light-application-fullscreen")
			window.Resize(fyne.NewSize(600, 640))
			captureWorkspace(t, window, "settings-light-application-narrow")
			tabs.SelectIndex(0)
			break
		}
	}
	window.Resize(fyne.NewSize(520, 400))
	captureWorkspace(t, window, "settings-light-small")
	for _, object := range descendants(content) {
		if entry, ok := object.(*widget.Entry); ok {
			position := app.Driver().AbsolutePositionForObject(entry)
			save := buttonNamed(t, content, "Save changes")
			footerTop := app.Driver().AbsolutePositionForObject(save).Y - 18
			if position.Y+entry.Size().Height > footerTop {
				t.Fatalf("first settings field is obscured in a small window: %v, %v", position, entry.Size())
			}
			break
		}
	}
	lightTasks := pages.NewTasksPage(db, server)
	content.Objects = []fyne.CanvasObject{lightTasks.Layout()}
	content.Refresh()
	lightNav.SetActive("Tasks")
	for _, size := range []fyne.Size{fyne.NewSize(1920, 1080), fyne.NewSize(600, 640), fyne.NewSize(520, 400)} {
		window.Resize(size)
		captureWorkspace(t, window, fmt.Sprintf("tasks-light-%d", int(size.Width)))
	}
	driver.pump()
}
