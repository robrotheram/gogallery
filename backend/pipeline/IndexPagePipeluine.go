package pipeline

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
)

func renderIndex(db *datastore.DataStore) {
	f, _ := os.Create(filepath.Join(root, "index.html"))
	w := bufio.NewWriter(f)
	indexPage := templateengine.NewPage(nil)
	images := db.Pictures.GetFilteredPictures(false)
	indexPage.Images = images
	indexPage.Albums = datastore.Sort(db.Albums.GetAlbumStructure(indexPage.Settings))
	if len(images) > 0 {
		indexPage.SEO.SetImage(images[0])
	}
	templateengine.Templates.RenderPage(w, templateengine.HomeTemplate, indexPage)

	f, _ = os.Create(filepath.Join(root, "manifest.json"))
	w = bufio.NewWriter(f)
	templateengine.ManifestWriter(w, "test")
	w.Flush()
	f.Close()

	f, _ = os.Create(filepath.Join(root, "service-worker.js"))
	w = bufio.NewWriter(f)
	templateengine.ServiceWorkerWriter(w)
	w.Flush()
	f.Close()
}

func renderAlbums(db *datastore.DataStore) {
	os.MkdirAll(albumsDir, os.ModePerm)
	f, _ := os.Create(filepath.Join(albumsDir, "index.html"))
	w := bufio.NewWriter(f)
	page := templateengine.NewPage(nil)
	page.Albums = datastore.Sort(db.Albums.GetAlbumStructure(page.Settings))
	templateengine.Templates.RenderPage(w, templateengine.AlbumTemplate, page)
}

func build() {
	err := templateengine.Templates.Load(config.Config.Gallery.Theme)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	templateengine.Dir(filepath.Join(config.Config.Gallery.Theme, "assets"), filepath.Join(root, "assets"))
}
