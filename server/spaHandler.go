package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/robrotheram/gogallery/datastore"
)

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath    *packr.Box
	indexTemplate *template.Template
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	//path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, e := h.staticPath.Find(path)
	if e != nil {
		//if os.IsNotExist(err) {
		// file does not exist, serve index.html+

		h.indexTemplate.Execute(w, getTemplateData(r.URL))
		// index, _ := h.staticPath.Find(h.indexPath)
		// w.Write(index)
		return
	}
	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(h.staticPath).ServeHTTP(w, r)
}

type M map[string]interface{}

func getTemplateData(url *url.URL) map[string]interface{} {
	model := M{
		"name":        Config.Gallery.Name,
		"site":        cleanURL(url),
		"description": Config.About.Description,
		"imageWidth":  1024,
		"imageHeight": 683,
	}
	urls := strings.Split(url.String(), "/")
	if len(urls) >= 3 {
		switch gtype := urls[1]; gtype {
		case "photo":
			model["socialImage"] = photoImgURL(urls[2])
		case "album":
			model["socialImage"] = albumImgURL(urls[2])
		default:
			model["socialImage"] = defaultImgURL()
		}
	} else {
		model["socialImage"] = defaultImgURL()
	}
	fmt.Println(url.String())
	return model
}

func cleanURL(url *url.URL) string {
	return fmt.Sprintf("%s://%s%s", url.Scheme, url.Host, url.Path)
}
func albumImgURL(id string) string {
	var album datastore.Album
	datastore.Cache.DB.One("Id", id, &album)
	return fmt.Sprintf("%s/img/%s", Config.Gallery.Basepath, album.ProfileID)
}
func photoImgURL(id string) string {
	var photo datastore.Picture
	datastore.Cache.DB.One("Id", id, &photo)
	return fmt.Sprintf("%s/img/%s", Config.Gallery.Basepath, photo.Id)
}
func defaultImgURL() string {
	return fmt.Sprintf("%s/img/%s", Config.Gallery.Basepath, Config.About.BackgroundPhoto)
}
