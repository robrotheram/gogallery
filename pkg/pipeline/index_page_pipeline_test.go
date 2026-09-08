package pipeline

import (
	"testing"

	"gogallery/pkg/datastore"
)

func TestPaginateImages(t *testing.T) {
	images := []datastore.Picture{{Id: "1"}, {Id: "2"}, {Id: "3"}}
	pages := paginateImages(images, 2)
	if len(pages) != 2 || len(pages[0]) != 2 || len(pages[1]) != 1 {
		t.Fatalf("paginateImages() returned unexpected page sizes: %#v", pages)
	}
}

func TestPaginateImagesRejectsInvalidChunkSize(t *testing.T) {
	images := []datastore.Picture{{Id: "1"}}
	if pages := paginateImages(images, 0); pages != nil {
		t.Fatalf("paginateImages() = %#v, want nil", pages)
	}
}
