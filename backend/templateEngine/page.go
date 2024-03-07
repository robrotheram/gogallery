package templateengine

import (
	"fmt"
	"net/http"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore/models"
)

type PagePicture struct {
	models.Picture
	OrginalImgPath string
}

type Page struct {
	Settings      config.GalleryConfiguration
	SEO           SocailSEO
	Author        config.AboutConfiguration
	Images        []models.Picture
	Albums        models.AlbumStrcure
	Album         models.Album
	LatestAlbum   string
	Picture       PagePicture
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

func (s *SocailSEO) SetImage(picture models.Picture) {
	s.ImageUrl = fmt.Sprintf("%s/img/%s", config.Config.Gallery.Url, picture.Id)
	s.ImageWidth = 1024
	s.ImageHeight = 683
}

func (s *SocailSEO) SetNameFromPhoto(picture models.Picture) {
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

func NewPage(r *http.Request, albumID string) Page {
	page := Page{
		Settings:    config.Config.Gallery,
		Author:      config.Config.About,
		LatestAlbum: albumID,
		ImgSizes:    ImageSizes,
	}
	if r != nil {
		page.SEO = NewSocailSEO(r.URL.EscapedPath())
		page.PagePath = r.URL.EscapedPath()
	}
	return page
}

func NewPagePicture(pic models.Picture) PagePicture {
	originalPath := fmt.Sprintf("/img/%s/xlarge.webp", pic.Id)
	if config.Config.Gallery.UseOriginal {
		originalPath = fmt.Sprintf("/img/%s/original%s", pic.Id, pic.Ext)
	}
	return PagePicture{
		Picture:        pic,
		OrginalImgPath: originalPath,
	}
}
