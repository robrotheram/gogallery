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
	images := datastore.GetFilteredPictures(false)
	page.Images = images
	page.Albums = datastore.Sort(datastore.GetAlbumStructure(page.Settings))
	if len(images) > 0 {
		page.SEO.SetImage(images[0])
	}
	w.Write([]byte(Templates.RenderPage(HomeTemplate, page)))
}

func RenderAlbums(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
	page.Albums = datastore.Sort(datastore.GetAlbumStructure(page.Settings))
	w.Write([]byte(Templates.RenderPage(AlbumTemplate, page)))
}

func RenderAlbumPage(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
	vars := mux.Vars(r)
	id := vars["id"]
	albums := datastore.GetAlbumStructure(page.Settings)
	album := datastore.GetAlbumFromStructure(albums, id)

	if album.Id == "" || datastore.IsAlbumInBlacklist(album.Name) {
		w.Write([]byte(Templates.RenderPage("404", page)))
		return
	}

	images := datastore.GetPicturesByAlbumID(id)
	page.Images = datastore.SortByTime(images)
	page.Album = album
	page.Picture, _ = datastore.GetPictureByID(album.ProfileID)
	page.SEO.Description = fmt.Sprintf("Album: %s", album.Name)

	if len(images) > 0 {
		page.SEO.SetImage(images[0])
	}
	w.Write([]byte(Templates.RenderPage(CollectionTemplate, page)))
}

func RenderAlbumPhoto(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
	vars := mux.Vars(r)
	albumID := vars["album"]
	id := vars["id"]
	album, err := datastore.GetAlbumByID(albumID)
	if album.Id == "" || err != nil || datastore.IsAlbumInBlacklist(album.Name) {
		w.Write([]byte(Templates.RenderPage("404", page)))
		return
	}
	images := datastore.GetPicturesByAlbumID(albumID)
	RenderPhoto(w, id, images, page)
}

func RenderCollection(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
	vars := mux.Vars(r)
	query := vars["query"]
	if query == "latest" {
		latest := datastore.GetLatestPhotoDate()
		http.Redirect(w, r, fmt.Sprintf("/collection/%s/", latest.Format("2006-01-02")), http.StatusTemporaryRedirect)
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
	page.Images = datastore.SortByTime(images)

	if len(images) > 0 {
		page.SEO.SetImage(images[0])
	}
	w.Write([]byte(Templates.RenderPage(CollectionTemplate, page)))
}

func RenderCollectionPhoto(w http.ResponseWriter, r *http.Request) {
	page := NewPage(r)
	vars := mux.Vars(r)
	query := vars["query"]
	id := vars["id"]

	yourDate, err := time.Parse("2006-01-02", query)
	if err != nil {
		yourDate, err = time.Parse("01-02-2006", query)
	}
	if err != nil {
		w.Write([]byte(Templates.RenderPage("404", page)))
		return
	}

	images := datastore.GetPhotosByDate(yourDate)
	RenderPhoto(w, id, images, page)
}

func CacheControlWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=86400")
		h.ServeHTTP(w, r)
	})
}

func depricatedRedirect(path string) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		http.Redirect(rw, r, fmt.Sprintf("%s/%s/", path, id), http.StatusTemporaryRedirect)
	}
}

func InitTemplateRoutes(r *mux.Router, config *config.Configuration) *mux.Router {
	err := Templates.Load(config.Gallery.Theme)
	if err != nil {
		fmt.Printf("there Was an error loading the templates, Error %v", err)
	}

	assetPath := config.Gallery.Theme
	if assetPath == "default" {
		assetPath = "/tmp/gogallery/theme"
	}
	fs := http.FileServer(http.Dir(assetPath + "/assets/"))

	r.HandleFunc("/albums", RenderAlbums)
	r.HandleFunc("/album/{id}/", RenderAlbumPage)
	r.HandleFunc("/collection/{query}/", RenderCollection)

	r.HandleFunc("/collection/{query}/photo/{id}", RenderCollectionPhoto)
	r.HandleFunc("/album/{album}/photo/{id}", RenderAlbumPhoto)
	r.HandleFunc("/photo/{id}", RenderPhotoHandle)

	//Depricated routes, need to end with /
	r.HandleFunc("/collection/{id}", depricatedRedirect("/collection"))
	r.HandleFunc("/album/{id}", depricatedRedirect("/album"))

	r.HandleFunc("/", RenderIndex)
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets", CacheControlWrapper(fs)))
	return r
}
