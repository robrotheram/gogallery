package templateengine

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/datastore"
)

func RenderPhotoHandle(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
	vars := mux.Vars(r)
	id := vars["id"]
	images := datastore.GetFilteredPictures(false)
	RenderPhoto(w, id, images, page)
}

func RenderPhoto(w http.ResponseWriter, picID string, images []datastore.Picture, page Page) {
	pic, err := datastore.GetPictureByID(picID)
	if err != nil {
		w.Write([]byte(Templates.RenderPage("404", Page{})))
		return
	}
	images = datastore.SortByTime(images)
	for i, p := range images {
		if p.Id == pic.Id {
			if i-1 >= 0 {
				page.PreImagePath = images[i-1].Id
			}
			if i+1 < len(images) {
				page.NextImagePath = images[i+1].Id
			}
		}
	}
	page.Picture = pic
	page.SEO.SetNameFromPhoto(pic)
	w.Write([]byte(Templates.RenderPage(PhotoTemplate, page)))
}
