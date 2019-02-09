package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	fmt.Println(config.Config.Gallery)
	templates().ExecuteTemplate(w, tmpl, M{
		"name":     config.Config.Gallery.Name,
		"twitter":  config.Config.Gallery.Twitter,
		"facebook": config.Config.Gallery.Facebook,
		"email":    config.Config.Gallery.Email,
		"about":    template.HTML(config.Config.Gallery.About),
		"footer":   template.HTML(config.Config.Gallery.Footer),
		"data":     data})
}

func Serve() {
	r := mux.NewRouter()
	r.HandleFunc("/albums", func(w http.ResponseWriter, r *http.Request) {
		al, _ := datastore.Cache.Tables("ALBUM").GetAll() //Query("Album","02")
		sArr := al.([]datastore.Album)
		for i := range sArr {
			albm := &sArr[i]
			pic, err := datastore.Cache.Tables("PICTURE").Query("Album", albm.Name, 1)
			if err == nil {
				picArr := pic.([]datastore.Picture)
				if len(picArr) > 0 {
					fmt.Println(picArr[0])
					albm.ProfileIMG = &picArr[0]
				}
			}
		}

		//tree := datastore.BuildTree(sArr)
		fmt.Println(sArr)
		renderTemplate(w, "albumsPage", sArr)
	})
	r.HandleFunc("/album/{name}", func(w http.ResponseWriter, r *http.Request) {
		//vars := mux.Vars(r)
		//title := vars["title"]

		vars := mux.Vars(r)
		name := vars["name"]
		albm, err := datastore.Cache.Tables("ALBUM").Query("Name", name, 1)
		if err != nil {
			return
		}
		albums := albm.([]datastore.Album)
		album := &albums[0]
		pics, err := datastore.Cache.Tables("PICTURE").Query("Album", name, 0)
		if err != nil {
			return
		}
		pictures := pics.([]datastore.Picture)
		album.Images = pictures
		album.ProfileIMG = &pictures[0]

		fmt.Println(pictures[0])
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
		picture.FormatTime = picture.ModTime.Format("01-02-2006 15:04:05")

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
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pics, _ := datastore.Cache.Tables("PICTURE").GetAll()
		renderTemplate(w, "indexPage", pics.([]datastore.Picture))
	})

	r.HandleFunc("/img/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		pic, err := datastore.Cache.Tables("PICTURE").Query("Name", name, 1)
		if err == nil {
			picArr := pic.([]datastore.Picture)
			if len(picArr) > 0 {
				fmt.Println(picArr[0])
				http.ServeFile(w, r, picArr[0].Path)
			}
		}
	})

	r.HandleFunc("/thumb/{name}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		vars := mux.Vars(r)
		name := vars["name"]
		pic, err := datastore.Cache.Tables("PICTURE").Query("Name", name, 1)
		if err == nil {
			picArr := pic.([]datastore.Picture)
			if len(picArr) == 0 {
				return
			}
			file, err := os.Open(picArr[0].Path)
			if err != nil {
				fmt.Println(picArr[0].Path)
				return
			}
			// decode jpeg into image.Image
			extension := filepath.Ext(picArr[0].Path)
			var img image.Image
			var img_err error
			switch extension {
			case ".jpg":
				img, img_err = jpeg.Decode(file)
			case ".png":
				img, img_err = png.Decode(file)
			}
			if img_err != nil {
				fmt.Println(picArr[0].Path)
				return
			}
			file.Close()
			// resize to width 1000 using Lanczos resampling
			// and preserve aspect ratio
			m := resize.Resize(300, 0, img, resize.Lanczos3)
			jpeg.Encode(w, m, nil)
		}
	})

	r.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeManifest())
	})

	r.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/"+themePath()+"static/js/sw.js")
	})

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		CacheControlWrapper(http.FileServer(http.Dir("web/"+themePath()+"static")))))

	log.Println("Starting server on port" + config.Config.Server.Port)
	log.Fatal(http.ListenAndServe(config.Config.Server.Port, r))
}
