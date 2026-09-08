package datastore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gogallery/pkg/monitor"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const DefaultDatabasePath = "gogallery.sql.db"

type DataStore struct {
	DB         *gorm.DB
	Pictures   *PictureCollection
	Albums     *AlbumCollection
	ImageCache *ImageCache
	Monitor    monitor.Monitor
	scanMu     sync.Mutex
}

func Open(path string, monitor monitor.Monitor) (*DataStore, error) {
	databasePath := strings.TrimSpace(path)
	if databasePath == "" {
		return nil, fmt.Errorf("database path is not configured")
	}
	if strings.ContainsRune(databasePath, '?') {
		return nil, fmt.Errorf("database path must not contain query parameters")
	}
	databasePath = filepath.Clean(databasePath)
	db, err := gorm.Open(sqlite.Open(databasePath+"?_busy_timeout=5000&_journal_mode=WAL"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	opened := false
	defer func() {
		if !opened {
			_ = sqlDB.Close()
		}
	}()
	if err := db.AutoMigrate(&Picture{}, &Album{}); err != nil {
		return nil, err
	}
	if err := os.Chmod(databasePath, 0o600); err != nil {
		return nil, fmt.Errorf("secure database permissions: %w", err)
	}
	for _, sidecar := range []string{databasePath + "-wal", databasePath + "-shm"} {
		if err := os.Chmod(sidecar, 0o600); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("secure database sidecar permissions: %w", err)
		}
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)
	imageCache, err := NewImageCache()
	if err != nil {
		return nil, err
	}

	store := &DataStore{
		DB:         db,
		Pictures:   NewPictureCollection(db),
		Albums:     NewAlbumCollection(db),
		ImageCache: imageCache,
		Monitor:    monitor,
	}
	opened = true
	return store, nil
}

func (d *DataStore) GetLatestAlbum() (string, error) {
	pics, err := d.Pictures.GetFilteredPictures(false)
	if err != nil {
		return "", err
	}
	if len(pics) == 0 {
		return "", nil
	}
	return pics[0].Album, nil
}

func (d *DataStore) Reset() {
	_ = d.Pictures.Reset()
	_ = d.Albums.Reset()
	d.ImageCache.Reset()
}

func (d *DataStore) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *DataStore) GetTasks() []monitor.MonitorStat {
	return d.Monitor.GetTasks()
}

func (d *DataStore) NewTask(name string, total int) monitor.MonitorStat {
	return d.Monitor.NewTask(name, total)
}
