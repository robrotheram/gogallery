package web

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/worker"
)

func renderAlbum(w http.ResponseWriter, r *http.Request) {
	if datastore.IsScanning {
		renderSettingsTemplate(w, "errorPage", fmt.Sprintf("Looks like the gallery is rebuilding only got %d photos left to look at", worker.QueSize()))
		return
	}

	ViewCount++
	var albms []datastore.Album
	datastore.Cache.DB.All(&albms)
	if len(albms) > 1 {
		renderTemplate(w, "albumsPage", albms, *albms[1].ProfileIMG)
	}
}

func renderAlbumPage(w http.ResponseWriter, r *http.Request) {
	if datastore.IsScanning {
		renderSettingsTemplate(w, "errorPage", fmt.Sprintf("Looks like the gallery is rebuilding only got %d photos left to look at", worker.QueSize()))
		return
	}

	ViewCount++
	size := config.Gallery.ImagesPerPage

	vars := mux.Vars(r)
	name := vars["name"]

	var album datastore.Album
	var pictures []datastore.Picture
	err := datastore.Cache.DB.Find("Album", name, &pictures)
	if err != nil {
		fmt.Println(err)
		renderSettingsTemplate(w, "errorPage", "No Images For the Album")
		return
	}
	datastore.Cache.DB.One("Name", name, &album)

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
	fmt.Println(album.Name)
	renderGalleryTemplate(w, "albumPage", album, *album.ProfileIMG, max)
}

func renderAlbumPagination(w http.ResponseWriter, r *http.Request) {
	if datastore.IsScanning {
		renderSettingsTemplate(w, "errorPage", fmt.Sprintf("Looks like the gallery is rebuilding only got %d photos left to look at", worker.QueSize()))
		return
	}

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

	var album datastore.Album
	datastore.Cache.DB.One("Name", name, &album)
	var pictures []datastore.Picture
	datastore.Cache.DB.Find("Album", name, &pictures)

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

func renderAlbumPicturePage(w http.ResponseWriter, r *http.Request) {
	if datastore.IsScanning {
		renderSettingsTemplate(w, "errorPage", fmt.Sprintf("Looks like the gallery is rebuilding only got %d photos left to look at", worker.QueSize()))
		return
	}

	ViewCount++
	vars := mux.Vars(r)
	name := vars["picture"]
	var picture datastore.Picture
	datastore.Cache.DB.One("Name", name, &picture)

	picture.FormatTime = picture.Exif.DateTaken.Format("01-02-2006 15:04:05")

	var pictures []datastore.Picture
	err := datastore.Cache.DB.Find("Album", picture.Album, &pictures)
	if err != nil {
		fmt.Println(err)
		renderSettingsTemplate(w, "errorPage", fmt.Sprintf("Looks like this Album has no photos"))
		return
	}

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
	model.Set("prePic", prePic)
	model.Set("nextPic", nextPic)
	model.Set("picture", picture)
	fmt.Println(picture.Exif)
	executeTemplate(w, "albumPicturePage", model)
}

func renderPicturePage(w http.ResponseWriter, r *http.Request) {
	if datastore.IsScanning {
		renderSettingsTemplate(w, "errorPage", fmt.Sprintf("Looks like the gallery is rebuilding only got %d photos left to look at", worker.QueSize()))
		return
	}

	ViewCount++
	vars := mux.Vars(r)
	name := vars["picture"]
	var picture datastore.Picture
	datastore.Cache.DB.One("Name", name, &picture)
	picture.FormatTime = picture.Exif.DateTaken.Format("01-02-2006 15:04:05")

	/*Find next and previous picture*/
	var pictures []datastore.Picture
	err := datastore.Cache.DB.Find("Album", picture.Album, &pictures)
	if err != nil {
		fmt.Println(err)
		renderSettingsTemplate(w, "errorPage", "The seems the be a photo missing")
		return
	}

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
	model.Set("prePic", prePic)
	model.Set("nextPic", nextPic)
	model.Set("picture", picture)
	executeTemplate(w, "albumPicturePage", model)
}

func loadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	var picture datastore.Picture
	datastore.Cache.DB.One("Name", name, &picture)
	http.ServeFile(w, r, picture.Path)
}

func loadThumbnail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=604800") // 7 days
	vars := mux.Vars(r)
	name := vars["name"]
	var picture datastore.Picture
	datastore.Cache.DB.One("Name", name, &picture)

	cachePath := fmt.Sprintf("cache/%s.jpg", worker.GetMD5Hash(picture.Path))
	if _, err := os.Stat(cachePath); err == nil {
		http.ServeFile(w, r, cachePath)
	} else {
		http.ServeFile(w, r, "/static/img/placeholder.png")
		worker.SendToThumbnail(picture.Path)
	}

}
func loadLargeThumbnail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=604800") // 7 days
	vars := mux.Vars(r)
	name := vars["name"]
	var picture datastore.Picture
	datastore.Cache.DB.One("Name", name, &picture)

	cachePath := fmt.Sprintf("cache/%s.jpg", worker.GetMD5Hash(picture.Path))
	if _, err := os.Stat(cachePath); err == nil {
		http.ServeFile(w, r, cachePath)
	} else {
		http.ServeFile(w, r, "/static/img/placeholder.png")
		worker.SendToThumbnail(picture.Path)
	}
}

func renderErrorPage(w http.ResponseWriter, r *http.Request) {
	renderSettingsTemplate(w, "errorPage", fmt.Sprintf("Looks like the gallery is rebuilding only got %d photos left to look at", worker.QueSize()))
}

func renderIndexPage(w http.ResponseWriter, r *http.Request) {
	ViewCount++
	if datastore.IsScanning {
		renderSettingsTemplate(w, "errorPage", fmt.Sprintf("Looks like the gallery is rebuilding only got %d photos left to look at", worker.QueSize()))
		return
	}

	size := config.Gallery.ImagesPerPage

	var pics []datastore.Picture
	datastore.Cache.DB.All(&pics)
	pictures := filterOutAlbum(pics, "instagram")

	if len(pictures) == 0 {
		http.Redirect(w, r, "/about", http.StatusTemporaryRedirect)
		return
	}
	sort.Slice(pictures, func(i, j int) bool {
		return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
	})
	pictures = paginate(pictures, 0, size)
	renderGalleryTemplate(w, "indexPage", pictures, pictures[0], maxPages(pics))
}

func renderIndexPaginationPage(w http.ResponseWriter, r *http.Request) {
	if datastore.IsScanning {
		renderSettingsTemplate(w, "errorPage", fmt.Sprintf("Looks like the gallery is rebuilding only got %d photos left to look at", worker.QueSize()))
		return
	}

	size := config.Gallery.ImagesPerPage

	i, err := strconv.Atoi(mux.Vars(r)["name"])
	if err != nil {
		return
	}
	// JS Image Infinite scroll assume page starts at 1 not 0 causing a whole page of images to go missing.
	if i > 1 {
		//i = i - 1
	}

	var pics []datastore.Picture
	datastore.Cache.DB.All(&pics)

	pictures := filterOutAlbum(pics, "instagram")
	sort.Slice(pictures, func(i, j int) bool {
		return pictures[i].Exif.DateTaken.Sub(pictures[j].Exif.DateTaken) > 0
	})
	pictures = paginate(pictures, i*size, size)
	if len(pictures) > 0 {
		renderGalleryTemplate(w, "indexPage", pictures, pictures[0], 1)
	}

}
