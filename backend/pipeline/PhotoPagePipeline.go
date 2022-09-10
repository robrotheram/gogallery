package pipeline

import (
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/datastore/models"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
)

func renderPhotoTemplate(db *datastore.DataStore) func(alb models.Picture) error {
	return func(pic models.Picture) error {
		pic_path := filepath.Join(photoDir, slug.Make(pic.Id))
		os.MkdirAll(pic_path, os.ModePerm)
		f, err := os.Create(filepath.Join(pic_path, "index.html"))
		if err != nil {
			return err
		}

		templateengine.RenderPhoto(f, pic, db.Pictures.GetByAlbumID(pic.Album), templateengine.NewPage(nil))

		f.Close()
		return nil
	}
}
