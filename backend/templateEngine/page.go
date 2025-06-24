package templateengine

import (
	"fmt"
	"net/http"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
)

type PagePicture struct {
	datastore.Picture
	OrginalImgPath string
}

type Page struct {
	Settings      config.GalleryConfiguration
	SEO           SocailSEO
	Author        config.AboutConfiguration
	Images        []datastore.Picture
	Albums        datastore.AlbumStrcure
	Album         datastore.AlbumNode
	FeaturedAlbum datastore.AlbumNode
	Picture       PagePicture
	NextImagePath string
	PreImagePath  string
	Body          string
	PagePath      string
	ImgSizes      map[string]ImgSize
}

type SocailSEO struct {
	Site        string
	Title       string
	Description string
	ImageUrl    string
	ImageWidth  int
	ImageHeight int
}
type ImgSize struct {
	MinWidth int // Minimum screen width in pixels for this image source
	ImgWidth int // Recommended image width to generate for this breakpoint
}

var ImageSizes = map[string]ImgSize{
	"xsmall":  {MinWidth: 0, ImgWidth: 360},     // Phones (default)
	"small":   {MinWidth: 480, ImgWidth: 640},   // Small tablets / landscape phones
	"medium":  {MinWidth: 768, ImgWidth: 960},   // Tablets
	"large":   {MinWidth: 1024, ImgWidth: 1280}, // Laptops / small desktops
	"xlarge":  {MinWidth: 1440, ImgWidth: 1600}, // Desktops / large screens
	"2xlarge": {MinWidth: 1920, ImgWidth: 1920}, // Retina / ultra-wide
}

func (s *SocailSEO) SetImage(picture datastore.Picture) {
	s.ImageUrl = fmt.Sprintf("%s/img/%s/xlarge.webp", config.Config.Gallery.Url, picture.Id)
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
		Settings: config.Config.Gallery,
		Author:   config.Config.About,
		ImgSizes: ImageSizes,
	}
	if r != nil {
		page.SEO = NewSocailSEO(r.URL.EscapedPath())
		page.PagePath = r.URL.EscapedPath()
	}
	return page
}

func NewPagePicture(pic datastore.Picture) PagePicture {
	originalPath := fmt.Sprintf("/img/%s/xlarge.webp", pic.Id)
	if config.Config.Gallery.UseOriginal {
		originalPath = fmt.Sprintf("/img/%s/original%s", pic.Id, pic.Ext)
	}
	return PagePicture{
		Picture:        pic,
		OrginalImgPath: originalPath,
	}
}
