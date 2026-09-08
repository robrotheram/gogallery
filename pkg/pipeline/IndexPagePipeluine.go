package pipeline

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/embeds"
	templateengine "gogallery/pkg/templateEngine"
)

func (r *RenderPipeline) BuildIndex(w io.Writer) error {
	imagesPerPage := r.config.ImagesPerPage
	if imagesPerPage <= 0 {
		imagesPerPage = 24
	}
	indexPage := templateengine.NewPage(nil)
	images, err := r.Pictures.GetFilteredPictures(false)
	if err != nil {
		return fmt.Errorf("load public pictures: %w", err)
	}

	if len(images) > 0 {
		featuredImage := images[0]
		images = images[1:]
		indexPage.SEO.SetImage(featuredImage)
		indexPage.Picture = templateengine.NewPagePicture(featuredImage)
	}

	if pages := paginateImages(images, imagesPerPage); len(pages) > 0 {
		indexPage.Images = pages[0]
	}

	albums, err := r.Albums.GetLatestAlbums()
	if err != nil {
		return fmt.Errorf("load latest albums: %w", err)
	}
	indexPage.Albums = make([]datastore.AlbumNode, 0, min(3, max(0, len(albums)-1)))
	// Skip the first album because it is featured separately.
	for i := 1; i < len(albums); i++ {
		if i >= 4 {
			break
		}
		indexPage.Albums = append(indexPage.Albums, albums[i].ToAlbumNode())
	}

	if len(albums) > 0 {
		indexPage.FeaturedAlbum = albums[0].ToAlbumNode()
	}

	return templateengine.Templates.RenderPage(w, templateengine.HomeTemplate, indexPage)
}

func (r *RenderPipeline) BuildAlbums(w io.Writer) error {
	page := templateengine.NewPage(nil)
	page.SEO.SetPath("/albums/")
	page.SEO.Title = "Photo albums"
	tree, err := r.Albums.GetAlbumStructure(page.Settings)
	if err != nil {
		return fmt.Errorf("load album structure: %w", err)
	}
	page.Albums = datastore.GetAlbumsFromTree(tree)
	return templateengine.Templates.RenderPage(w, templateengine.AlbumTemplate, page)
}

func (r *RenderPipeline) renderIndex() error {
	if err := writeGeneratedFile(filepath.Join(r.root, "index.html"), func(w io.Writer) error {
		return r.BuildIndex(w)
	}); err != nil {
		return err
	}
	if err := writeGeneratedFile(filepath.Join(r.root, "manifest.json"), func(w io.Writer) error {
		return templateengine.ManifestWriter(w, r.config)
	}); err != nil {
		return err
	}
	return writeGeneratedFile(filepath.Join(r.root, "service-worker.js"), templateengine.ServiceWorkerWriter)
}

func paginateImages(slice []datastore.Picture, chunkSize int) [][]datastore.Picture {
	if len(slice) == 0 || chunkSize <= 0 {
		return nil
	}
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

func (r *RenderPipeline) renderAlbums() error {
	// #nosec G301 -- generated website directories must be readable by a web server.
	if err := os.MkdirAll(r.albumsDir, 0o755); err != nil {
		return err
	}
	return writeGeneratedFile(filepath.Join(r.albumsDir, "index.html"), func(w io.Writer) error {
		return r.BuildAlbums(w)
	})
}

func (r *RenderPipeline) assets() error {
	theme := config.NormalizeTheme(r.config.Theme)
	if embeds.DoesThemeExist(theme) {
		return embeds.CopyThemeAssets(theme, filepath.Join(r.root, "assets"))
	}
	return templateengine.Dir(filepath.Join(theme, "assets"), filepath.Join(r.root, "assets"))
}
