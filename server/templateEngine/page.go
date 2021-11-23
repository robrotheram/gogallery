package templateengine

import (
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
)

type Page struct {
	Settings config.GalleryConfiguration
	Author   config.AboutConfiguration
	Images   []datastore.Picture
	Body     string
}

func NewPage(config *config.Configuration) Page {
	return Page{
		Settings: config.Gallery,
		Author:   config.About,
	}
}

var te = NewTemplateEgine()

func (p *Page) Handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page := vars["page"]
	if te.Pages[page] == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Page " + page + " Not found"))
		return
	}

	var pics []datastore.Picture
	var filterPics []datastore.Picture
	datastore.Cache.DB.All(&pics)
	for _, pic := range pics {
		if !datastore.IsAlbumInBlacklist(pic.Album) {
			if pic.Meta.Visibility == "PUBLIC" {
				var album datastore.Album
				datastore.Cache.DB.One("Id", pic.Album, &album)
				cleanpic := datastore.Picture{
					Id:         pic.Id,
					Name:       pic.Name,
					Caption:    pic.Caption,
					Album:      pic.Album,
					AlbumName:  album.Name,
					FormatTime: pic.Exif.DateTaken.Format("01-02-2006 15:04:05"),
					Exif:       pic.Exif,
					Meta:       pic.Meta,
				}
				filterPics = append(filterPics, cleanpic)
			}

		}
	}
	sort.Slice(filterPics, func(i, j int) bool {
		return filterPics[i].Exif.DateTaken.Sub(filterPics[j].Exif.DateTaken) > 0
	})
	p.Images = filterPics
	w.Write([]byte(te.RenderPage(page, *p)))
}
