package pipeline

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/embeds"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
)

func (r *RenderPipeline) BuildIndex(w io.Writer) {
	imagesPerPage := 24
	latestAlbumID := r.GetLatestAlbum()

	if r.config.ImagesPerPage == 0 {
		alb, _ := r.Pictures.FindByField("Album", latestAlbumID)
		imagesPerPage = len(alb) - 2 // reserve 2 images for the featured image and the picture of the day
	}

	indexPage := templateengine.NewPage(nil)
	images := r.Pictures.GetFilteredPictures(false)

	featuredImage := images[0]
	images = images[1:]

	pages := paginateImages(images, imagesPerPage)
	indexPage.Images = pages[0]

	albums := r.Albums.GetAlbumStructure(indexPage.Settings)

	firstThreeAlbums := make(datastore.AlbumStrcure, 3)
	count := 0
	for _, album := range albums {
		if count >= 3 {
			break
		}
		if album.Id != latestAlbumID {
			firstThreeAlbums[album.Id] = album.ToAlbumNode()
			count++
		}
	}
	indexPage.Albums = firstThreeAlbums

	if len(images) > 0 {
		indexPage.SEO.SetImage(featuredImage)
		indexPage.Picture = templateengine.PagePicture{
			Picture: featuredImage,
		}
	}
	featuedAlbum, _ := r.Albums.FindById(latestAlbumID)
	indexPage.FeaturedAlbum = featuedAlbum.ToAlbumNode()

	templateengine.Templates.RenderPage(w, templateengine.HomeTemplate, indexPage)
}

func (r *RenderPipeline) BuildAlbums(w io.Writer) {
	page := templateengine.NewPage(nil)
	page.Albums = r.Albums.GetAlbumStructure(page.Settings)
	templateengine.Templates.RenderPage(w, templateengine.AlbumTemplate, page)
}

func (r *RenderPipeline) renderIndex() {

	f, _ := os.Create(filepath.Join(root, "index.html"))
	w := bufio.NewWriter(f)
	r.BuildIndex(w)
	// renderPages(pages, latestAlbumID, albums)
	w.Flush()
	f.Close()

	f, _ = os.Create(filepath.Join(root, "manifest.json"))
	w = bufio.NewWriter(f)
	templateengine.ManifestWriter(w, r.config)
	w.Flush()
	f.Close()

	f, _ = os.Create(filepath.Join(root, "service-worker.js"))
	w = bufio.NewWriter(f)
	templateengine.ServiceWorkerWriter(w)
	w.Flush()
	f.Close()
}

func paginateImages(slice []datastore.Picture, chunkSize int) [][]datastore.Picture {
	var chunks [][]datastore.Picture
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func (r *RenderPipeline) renderPages(pages [][]datastore.Picture, albumID string, albums datastore.AlbumStrcure) {
	pagesPath := filepath.Join(root, "page")
	os.MkdirAll(pagesPath, os.ModePerm)
	for page, pageImages := range pages {
		pagePath := filepath.Join(pagesPath, fmt.Sprint(page))
		os.MkdirAll(pagePath, os.ModePerm)
		page := templateengine.NewPage(nil)
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

func (r *RenderPipeline) renderAlbums() {
	os.MkdirAll(albumsDir, os.ModePerm)
	f, _ := os.Create(filepath.Join(albumsDir, "index.html"))
	w := bufio.NewWriter(f)
	r.BuildAlbums(w)
	w.Flush()
	f.Close()
}

func build() {
	err := templateengine.Templates.Load(config.Config.Gallery.Theme)
	if err != nil {
		fmt.Println(err)
	}
	if config.Config.Gallery.Theme == "default" {
		embeds.CopyThemeAssets(filepath.Join(root, "assets"))
	} else {
		templateengine.Dir(filepath.Join(config.Config.Gallery.Theme, "assets"), filepath.Join(root, "assets"))
	}
}
