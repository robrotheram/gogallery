package templateengine

import (
	"fmt"
	"net/http"
	"strings"

	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
)

type PagePicture struct {
	datastore.Picture
	OriginalImagePath string
}

type Page struct {
	Settings      config.GalleryConfiguration
	SEO           SocialSEO
	Author        config.AboutConfiguration
	Images        []datastore.Picture
	Albums        []datastore.AlbumNode
	Album         datastore.AlbumNode
	FeaturedAlbum datastore.AlbumNode
	Picture       PagePicture
	NextImagePath string
	PreImagePath  string
	Body          string
	PagePath      string
	ImgSizes      map[string]ImgSize
}

type SocialSEO struct {
	Site        string
	Type        string
	Title       string
	Description string
	Keywords    string
	Tags        []string
	ImageUrl    string
	ImageWidth  int
	ImageHeight int
}
type ImgSize struct {
	MinWidth int // Minimum screen width in pixels for this image source
	ImgWidth int // Recommended image width to generate for this breakpoint
}

var ImageSizes = map[string]ImgSize{
	"xsmall": {MinWidth: 0, ImgWidth: 360},     // Phones (default)
	"small":  {MinWidth: 480, ImgWidth: 640},   // Small tablets / landscape phones
	"medium": {MinWidth: 768, ImgWidth: 960},   // Tablets
	"large":  {MinWidth: 1024, ImgWidth: 1280}, // Laptops / small desktops
	"xlarge": {MinWidth: 1440, ImgWidth: 0},    // Large desktops (0 means use original size)
}

func (s *SocialSEO) SetImage(picture datastore.Picture) {
	if picture.Id == "" {
		s.ImageUrl = ""
		s.ImageWidth = 0
		s.ImageHeight = 0
		return
	}
	s.ImageUrl = absoluteURL(fmt.Sprintf("/img/%s/xlarge.webp", picture.Id))
	if _, err := fmt.Sscanf(picture.Dimension, "%dx%d", &s.ImageWidth, &s.ImageHeight); err != nil {
		s.ImageWidth = 0
		s.ImageHeight = 0
	}
}

func (s *SocialSEO) SetFromPhoto(picture datastore.Picture) {
	s.Title = picture.Name
	if picture.Caption != "" {
		s.Description = picture.Caption
	}
	s.Type = "article"
	s.Tags = picture.TagList()
	s.Keywords = strings.Join(s.Tags, ", ")
	s.SetPath(fmt.Sprintf("/photo/%s/", picture.Id))
	s.SetImage(picture)
}

func (s *SocialSEO) SetPath(path string) {
	s.Site = absoluteURL(path)
}

func NewSocialSEO(path string) SocialSEO {
	return SocialSEO{
		Site:        absoluteURL(path),
		Type:        "website",
		Title:       config.Config.Gallery.Name,
		Description: config.Config.About.Description,
	}
}

func absoluteURL(path string) string {
	base := strings.TrimRight(config.Config.Gallery.Url, "/")
	path = "/" + strings.TrimLeft(path, "/")
	if base == "" {
		return path
	}
	return base + path
}

func NewPage(r *http.Request) Page {
	page := Page{
		Settings: config.Config.Gallery,
		Author:   config.Config.About,
		ImgSizes: ImageSizes,
		SEO:      NewSocialSEO("/"),
	}
	if r != nil {
		page.SEO.SetPath(r.URL.EscapedPath())
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
		Picture:           pic,
		OriginalImagePath: originalPath,
	}
}
