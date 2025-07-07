package datastore

import (
	"gogallery/pkg/monitor"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DataStore struct {
	DB         *gorm.DB
	Pictures   *PictureCollection
	Albums     *AlbumCollection
	ImageCache *ImageCache
	Monitor    monitor.Monitor
}

func Open(path string, monitor monitor.Monitor) (*DataStore, error) {
	db, err := gorm.Open(sqlite.Open("gogallery.sql.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	return &DataStore{
		DB:         db,
		Pictures:   NewPictureCollection(db),
		Albums:     NewAlbumCollection(db),
		ImageCache: NewImageCache(),
		Monitor:    monitor,
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

func (d *DataStore) Reset() {
	d.Pictures.Reset()
	d.Albums.Reset()
	d.ImageCache.Reset()
}

func (d *DataStore) GetTasks() []monitor.MonitorStat {
	return d.Monitor.GetTasks()
}

func (d *DataStore) NewTask(name string, total int) monitor.MonitorStat {
	return d.Monitor.NewTask(name, total)
}
