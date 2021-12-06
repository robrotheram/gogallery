package templateengine

import (
	"fmt"
	"net/http"

	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
)

type Page struct {
	Settings    config.GalleryConfiguration
	SEO         SocailSEO
	Author      config.AboutConfiguration
	Images      []datastore.Picture
	Albums      datastore.AlbumStrcure
	Album       datastore.Album
	Picture     datastore.Picture
	NextImageID string
	PreImageID  string
	Body        string
	pagePath    string
}

type SocailSEO struct {
	Site        string
	Title       string
	Description string
	ImageUrl    string
	ImageWidth  int
	ImageHeight int
}

func (s *SocailSEO) SetImage(picture datastore.Picture) {
	s.ImageUrl = fmt.Sprintf("%s/img/%s", config.Config.Gallery.Url, picture.Id)
	s.ImageWidth = 1024
	s.ImageHeight = 683
}

func (s *SocailSEO) SetNameFromPhoto(picture datastore.Picture) {
	s.Title = picture.Name
	if picture.Caption != "" {
		s.Description = picture.Caption
	}
	s.SetImage(picture)
}

func NewSocailSEO(path string) SocailSEO {
	return SocailSEO{
		Site:        fmt.Sprintf("%s%s", config.Config.Gallery.Url, path),
		Title:       config.Config.Gallery.Name,
		Description: config.Config.About.Description,
	}
}

func NewPage(r *http.Request) Page {
	return Page{
		Settings: config.Config.Gallery,
		Author:   config.Config.About,
		SEO:      NewSocailSEO(r.URL.EscapedPath()),
		pagePath: r.URL.EscapedPath(),
	}
}
