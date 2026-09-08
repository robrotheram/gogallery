package pipeline

import (
	"fmt"
	"gogallery/pkg/datastore"
	templateengine "gogallery/pkg/templateEngine"
	"io"
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
)

func (r *RenderPipeline) BuildAlbum(albId string, w io.Writer) error {
	page := templateengine.NewPage(nil)

	albums, err := r.Albums.GetAlbumStructure(page.Settings)
	if err != nil {
		return fmt.Errorf("load album structure: %w", err)
	}
	album := datastore.GetAlbumFromStructure(albums, albId)
	if album.Id == "" {
		return fmt.Errorf("album %q not found", albId)
	}

	images, err := r.Pictures.FindPublicByAlbum(album.Id)
	if err != nil {
		return fmt.Errorf("load public album pictures: %w", err)
	}
	page.Images = images

	var profile datastore.Picture
	if album.ProfileId != "" {
		profile, err = r.Pictures.FindByID(album.ProfileId)
		if err != nil {
			return fmt.Errorf("load album profile picture: %w", err)
		}
		if !datastore.IsPicturePublishable(profile) {
			return fmt.Errorf("album profile picture is not public")
		}
	} else if len(images) > 0 {
		profile = images[0]
		album.ProfileId = profile.Id
	}

	page.Album = album
	page.Picture = templateengine.NewPagePicture(profile)
	page.SEO.SetPath(fmt.Sprintf("/album/%s/", slug.Make(album.Id)))
	page.SEO.Description = fmt.Sprintf("Explore %d photographs from %s.", len(page.Images), album.Name)
	page.SEO.Title = fmt.Sprintf("%s photo album", album.Name)
	page.SEO.SetImage(profile)

	return templateengine.Templates.RenderPage(w, templateengine.CollectionTemplate, page)
}

func (r *RenderPipeline) renderAlbumTemplate() func(alb datastore.Album) error {
	return func(alb datastore.Album) error {
		albPath := filepath.Join(r.albumDir, slug.Make(alb.Id))
		// #nosec G301 -- generated website directories must be readable by a web server.
		if err := os.MkdirAll(albPath, 0o755); err != nil {
			return err
		}
		return writeGeneratedFile(filepath.Join(albPath, "index.html"), func(w io.Writer) error {
			return r.BuildAlbum(alb.Id, w)
		})
	}
}
