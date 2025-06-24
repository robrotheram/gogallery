package datastore

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robrotheram/gogallery/backend/config"
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
	CreateExif(&p)
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

func (db *DataStore) ScanPath(path string) error {
	cfg := config.Config.Gallery

	log.Println("Scanning Folders at:" + path)
	absRoot, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	pictures := []Picture{}
	albums := []Album{}
	albumUpdates := []albumUpdate{}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if CheckEXT(path) && !info.IsDir() {
			albumName := filepath.Base(filepath.Dir(path))
			picName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			if !IsAlbumInBlacklist(albumName) && !IsPictureInBlacklist(picName) {
				p := createPicture(picName, path, albumName)
				pictures = append(pictures, p)
				albumUpdates = append(albumUpdates, albumUpdate{
					AlbumId:   p.Album,
					ProfileId: p.Id,
				})
			}
		}
		if info.IsDir() {
			if !IsAlbumInBlacklist(info.Name()) {
				if filepath.Base(filepath.Dir(path)) != cfg.Basepath {
					albums = append(albums, createAlbum(info, path))
				}
			}
		}
		return nil
	}
	err = filepath.Walk(absRoot, walkFunc)
	db.Pictures.BatchInsert(pictures)
	db.Albums.BatchInsert(albums)

	// Upsert album updates in a single transaction for performance
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

	log.Println("Scanning Complete")
	return err
}
