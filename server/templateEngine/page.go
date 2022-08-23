package templateengine

import (
	"fmt"
	"net/http"

	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
)

type Page struct {
	Settings      config.GalleryConfiguration
	SEO           SocailSEO
	Author        config.AboutConfiguration
	Images        []datastore.Picture
	Albums        datastore.AlbumStrcure
	Album         datastore.Album
	LatestAlbum   string
	Picture       datastore.Picture
	NextImagePath string
	PreImagePath  string
	Body          string
	PagePath      string
	ImgSizes      map[string]int
}

type SocailSEO struct {
	Site        string
	Title       string
	Description string
	ImageUrl    string
	ImageWidth  int
	ImageHeight int
}

var ImageSizes = map[string]int{
	"xsmall": 350,
	"small":  640,
	"medium": 1024,
	"large":  1600,
	"xlarge": 1920,
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
	page := Page{
		Settings:    config.Config.Gallery,
		Author:      config.Config.About,
		LatestAlbum: datastore.GetLatestAlbumId(),
		ImgSizes:    ImageSizes,
	}
	if r != nil {
		page.SEO = NewSocailSEO(r.URL.EscapedPath())
		page.PagePath = r.URL.EscapedPath()
	}
	return page
}
