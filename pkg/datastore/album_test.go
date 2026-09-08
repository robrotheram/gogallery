package datastore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gogallery/pkg/config"
)

func TestAlbumViewsReplacePrivateProfilePicture(t *testing.T) {
	galleryRoot := t.TempDir()
	previousBasepath := config.Config.Gallery.Basepath
	config.Config.Gallery.Basepath = galleryRoot
	t.Cleanup(func() { config.Config.Gallery.Basepath = previousBasepath })
	privatePath := filepath.Join(galleryRoot, "private.jpg")
	publicPath := filepath.Join(galleryRoot, "public.jpg")
	albumPath := filepath.Join(galleryRoot, "Holiday")
	if err := os.Mkdir(albumPath, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{privatePath, publicPath} {
		if err := os.WriteFile(path, []byte("image"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Album{}, &Picture{}); err != nil {
		t.Fatal(err)
	}
	albums := NewAlbumCollection(db)
	pictures := NewPictureCollection(db)
	album := Album{Id: "album", Name: "Holiday", ParentPath: galleryRoot, Path: albumPath, ProfileId: "private"}
	if err := albums.Save(album); err != nil {
		t.Fatal(err)
	}
	if err := pictures.Save(Picture{
		Id: "private", Album: album.Id, AlbumName: album.Name, Visibility: "PRIVATE", Path: privatePath,
	}); err != nil {
		t.Fatal(err)
	}
	if err := pictures.Save(Picture{
		Id: "public", Album: album.Id, AlbumName: album.Name, Visibility: "PUBLIC", Path: publicPath, DateTaken: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	tree, err := albums.GetAlbumStructure(config.GalleryConfiguration{Basepath: galleryRoot})
	if err != nil {
		t.Fatal(err)
	}
	if got := GetAlbumFromStructure(tree, album.Id).ProfileId; got != "public" {
		t.Fatalf("album profile = %q, want public", got)
	}
	latest, err := albums.GetLatestAlbums()
	if err != nil {
		t.Fatal(err)
	}
	if len(latest) != 1 || latest[0].ProfileId != "public" {
		t.Fatalf("latest albums = %#v", latest)
	}
}

func TestSliceToTreeHandlesNestedAndOutOfRootAlbums(t *testing.T) {
	albums := []Album{
		{Id: "parent", Name: "Parent", ParentPath: "/gallery"},
		{Id: "child", Name: "Child", ParentPath: "/gallery/Parent"},
		{Id: "grandchild", Name: "Grandchild", ParentPath: "/gallery/Parent/Child"},
		{Id: "outside", Name: "Outside", ParentPath: "/somewhere/else"},
	}
	tree := SliceToTree(albums, "/gallery")
	if got := GetAlbumFromStructure(tree, "grandchild"); got.Id != "grandchild" {
		t.Fatalf("nested album was not preserved: %#v", got)
	}
	if got := GetAlbumFromStructure(tree, "outside"); got.Id != "" {
		t.Fatalf("out-of-root album was included: %#v", got)
	}
}
