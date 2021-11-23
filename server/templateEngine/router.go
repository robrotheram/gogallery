package templateengine

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/config"
)

const HomeTemplate = "main"
const AlbumTemplate = "albums"
const CollectionTemplate = "collection"
const PhotoTemplate = "photo"

func (p *Page) renderIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(te.RenderPage(HomeTemplate, *p)))
}

func (p *Page) renderAlbum(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(te.RenderPage(AlbumTemplate, *p)))
}

func (p *Page) renderCollection(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(te.RenderPage(CollectionTemplate, *p)))
}

func (p *Page) renderPhoto(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(te.RenderPage(PhotoTemplate, *p)))
}

func InitApiRoutes(r *mux.Router, config *config.Configuration) *mux.Router {
	err := te.Load("../templates/beta")

	if err != nil {
		fmt.Printf("there Was an error loading the templates, Error %v", err)
	}

	page := NewPage(config)

	r.HandleFunc("/album/{id}", page.renderAlbum)
	r.HandleFunc("/collection/{query}", page.renderCollection)
	r.HandleFunc("/photo/{id}", page.renderIndex)
	r.HandleFunc("/", page.renderIndex)

	return r
}
