package datastore

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DataStore struct {
	DB         *gorm.DB
	Pictures   *PictureCollection
	Albums     *AlbumCollection
	ImageCache *ImageCache
}

func Open(path string) (*DataStore, error) {
	db, err := gorm.Open(sqlite.Open("gogallery.sql.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &DataStore{
		DB:         db,
		Pictures:   NewPictureCollection(db),
		Albums:     NewAlbumCollection(db),
		ImageCache: NewImageCache(),
	}, nil
}

func (d *DataStore) GetLatestAlbum() string {
	pics := d.Pictures.GetFilteredPictures(false)
	latests := pics[0].DateTaken
	album := pics[0].Album
	for _, p := range pics {
		if p.DateTaken.After(latests) {
			latests = p.DateTaken
			album = p.Album
		}
	}
	return album
}
