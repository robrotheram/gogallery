package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/worker"
	"html/template"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
)

// Shorthand - useful!
type M map[string]interface{}

func themePath() string {
	return fmt.Sprintf("../themes/%s/", config.Config.Gallery.Theme)
}
func templates() *template.Template {
	return template.Must(template.ParseGlob("web/" + themePath() + "templates/*"))
}

func writeImage(w http.ResponseWriter, img *image.Image) {

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		log.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}

func CacheControlWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		h.ServeHTTP(w, r)
	})
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templates().ExecuteTemplate(w, tmpl, M{
		"name":     config.Config.Gallery.Name,
		"twitter":  config.Config.Gallery.Twitter,
		"facebook": config.Config.Gallery.Facebook,
		"email":    config.Config.Gallery.Email,
		"about":    template.HTML(config.Config.Gallery.About),
		"footer":   template.HTML(config.Config.Gallery.Footer),
		"data":     data})
}

var size = 50

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

func Serve() {
	r := mux.NewRouter()
	r.HandleFunc("/albums", func(w http.ResponseWriter, r *http.Request) {
		al, _ := datastore.Cache.Tables("ALBUM").GetAll() //Query("Album","02")
		sArr := al.([]datastore.Album)
		renderTemplate(w, "albumsPage", sArr)
	})
	r.HandleFunc("/album/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		albm, err := datastore.Cache.Tables("ALBUM").Query("Name", name, 1)
		if err != nil {
			return
		}
		albums := albm.([]datastore.Album)
		if len(albums) == 0 {
			return
		}
		album := &albums[0]
		pics, err := datastore.Cache.Tables("PICTURE").Query("Album", name, 0)
		if err != nil {
			return
		}
		pictures := pics.([]datastore.Picture)
		pictures = paginate(pictures, 0, size)
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
		})

		album.Images = pictures
		album.ProfileIMG = &pictures[0]
		renderTemplate(w, "albumPage", album)
	})

	r.HandleFunc("/album/{name}/{page}", func(w http.ResponseWriter, r *http.Request) {
		i, err := strconv.Atoi(mux.Vars(r)["page"])
		if err != nil {
			return
		}

		vars := mux.Vars(r)
		name := vars["name"]
		albm, err := datastore.Cache.Tables("ALBUM").Query("Name", name, 1)
		if err != nil {
			return
		}
		albums := albm.([]datastore.Album)
		if len(albums) == 0 {
			return
		}
		album := &albums[0]
		pics, err := datastore.Cache.Tables("PICTURE").Query("Album", name, 0)
		if err != nil {
			return
		}
		pictures := pics.([]datastore.Picture)
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
		})
		album.ProfileIMG = &pictures[0]
		pictures = paginate(pictures, i*size, size)
		album.Images = pictures
		renderTemplate(w, "albumPage", album)
	})

	r.HandleFunc("/pic/{picture}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["picture"]
		pics, err := datastore.Cache.Tables("PICTURE").Query("Name", name, 1)
		if err != nil {
			return
		}
		if len(pics.([]datastore.Picture)) == 0 {
			return
		}
		picture := pics.([]datastore.Picture)[0]
		picture.FormatTime = picture.Exif.DateTaken.Format("01-02-2006 15:04:05")

		/*Find next and previous picture*/
		pics, err = datastore.Cache.Tables("PICTURE").Query("Album", picture.Album, 0)
		pictures := pics.([]datastore.Picture)
		var nextPic, prePic *datastore.Picture
		for i := range pictures {
			if pictures[i].Name == name {
				if i+1 < len(pictures) {
					nextPic = &pictures[i+1]
				}
				if i != 0 {
					prePic = &pictures[i-1]
				}
				break
			}
		}
		renderTemplate(w, "picturePage", M{
			"prePic":  prePic,
			"nextPic": nextPic,
			"picture": picture})

	})
	r.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeManifest())
	})

	r.HandleFunc("/img/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		pic, err := datastore.Cache.Tables("PICTURE").Query("Name", name, 1)
		if err == nil {
			picArr := pic.([]datastore.Picture)
			if len(picArr) > 0 {
				http.ServeFile(w, r, picArr[0].Path)
			}
		}
	})

	r.HandleFunc("/thumb/{name}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800") // 7 days
		vars := mux.Vars(r)
		name := vars["name"]
		pic, err := datastore.Cache.Tables("PICTURE").Query("Name", name, 1)
		if err == nil {
			picArr := pic.([]datastore.Picture)
			if len(picArr) == 0 {
				return
			}
			cachePath := fmt.Sprintf("cache/%s.jpg", config.GetMD5Hash(picArr[0].Path))
			if _, err := os.Stat(cachePath); err == nil {
				http.ServeFile(w, r, cachePath)
			} else {
				http.ServeFile(w, r, cachePath)
				worker.MakeThumbnail(picArr[0].Path)
			}

		}
	})

	r.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/"+themePath()+"static/js/sw.js")
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pics, _ := datastore.Cache.Tables("PICTURE").GetAll()
		pictures := pics.([]datastore.Picture)
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
		})
		pictures = paginate(pictures, 0, size)
		renderTemplate(w, "indexPage", pictures)
	})
	r.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		i, err := strconv.Atoi(mux.Vars(r)["name"])
		if err != nil {
			return
		}
		pics, _ := datastore.Cache.Tables("PICTURE").GetAll()
		pictures := pics.([]datastore.Picture)
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
		})
		pictures = paginate(pictures, i*size, size)
		renderTemplate(w, "indexPage", pictures)
	})

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		CacheControlWrapper(http.FileServer(http.Dir("web/"+themePath()+"static")))))

	log.Println("Starting server on port" + config.Config.Server.Port)
	log.Fatal(http.ListenAndServe(config.Config.Server.Port, r))
}
