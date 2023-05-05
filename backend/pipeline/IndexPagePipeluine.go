package pipeline

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/datastore/models"
	"github.com/robrotheram/gogallery/backend/embeds"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
)

func renderIndex(db *datastore.DataStore, config *config.GalleryConfiguration) {
	latestAlbumID := db.Pictures.GetLatestAlbum()
	indexPage := templateengine.NewPage(nil, latestAlbumID)
	images := db.Pictures.GetFilteredPictures(false)
	pages := paginateImages(images, 16)
	albums := datastore.Sort(db.Albums.GetAlbumStructure(indexPage.Settings))

	indexPage.Images = pages[0]
	indexPage.Albums = albums
	if len(images) > 0 {
		indexPage.SEO.SetImage(images[0])
	}

	f, _ := os.Create(filepath.Join(root, "index.html"))
	w := bufio.NewWriter(f)
	templateengine.Templates.RenderPage(w, templateengine.HomeTemplate, indexPage)
	renderPages(pages, latestAlbumID, albums)
	w.Flush()
	f.Close()

	f, _ = os.Create(filepath.Join(root, "manifest.json"))
	w = bufio.NewWriter(f)
	templateengine.ManifestWriter(w, config)
	w.Flush()
	f.Close()

	f, _ = os.Create(filepath.Join(root, "service-worker.js"))
	w = bufio.NewWriter(f)
	templateengine.ServiceWorkerWriter(w)
	w.Flush()
	f.Close()
}

func paginateImages(slice []models.Picture, chunkSize int) [][]models.Picture {
	var chunks [][]models.Picture
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func renderPages(pages [][]models.Picture, albumID string, albums models.AlbumStrcure) {
	pagesPath := filepath.Join(root, "page")
	os.MkdirAll(pagesPath, os.ModePerm)
	for page, pageImages := range pages {
		pagePath := filepath.Join(pagesPath, fmt.Sprint(page))
		os.MkdirAll(pagePath, os.ModePerm)
		page := templateengine.NewPage(nil, albumID)
		page.Images = pageImages
		page.Albums = albums
		if len(pageImages) > 0 {
			page.SEO.SetImage(pageImages[0])
		}
		f, _ := os.Create(filepath.Join(pagePath, "index.html"))
		w := bufio.NewWriter(f)
		templateengine.Templates.RenderPage(w, templateengine.PaginationTemplate, page)
		w.Flush()
		f.Close()
	}
}

func renderAlbums(db *datastore.DataStore) {
	os.MkdirAll(albumsDir, os.ModePerm)
	f, _ := os.Create(filepath.Join(albumsDir, "index.html"))
	w := bufio.NewWriter(f)
	page := templateengine.NewPage(nil, db.Pictures.GetLatestAlbum())
	page.Albums = datastore.Sort(db.Albums.GetAlbumStructure(page.Settings))
	templateengine.Templates.RenderPage(w, templateengine.AlbumTemplate, page)
	w.Flush()
	f.Close()
}

func build() {
	err := templateengine.Templates.Load(config.Config.Gallery.Theme)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if config.Config.Gallery.Theme == "default" {
		embeds.CopyThemeAssets(filepath.Join(root, "assets"))
	} else {
		templateengine.Dir(filepath.Join(config.Config.Gallery.Theme, "assets"), filepath.Join(root, "assets"))
	}

}
