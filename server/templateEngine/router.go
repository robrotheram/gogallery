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

func RenderIndex(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
	images := datastore.GetFilteredPictures()
	page.Images = images
	page.Albums = datastore.GetAlbumStructure(page.Settings)
	if len(images) > 0 {
		page.SEO.SetImage(images[0])
	}
	w.Write([]byte(Templates.RenderPage(HomeTemplate, page)))
}

func RenderAlbums(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
	page.Albums = datastore.GetAlbumStructure(page.Settings)
	w.Write([]byte(Templates.RenderPage(AlbumTemplate, page)))
}

func RenderAlbumPage(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
	vars := mux.Vars(r)
	id := vars["id"]

	as := datastore.GetAlbumStructure(page.Settings)
	album := datastore.GetAlbumFromStructure(as, id)

	if album.Id == "" {
		w.Write([]byte(Templates.RenderPage("404", page)))
		return
	}

	images := datastore.GetPicturesByAlbumID(id)
	page.Images = images
	page.Album = album
	page.Picture, _ = datastore.GetPictureByID(album.ProfileID)
	if len(images) > 0 {
		page.SEO.SetImage(images[0])
	}
	w.Write([]byte(Templates.RenderPage(CollectionTemplate, page)))
}

func RenderCollection(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
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
		w.Write([]byte(Templates.RenderPage("404", page)))
		return
	}
	images := datastore.GetPhotosByDate(yourDate)
	page.Images = images
	if len(images) > 0 {
		page.SEO.SetImage(images[0])
	}
	w.Write([]byte(Templates.RenderPage(CollectionTemplate, page)))
}

func InitTemplateRoutes(r *mux.Router, config *config.Configuration) *mux.Router {
	fs := http.FileServer(http.Dir(config.Gallery.Theme + "/assets/"))
	err := Templates.Load(config.Gallery.Theme)
	if err != nil {
		fmt.Printf("there Was an error loading the templates, Error %v", err)
	}
	r.HandleFunc("/album/{id}", RenderAlbumPage)
	r.HandleFunc("/albums", RenderAlbums)
	r.HandleFunc("/collection/{query}", RenderCollection)
	r.HandleFunc("/photo/{id}", RenderPhoto)
	r.HandleFunc("/", RenderIndex)
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets", fs))

	return r
}
