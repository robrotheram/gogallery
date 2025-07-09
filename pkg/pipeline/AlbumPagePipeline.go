package pipeline

import (
	"bufio"
	"fmt"
	"gogallery/pkg/datastore"
	templateengine "gogallery/pkg/templateEngine"
	"io"
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
)

func (r *RenderPipeline) BuildAlbum(albId string, w io.Writer) {
	page := templateengine.NewPage(nil)

	albums := r.Albums.GetAlbumStructure(page.Settings)
	album := datastore.GetAlbumFromStructure(albums, albId)

	page.Album = album
	page.Images, _ = r.Pictures.FindByField("Album", album.Id)

	profile, _ := r.Pictures.FindById(album.ProfileId)

	page.Picture = templateengine.NewPagePicture(profile)
	page.SEO.Description = fmt.Sprintf("Album: %s", album.Name)
	page.SEO.Title = fmt.Sprintf("Album: %s", album.Name)
	page.SEO.SetImage(profile)

	templateengine.Templates.RenderPage(w, templateengine.CollectionTemplate, page)

}

func (r *RenderPipeline) renderAlbumTemplate() func(alb datastore.Album) error {
	return func(alb datastore.Album) error {
		albPath := filepath.Join(albumDir, slug.Make(alb.Id))
		os.MkdirAll(albPath, os.ModePerm)
		f, _ := os.Create(filepath.Join(albPath, "index.html"))
		w := bufio.NewWriter(f)
		r.BuildAlbum(alb.Id, w)
		f.Close()
		return nil
	}
}
