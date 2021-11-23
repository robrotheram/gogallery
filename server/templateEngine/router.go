package templateengine

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
)

const HomeTemplate = "main"
const AlbumTemplate = "albums"
const CollectionTemplate = "collections"
const PhotoTemplate = "photo"

func (p *Page) renderIndex(w http.ResponseWriter, r *http.Request) {
	p.Images = datastore.GetFilteredPictures()
	p.Albums = datastore.GetAlbumStructure(p.Settings)
	w.Write([]byte(te.RenderPage(HomeTemplate, *p)))
}

func (p *Page) renderAlbum(w http.ResponseWriter, r *http.Request) {
	p.Albums = datastore.GetAlbumStructure(p.Settings)
	w.Write([]byte(te.RenderPage(AlbumTemplate, *p)))
}

func (p *Page) renderAlbumPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	album, err := datastore.GetAlbumByID(id)
	if err != nil {
		w.Write([]byte(te.RenderPage("404", *p)))
		return
	}
	p.Images = datastore.GetPicturesByAlbumID(id)
	p.Album = album
	w.Write([]byte(te.RenderPage(CollectionTemplate, *p)))
}

func (p *Page) renderCollection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]
	if query == "latest" {
		latest := datastore.GetLatestPhotoDate()
		http.Redirect(w, r, "/collection/"+latest.Format("2006-01-02"), http.StatusTemporaryRedirect)
	}
	yourDate, err := time.Parse("2006-01-02", query)
	if err != nil {
		yourDate, err = time.Parse("01-02-2006", query)
	}
	if err != nil {
		w.Write([]byte(te.RenderPage("404", *p)))
		return
	}
	p.Images = datastore.GetPhotosByDate(yourDate)
	w.Write([]byte(te.RenderPage(CollectionTemplate, *p)))
}

func (p *Page) renderPhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	pic, err := datastore.GetPictureByID(id)
	if err != nil {
		w.Write([]byte(te.RenderPage("404", *p)))
		return
	}
	p.Picture = pic
	w.Write([]byte(te.RenderPage(PhotoTemplate, *p)))
}

func InitTemplateRoutes(r *mux.Router, config *config.Configuration) *mux.Router {
	err := te.Load("../templates/beta")
	if err != nil {
		fmt.Printf("there Was an error loading the templates, Error %v", err)
	}
	page := NewPage(config)

	r.HandleFunc("/album/{id}", page.renderAlbumPage)
	r.HandleFunc("/albums", page.renderAlbum)
	r.HandleFunc("/collection/{query}", page.renderCollection)
	r.HandleFunc("/photo/{id}", page.renderPhoto)
	r.HandleFunc("/", page.renderIndex)

	return r
}
