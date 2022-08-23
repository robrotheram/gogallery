package pipeline

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
	"github.com/robrotheram/gogallery/datastore"
	templateengine "github.com/robrotheram/gogallery/templateEngine"
)

var photoDir = filepath.Join(root, "photo")

func renderPhotoTemplate(pic datastore.Picture) error {
	pic_path := filepath.Join(photoDir, slug.Make(pic.Id))
	os.MkdirAll(pic_path, os.ModePerm)
	f, err := os.Create(filepath.Join(pic_path, "index.html"))
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	templateengine.RenderPhoto(w, pic.Id, datastore.GetPicturesByAlbumID(pic.Album), templateengine.NewPage(nil))
	f.Close()
	return nil
}
