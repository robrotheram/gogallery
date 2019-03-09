package web

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/worker"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
)

func Serve() {
	r := mux.NewRouter()
	r.HandleFunc("/albums", func(w http.ResponseWriter, r *http.Request) {
		al, _ := datastore.Cache.Tables("ALBUM").GetAll() //Query("Album","02")
		sArr := al.([]datastore.Album)
		renderTemplate(w, "albumsPage", sArr, *sArr[1].ProfileIMG)
	})
	r.HandleFunc("/album/{name}", func(w http.ResponseWriter, r *http.Request) {
		size := config.Config.Gallery.ImagesPerPage

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
		max := maxPages(pictures)
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
		})
		album.ProfileIMG = &pictures[0]
		pictures = paginate(pictures, 0, size)
		album.Images = pictures
		if len(album.Images) == 0 {
			return
		}
		renderGalleryTemplate(w, "albumPage", album, *album.ProfileIMG, max)
	})

	r.HandleFunc("/album/{name}/{page}", func(w http.ResponseWriter, r *http.Request) {
		size := config.Config.Gallery.ImagesPerPage

		i, err := strconv.Atoi(mux.Vars(r)["page"])
		if err != nil {
			return
		}
		fmt.Println("ON PAGE: " + mux.Vars(r)["page"])
		if i > 1 {
			i = i - 1
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
		if len(album.Images) == 0 {
			return
		}
		renderGalleryTemplate(w, "albumPage", album, album.Images[0], 1)
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
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
		})

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
			"picture": picture},
			picture)

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
		size := config.Config.Gallery.ImagesPerPage

		pics, _ := datastore.Cache.Tables("PICTURE").GetAll()
		pictures := pics.([]datastore.Picture)
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
		})
		pictures = paginate(pictures, 0, size)
		renderGalleryTemplate(w, "indexPage", pictures, pictures[0], maxPages(pics.([]datastore.Picture)))
	})
	r.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		size := config.Config.Gallery.ImagesPerPage

		i, err := strconv.Atoi(mux.Vars(r)["name"])
		if err != nil {
			return
		}
		// JS Image Infinite scroll assume page starts at 1 not 0 causing a whole page of images to go missing.
		if i > 1 {
			i = i - 1
		}

		pics, _ := datastore.Cache.Tables("PICTURE").GetAll()
		pictures := pics.([]datastore.Picture)
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
		})
		pictures = paginate(pictures, i*size, size)
		if len(pictures) > 0 {
			renderGalleryTemplate(w, "indexPage", pictures, pictures[0], 1)
		} else {
			renderNonSocialTemplate(w, "indexPage", pictures)
		}

	})

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		CacheControlWrapper(http.FileServer(http.Dir("web/"+themePath()+"static")))))

	log.Println("Starting server on port" + config.Config.Server.Port)
	log.Fatal(http.ListenAndServe(config.Config.Server.Port, r))
}
