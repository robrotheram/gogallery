package pipeline

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
	"github.com/robrotheram/gogallery/datastore"
	templateengine "github.com/robrotheram/gogallery/templateEngine"
)

func renderAlbumTemplate(alb datastore.Album) error {
	alb_path := filepath.Join(albumDir, slug.Make(alb.Id))
	os.MkdirAll(alb_path, os.ModePerm)

	page := templateengine.NewPage(nil)
	albums := datastore.GetAlbumStructure(page.Settings)
	album := datastore.GetAlbumFromStructure(albums, alb.Id)

	f, _ := os.Create(filepath.Join(alb_path, "index.html"))
	w := bufio.NewWriter(f)
	page.Album = album
	page.Images = datastore.GetPicturesByAlbumID(alb.Id)
	page.Picture, _ = datastore.GetPictureByID(alb.ProfileID)
	page.SEO.Description = fmt.Sprintf("Album: %s", alb.Name)
	w.Write([]byte(templateengine.Templates.RenderPage(templateengine.CollectionTemplate, page)))
	f.Close()
	return nil
}
