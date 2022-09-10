package pipeline

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/datastore/models"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
)

func renderAlbumTemplate(db *datastore.DataStore) func(alb models.Album) error {
	return func(alb models.Album) error {
		alb_path := filepath.Join(albumDir, slug.Make(alb.Id))
		os.MkdirAll(alb_path, os.ModePerm)

		page := templateengine.NewPage(nil)
		albums := db.Albums.GetAlbumStructure(page.Settings)
		album := datastore.GetAlbumFromStructure(albums, alb.Id)

		f, _ := os.Create(filepath.Join(alb_path, "index.html"))
		w := bufio.NewWriter(f)
		page.Album = album
		page.Images = db.Pictures.GetByAlbumID(alb.Id)
		page.Picture = db.Pictures.FindByID(alb.ProfileID)
		page.SEO.Description = fmt.Sprintf("Album: %s", alb.Name)
		templateengine.Templates.RenderPage(w, templateengine.CollectionTemplate, page)
		f.Close()
		return nil
	}
}
