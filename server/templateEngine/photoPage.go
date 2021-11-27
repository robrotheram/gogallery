package templateengine

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/datastore"
)

func RenderPhoto(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
	vars := mux.Vars(r)
	id := vars["id"]
	pic, err := datastore.GetPictureByID(id)
	if err != nil {
		w.Write([]byte(Templates.RenderPage("404", Page{})))
		return
	}
	alb, err := datastore.GetAlbumByID(pic.Album)
	images := datastore.GetPicturesByAlbumID(alb.Id)
	for i, p := range images {
		if p.Id == pic.Id {
			if i-1 > 0 {
				page.PreImageID = images[i-1].Id
			}
			if i+1 < len(images) {
				page.NextImageID = images[i+1].Id
			}
		}
	}
	page.Picture = pic
	page.SEO.SetImage(pic)
	w.Write([]byte(Templates.RenderPage(PhotoTemplate, page)))
}
