package datastore

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testingFyne/pkg/config"
	"time"

	"gorm.io/gorm"
)

func createPicture(name string, path string, album string) Picture {
	p := Picture{
		Id:           config.GetMD5Hash(path),
		Name:         name,
		Path:         path,
		Ext:          filepath.Ext(path),
		Album:        config.GetMD5Hash(filepath.Dir(path)),
		AlbumName:    album,
		RootPath:     config.Config.Gallery.Basepath,
		PostedToIG:   false,
		Visibility:   "PUBLIC",
		DateAdded:    time.Now(),
		DateModified: time.Now(),
	}
	// p.CreateExif()
	return p
}

func createAlbum(fInfo os.FileInfo, path string) Album {
	info := FileInfoFromInterface(fInfo)
	albumId := config.GetMD5Hash(path)
	return Album{
		Id:         albumId,
		Name:       info.Name,
		ModTime:    info.ModTime,
		Parent:     filepath.Base(filepath.Dir(path)),
		ParentPath: (filepath.Dir(path)),
	}
}

type albumUpdate struct {
	AlbumId   string
	ProfileId string
}

func (db *DataStore) updateExif(pics []Picture) {
	stat := db.Monitor.NewTask("Update Exif Data", len(pics))
	defer stat.Complete()
	workerCount := runtime.NumCPU()
	jobs := make(chan *Picture, len(pics))
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pic := range jobs {
				pic.CreateExif()
				stat.Update()
			}
		}()
	}

	// Send jobs to workers
	for i := range pics {
		jobs <- &pics[i]
	}
	close(jobs)
	wg.Wait()
}

func (db *DataStore) ScanPath(path string) error {
	cfg := config.Config.Gallery

	log.Println("Scanning Folders at:" + path)
	absRoot, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	pictures, albums, albumUpdates, err := db.walkPath(absRoot, cfg)
	if err != nil {
		log.Printf("Error walking path %s: %v", absRoot, err)
		return err
	}

	log.Println("Start processing Exif data for pictures")
	db.updateExif(pictures)
	// log.Println("Exif data processing complete")

	db.Pictures.BatchInsert(pictures)
	db.Albums.BatchInsert(albums)
	db.updateAlbumProfiles(albumUpdates)

	// log.Println("Scanning Complete")
	return nil
}

func (db *DataStore) walkPath(absRoot string, cfg config.GalleryConfiguration) ([]Picture, []Album, []albumUpdate, error) {
	pictures := []Picture{}
	albums := []Album{}
	albumUpdates := []albumUpdate{}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if CheckEXT(path) && !info.IsDir() {
			db.processFile(path, info, &pictures, &albumUpdates)
		}
		if info.IsDir() {
			db.processDirectory(path, info, cfg, &albums)
		}
		return nil
	}

	err := filepath.Walk(absRoot, walkFunc)
	return pictures, albums, albumUpdates, err
}

func (db *DataStore) processFile(path string, info os.FileInfo, pictures *[]Picture, albumUpdates *[]albumUpdate) {
	albumName := filepath.Base(filepath.Dir(path))
	picName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
	if !IsAlbumInBlacklist(albumName) && !IsPictureInBlacklist(picName) {
		p := createPicture(picName, path, albumName)
		*pictures = append(*pictures, p)
		*albumUpdates = append(*albumUpdates, albumUpdate{
			AlbumId:   p.Album,
			ProfileId: p.Id,
		})
	}
}

func (db *DataStore) processDirectory(path string, info os.FileInfo, cfg config.GalleryConfiguration, albums *[]Album) {
	if !IsAlbumInBlacklist(info.Name()) {
		if filepath.Base(filepath.Dir(path)) != cfg.Basepath {
			*albums = append(*albums, createAlbum(info, path))
		}
	}
}

func (db *DataStore) updateAlbumProfiles(albumUpdates []albumUpdate) {
	if len(albumUpdates) > 0 {
		db.Albums.DB.Transaction(func(tx *gorm.DB) error {
			for _, au := range albumUpdates {
				if err := tx.Model(&Album{}).Where("id = ?", au.AlbumId).Update("profile_id", au.ProfileId).Error; err != nil {
					log.Printf("Album profile_id update failed for %s: %v", au.AlbumId, err)
				}
			}
			return nil
		})
	}
}
