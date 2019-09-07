package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/robrotheram/gogallery/datastore"

	"github.com/CloudyKit/jet"
)

func themePath() string {
	return fmt.Sprintf("themes/%s/", config.Gallery.Theme)
}

//

type M map[string]interface{}

func templateModel(data interface{}, image datastore.Picture, numOfPic int) jet.VarMap {
	vars := make(jet.VarMap)
	vars.Set("name", config.Gallery.Name)
	vars.Set("site", config.Gallery.Url)
	vars.Set("about", M{
		"enable":       config.About.Enable,
		"twitter":      config.About.Twitter,
		"blog":         config.About.Blog,
		"website":      config.About.Website,
		"facebook":     config.About.Facebook,
		"email":        config.About.Email,
		"instagram":    config.About.Instagram,
		"photographer": config.About.Photographer,
		"description":  config.About.Description,
		"photo":        config.About.BackgroundPhoto,
		"profilePhoto": config.About.ProfilePhoto,
		"footer":       template.HTML(config.About.Footer),
	})
	vars.Set("data", data)
	vars.Set("socialImage", "")
	vars.Set("imageWidth", "")
	vars.Set("imageHeight", "")
	vars.Set("imageMaxCount", 0)

	if image.Name != "" {
		vars.Set("socialImage", image.Name)
	}
	if image.Exif.Dimension != "" {
		vars.Set("imageWidth", strings.Split(image.Exif.Dimension, "x")[0])
		vars.Set("imageHeight", strings.Split(image.Exif.Dimension, "x")[1])
	}

	if numOfPic != -1 {
		vars.Set("imageMaxCount", numOfPic)
	}
	return vars
}
func executeTemplate(w http.ResponseWriter, tmpl string, model jet.VarMap) {
	templateName := tmpl + ".jet"
	fmt.Println(templateName)
	t, _ := View.GetTemplate(templateName)
	t.Execute(w, model, nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}, image datastore.Picture) {
	templateName := tmpl + ".jet"
	fmt.Println(templateName)
	t, _ := View.GetTemplate(templateName)
	t.Execute(w, templateModel(data, datastore.Picture{}, -1), nil)
}

var View = jet.NewHTMLSet("themes/default/views")

func renderSettingsTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templateName := tmpl + ".jet"
	fmt.Println(templateName)
	t, _ := View.GetTemplate(templateName)
	t.Execute(w, templateModel(data, datastore.Picture{}, -1), nil)
}

func renderGalleryTemplate(w http.ResponseWriter, tmpl string, data interface{}, image datastore.Picture, numOfPic int) {
	templateName := tmpl + ".jet"
	fmt.Println(templateName)
	t, _ := View.GetTemplate(templateName)

	t.Execute(w, templateModel(data, image, numOfPic), nil)
}
func renderNonSocialTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templateName := tmpl + ".jet"
	fmt.Println(templateName)
	t, _ := View.GetTemplate(templateName)
	t.Execute(w, templateModel(data, datastore.Picture{}, -1), nil)
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

func filterOutAlbum(pictures []datastore.Picture, album string) []datastore.Picture {
	tmp := pictures[:0]
	for _, p := range pictures {
		if p.Album != album {
			tmp = append(tmp, p)
		}
	}
	return tmp
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
