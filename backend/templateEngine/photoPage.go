package templateengine

import (
	"io"

	"github.com/robrotheram/gogallery/backend/datastore/models"
)

func RenderPhoto(w io.Writer, pic models.Picture, images []models.Picture, page Page) {
	images = models.SortByTime(images)
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
	page.SEO.SetNameFromPhoto(pic)
	Templates.RenderPage(w, PhotoTemplate, page)
}
