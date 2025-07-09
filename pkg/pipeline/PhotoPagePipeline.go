package pipeline

import (
	"gogallery/pkg/datastore"
	templateengine "gogallery/pkg/templateEngine"
	"io"
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
)

func (r *RenderPipeline) BuildPhoto(pic datastore.Picture, w io.Writer) {

	album, _ := r.Pictures.FindByField("Album", pic.Album)
	templateengine.RenderPhoto(w, pic, album, templateengine.NewPage(nil))
}

func (r *RenderPipeline) renderPhotoTemplate() func(alb datastore.Picture) error {
	return func(pic datastore.Picture) error {
		picPath := filepath.Join(photoDir, slug.Make(pic.Id))
		os.MkdirAll(picPath, os.ModePerm)
		f, err := os.Create(filepath.Join(picPath, "index.html"))
		if err != nil {
			return err
		}
		r.BuildPhoto(pic, f)
		f.Close()
		return nil
	}
}
