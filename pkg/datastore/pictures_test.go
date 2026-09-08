package datastore

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gogallery/pkg/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPictureTagList(t *testing.T) {
	picture := Picture{Tags: "Landscape, golden hour, landscape,  mountains ,"}
	want := []string{"Landscape", "golden hour", "mountains"}

	if got := picture.TagList(); !reflect.DeepEqual(got, want) {
		t.Fatalf("TagList() = %#v, want %#v", got, want)
	}
}

func TestPictureUpdatePersistsEmptyMetadata(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&Picture{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	pictures := NewPictureCollection(db)
	picture := Picture{Id: "photo-id", Name: "Title", Caption: "Description", Tags: "one, two"}
	if err := pictures.Save(picture); err != nil {
		t.Fatalf("save picture: %v", err)
	}

	picture.Caption = ""
	picture.Tags = ""
	if err := pictures.Update(picture.Id, picture); err != nil {
		t.Fatalf("update picture: %v", err)
	}

	updated, err := pictures.FindByID(picture.Id)
	if err != nil {
		t.Fatalf("find updated picture: %v", err)
	}
	if updated.Caption != "" || updated.Tags != "" {
		t.Fatalf("cleared metadata was not persisted: caption=%q tags=%q", updated.Caption, updated.Tags)
	}
}

func TestPictureFindByFieldRejectsUnknownColumn(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	pictures := NewPictureCollection(db)
	if _, err := pictures.FindByField("album = ? OR 1=1 --", "value"); err == nil {
		t.Fatal("expected unsafe field name to be rejected")
	}
}

func TestIsPicturePublicHonoursVisibility(t *testing.T) {
	if IsPicturePublic(Picture{Name: "photo", AlbumName: "album", Visibility: "PRIVATE"}) {
		t.Fatal("private picture was treated as public")
	}
	if !IsPicturePublic(Picture{Name: "photo", AlbumName: "album", Visibility: "PUBLIC"}) {
		t.Fatal("public picture was unexpectedly filtered")
	}
}

func TestIsPicturePublishableRejectsPathOutsideGallery(t *testing.T) {
	galleryRoot := t.TempDir()
	previousBasepath := config.Config.Gallery.Basepath
	config.Config.Gallery.Basepath = galleryRoot
	t.Cleanup(func() { config.Config.Gallery.Basepath = previousBasepath })

	inside := filepath.Join(galleryRoot, "inside.jpg")
	outside := filepath.Join(t.TempDir(), "outside.jpg")
	for _, path := range []string{inside, outside} {
		if err := os.WriteFile(path, []byte("image"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if !IsPicturePublishable(Picture{Name: "inside", AlbumName: "album", Visibility: "PUBLIC", Path: inside}) {
		t.Fatal("picture inside gallery root was rejected")
	}
	if IsPicturePublishable(Picture{Name: "outside", AlbumName: "album", Visibility: "PUBLIC", Path: outside}) {
		t.Fatal("picture outside gallery root was accepted")
	}
}
