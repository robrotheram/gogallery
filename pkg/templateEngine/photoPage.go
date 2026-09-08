package templateengine

import (
	"gogallery/pkg/datastore"
	"io"
)

func RenderPhoto(w io.Writer, pic datastore.Picture, images []datastore.Picture, page Page) error {
	for i, p := range images {
		if p.Id == pic.Id {
			if i-1 >= 0 {
				page.PreImagePath = images[i-1].Id
			}
			if i+1 < len(images) {
				page.NextImagePath = images[i+1].Id
			}
		}
	}
	page.Picture = NewPagePicture(pic)
	page.SEO.SetFromPhoto(pic)
	return Templates.RenderPage(w, PhotoTemplate, page)
}
