package pipeline

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	templateengine "github.com/robrotheram/gogallery/templateEngine"
)

var albumsDir = filepath.Join(root, "albums")

func renderIndex() {
	f, _ := os.Create(filepath.Join(root, "index.html"))
	w := bufio.NewWriter(f)
	indexPage := templateengine.NewPage(nil)
	images := datastore.GetFilteredPictures(false)
	indexPage.Images = images
	indexPage.Albums = datastore.Sort(datastore.GetAlbumStructure(indexPage.Settings))
	if len(images) > 0 {
		indexPage.SEO.SetImage(images[0])
	}
	w.Write([]byte(templateengine.Templates.RenderPage(templateengine.HomeTemplate, indexPage)))

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

func renderAlbums() {
	os.MkdirAll(albumsDir, os.ModePerm)
	f, _ := os.Create(filepath.Join(albumsDir, "index.html"))
	w := bufio.NewWriter(f)
	page := templateengine.NewPage(nil)
	page.Albums = datastore.Sort(datastore.GetAlbumStructure(page.Settings))
	w.Write([]byte(templateengine.Templates.RenderPage(templateengine.AlbumTemplate, page)))
}

func build() {
	templateengine.Templates.Load(config.Config.Gallery.Theme)
	templateengine.Dir(filepath.Join(config.Config.Gallery.Theme, "assets"), filepath.Join(root, "assets"))
}
