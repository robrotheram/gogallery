package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/worker"
	"net/http"
	"os"
	"sort"
	"strconv"
)

func renderAlbum(w http.ResponseWriter, r *http.Request) {
	ViewCount++
	al, _ := datastore.Cache.Tables("ALBUM").GetAll() //Query("Album","02")
	sArr := al.([]datastore.Album)
	if len(sArr) > 1 {
		renderTemplate(w, "albumsPage", sArr, *sArr[1].ProfileIMG)
	}
}

func renderAlbumPage(w http.ResponseWriter, r *http.Request) {
	ViewCount++
	size := config.Gallery.ImagesPerPage

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
}

func renderAlbumPagination(w http.ResponseWriter, r *http.Request) {
	size := config.Gallery.ImagesPerPage

	i, err := strconv.Atoi(mux.Vars(r)["page"])
	if err != nil {
		return
	}
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
}

func renderPicturePage(w http.ResponseWriter, r *http.Request) {
	ViewCount++
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

	model := templateModel(picture, picture, -1)
	model["prePic"] = prePic
	model["nextPic"] = nextPic
	model["picture"] = picture
	templates().ExecuteTemplate(w, "picturePage", model)
}

func loadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	pic, err := datastore.Cache.Tables("PICTURE").Query("Name", name, 1)
	if err == nil {
		picArr := pic.([]datastore.Picture)
		if len(picArr) > 0 {
			http.ServeFile(w, r, picArr[0].Path)
		}
	}
}

func loadThumbnail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=604800") // 7 days
	vars := mux.Vars(r)
	name := vars["name"]
	pic, err := datastore.Cache.Tables("PICTURE").Query("Name", name, 1)
	if err == nil {
		picArr := pic.([]datastore.Picture)
		if len(picArr) == 0 {
			return
		}
		cachePath := fmt.Sprintf("cache/%s.jpg", worker.GetMD5Hash(picArr[0].Path))
		if _, err := os.Stat(cachePath); err == nil {
			http.ServeFile(w, r, cachePath)
		} else {
			http.ServeFile(w, r, cachePath)
			worker.MakeThumbnail(picArr[0].Path)
		}

	}
}
func loadLargeThumbnail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=604800") // 7 days
	vars := mux.Vars(r)
	name := vars["name"]
	pic, err := datastore.Cache.Tables("PICTURE").Query("Name", name, 1)
	if err == nil {
		picArr := pic.([]datastore.Picture)
		if len(picArr) == 0 {
			return
		}
		cachePath := fmt.Sprintf("cache/large_%s.jpg", worker.GetMD5Hash(picArr[0].Path))
		if _, err := os.Stat(cachePath); err == nil {
			http.ServeFile(w, r, cachePath)
		} else {
			http.ServeFile(w, r, cachePath)
			worker.MakeLargeThumbnail(picArr[0].Path)
		}

	}
}

func renderIndexPage(w http.ResponseWriter, r *http.Request) {
	ViewCount++
	size := config.Gallery.ImagesPerPage

	pics, _ := datastore.Cache.Tables("PICTURE").GetAll()
	pictures := pics.([]datastore.Picture)

	if len(pictures) == 0 {
		http.Redirect(w, r, "/about", http.StatusTemporaryRedirect)
		return
	}
	sort.Slice(pictures, func(i, j int) bool {
		return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
	})
	pictures = paginate(pictures, 0, size)

	renderGalleryTemplate(w, "indexPage", pictures, pictures[0], maxPages(pics.([]datastore.Picture)))
}

func renderIndexPaginationPage(w http.ResponseWriter, r *http.Request) {
	size := config.Gallery.ImagesPerPage

	i, err := strconv.Atoi(mux.Vars(r)["name"])
	if err != nil {
		return
	}
	// JS Image Infinite scroll assume page starts at 1 not 0 causing a whole page of images to go missing.
	if i > 1 {
		//i = i - 1
	}

	pics, _ := datastore.Cache.Tables("PICTURE").GetAll()
	pictures := pics.([]datastore.Picture)
	sort.Slice(pictures, func(i, j int) bool {
		return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
	})
	pictures = paginate(pictures, i*size, size)
	if len(pictures) > 0 {
		renderGalleryTemplate(w, "indexPage", pictures, pictures[0], 1)
	}

}
