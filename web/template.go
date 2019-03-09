package web

import (
	"fmt"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"html/template"
	"net/http"
	"strings"
)

type M map[string]interface{}

func themePath() string {
	return fmt.Sprintf("themes/%s/", config.Config.Gallery.Theme)
}
func templates() *template.Template {
	return template.Must(template.ParseGlob(themePath() + "templates/*"))
}
func templateModel(data interface{}, image datastore.Picture, numOfPic int) map[string]interface{} {
	model := M{
		"name":     config.Config.Gallery.Name,
		"site":     config.Config.Gallery.Url,
		"twitter":  config.Config.Gallery.Twitter,
		"facebook": config.Config.Gallery.Facebook,
		"email":    config.Config.Gallery.Email,
		"about":    template.HTML(config.Config.Gallery.About),
		"footer":   template.HTML(config.Config.Gallery.Footer),
		"data":     data}
	if image.Exif.Dimension != "" {
		model["imageWidth"] = strings.Split(image.Exif.Dimension, "x")[0]
		model["imageHeight"] = strings.Split(image.Exif.Dimension, "x")[1]
		model["socialImage"] = image.Name
	}

	if numOfPic != -1 {
		model["imageMaxCount"] = numOfPic
	}
	return model
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}, image datastore.Picture) {
	templates().ExecuteTemplate(w, tmpl, templateModel(data, datastore.Picture{}, -1))
}
func renderGalleryTemplate(w http.ResponseWriter, tmpl string, data interface{}, image datastore.Picture, numOfPic int) {
	templates().ExecuteTemplate(w, tmpl, templateModel(data, image, numOfPic))
}
func renderNonSocialTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templates().ExecuteTemplate(w, tmpl, templateModel(data, datastore.Picture{}, -1))
}

func CacheControlWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		h.ServeHTTP(w, r)
	})
}

func maxPages(x []datastore.Picture) int {
	size := config.Config.Gallery.ImagesPerPage
	if len(x) < size {
		return 0
	}
	page := len(x) / size
	if len(x)%size > 0 {
		page++
	}
	page++

	return page
}

func paginate(x []datastore.Picture, skip int, size int) []datastore.Picture {
	limit := func() int {
		if skip+size > len(x) {
			return len(x)
		} else {
			return skip + size
		}

	}
	start := func() int {
		if skip > len(x) {
			return len(x)
		} else {
			return skip
		}

	}
	return x[start():limit()]
}
