package web

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/robrotheram/gogallery/datastore"
	"html/template"
	"net/http"
	"strings"
)

type M map[string]interface{}

func themePath() string {
	return fmt.Sprintf("themes/%s/", config.Gallery.Theme)
}
func templates() *template.Template {
	return template.Must(template.ParseGlob(themePath() + "templates/*"))
}
func templateModel(data interface{}, image datastore.Picture, numOfPic int) map[string]interface{} {

	model := M{
		"name": config.Gallery.Name,
		"site": config.Gallery.Url,
		"about": M{
			"enable":       config.About.Enable,
			"twitter":      config.About.Twitter,
			"blog":         config.About.Blog,
			"website":      config.About.Website,
			"facebook":     config.About.Facebook,
			"email":        config.About.Email,
			"instagram":    config.About.Instagram,
			"photographer": config.About.Photographer,
			"photo":        config.About.BackgroundPhoto,
			"profilePhoto": config.About.ProfilePhoto,
			"footer":       template.HTML(config.About.Footer),
		},
		"data": data}

	if image.Name != "" {
		model["socialImage"] = image.Name
	}
	if image.Exif.Dimension != "" {
		model["imageWidth"] = strings.Split(image.Exif.Dimension, "x")[0]
		model["imageHeight"] = strings.Split(image.Exif.Dimension, "x")[1]
	}

	if numOfPic != -1 {
		model["imageMaxCount"] = numOfPic
	}
	return model
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}, image datastore.Picture) {
	templates().ExecuteTemplate(w, tmpl, templateModel(data, datastore.Picture{}, -1))
}

func renderSettingsTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	results := structs.Map(config)
	results["stats"] = structs.Map(data)
	templates().ExecuteTemplate(w, tmpl, templateModel(results, datastore.Picture{}, -1))
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
	size := config.Gallery.ImagesPerPage
	if len(x) < size {
		return 0
	}
	page := len(x) / size
	if len(x)%size > 0 {
		page++
	}
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
