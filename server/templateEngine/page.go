package templateengine

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
)

type Page struct {
	Settings config.GalleryConfiguration
	Author   config.AboutConfiguration
	Images   []datastore.Picture
	Albums   datastore.AlbumStrcure
	Album    datastore.Album
	Picture  datastore.Picture
	Body     string
}

func NewPage(config *config.Configuration) Page {
	return Page{
		Settings: config.Gallery,
		Author:   config.About,
	}
}

var te = NewTemplateEgine()

func (p *Page) Handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page := vars["page"]
	if te.Pages[page] == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Page " + page + " Not found"))
		return
	}

	w.Write([]byte(te.RenderPage(page, *p)))
}
