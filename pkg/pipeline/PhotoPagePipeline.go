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

func (r *RenderPipeline) BuildPhoto(pic datastore.Picture, w io.Writer) error {

	album, err := r.Pictures.FindPublicByAlbum(pic.Album)
	if err != nil {
		return fmt.Errorf("load public album pictures: %w", err)
	}
	return templateengine.RenderPhoto(w, pic, album, templateengine.NewPage(nil))
}

func (r *RenderPipeline) renderPhotoTemplate() func(alb datastore.Picture) error {
	return func(pic datastore.Picture) error {
		picPath := filepath.Join(r.photoDir, slug.Make(pic.Id))
		// #nosec G301 -- generated website directories must be readable by a web server.
		if err := os.MkdirAll(picPath, 0o755); err != nil {
			return err
		}
		return writeGeneratedFile(filepath.Join(picPath, "index.html"), func(w io.Writer) error {
			return r.BuildPhoto(pic, w)
		})
	}
}
