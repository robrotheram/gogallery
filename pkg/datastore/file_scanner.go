package datastore

import (
	"errors"
	"fmt"
	"gogallery/pkg/config"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

func createPicture(name string, path string, album string, modified time.Time) Picture {
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
		DateModified: modified,
		DateTaken:    modified,
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
		Path:       path,
	}
}

type albumUpdate struct {
	AlbumId   string
	ProfileId string
}

func (db *DataStore) updateExif(pics []Picture) {
	stat := db.Monitor.NewTask("Update Exif Data", len(pics))
	defer stat.Complete()
	workerCount := runtime.NumCPU() / 2
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > 4 {
		workerCount = 4
	}
	jobs := make(chan *Picture, len(pics))
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pic := range jobs {
				_ = pic.CreateExif() // EXIF is optional; unsupported images remain usable.
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
	db.scanMu.Lock()
	defer db.scanMu.Unlock()

	cfg := config.Config.Gallery

	absRoot, err := scanRoot(path)
	if err != nil {
		return err
	}
	log.Println("Scanning folders at:", absRoot)

	pictures, albums, albumUpdates, err := db.walkPath(absRoot, cfg)
	if err != nil {
		log.Printf("Error walking path %s: %v", absRoot, err)
		return err
	}

	existingIDs, err := db.Pictures.IDs()
	if err != nil {
		return fmt.Errorf("load existing picture IDs: %w", err)
	}
	known := make(map[string]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		known[id] = struct{}{}
	}
	newPictures := make([]Picture, 0, len(pictures))
	for _, picture := range pictures {
		if _, exists := known[picture.Id]; !exists {
			newPictures = append(newPictures, picture)
		}
	}
	if len(newPictures) > 0 {
		log.Printf("Processing EXIF data for %d new picture(s)", len(newPictures))
		db.updateExif(newPictures)
		metadata := make(map[string]Picture, len(newPictures))
		for _, picture := range newPictures {
			metadata[picture.Id] = picture
		}
		for i := range pictures {
			if picture, ok := metadata[pictures[i].Id]; ok {
				pictures[i] = picture
			}
		}
	}
	// log.Println("Exif data processing complete")

	if len(pictures) > 0 {
		if err := db.Pictures.BatchInsert(pictures); err != nil {
			return fmt.Errorf("save pictures: %w", err)
		}
	}
	if len(albums) > 0 {
		if err := db.Albums.BatchInsert(albums); err != nil {
			return fmt.Errorf("save albums: %w", err)
		}
	}
	if err := db.updateAlbumProfiles(albumUpdates); err != nil {
		return err
	}
	if err := db.PerformCleanup(); err != nil {
		return err
	}

	// log.Println("Scanning Complete")
	return nil
}

func scanRoot(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("gallery base path is not configured; set it in Settings")
	}

	absRoot, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve gallery base path: %w", err)
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return "", fmt.Errorf("open gallery base path %q: %w", absRoot, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("gallery base path %q is not a directory", absRoot)
	}
	return absRoot, nil
}

func (db *DataStore) walkPath(absRoot string, cfg config.GalleryConfiguration) ([]Picture, []Album, []albumUpdate, error) {
	pictures := []Picture{}
	albums := []Album{}
	albumUpdates := []albumUpdate{}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Symlinks are not gallery content. Ignoring them prevents a gallery tree
		// from publishing files that resolve outside its configured root.
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if CheckEXT(path) && !info.IsDir() {
			db.processFile(path, info, &pictures, &albumUpdates)
		}
		if info.IsDir() {
			db.processDirectory(path, info, absRoot, cfg, &albums)
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
		p := createPicture(picName, path, albumName, info.ModTime())
		*pictures = append(*pictures, p)
		*albumUpdates = append(*albumUpdates, albumUpdate{
			AlbumId:   p.Album,
			ProfileId: p.Id,
		})
	}
}

func (db *DataStore) processDirectory(path string, info os.FileInfo, root string, cfg config.GalleryConfiguration, albums *[]Album) {
	if !IsAlbumInBlacklist(info.Name()) {
		if filepath.Clean(path) != filepath.Clean(root) {
			*albums = append(*albums, createAlbum(info, path))
		}
	}
}

func (db *DataStore) updateAlbumProfiles(albumUpdates []albumUpdate) error {
	if len(albumUpdates) > 0 {
		if err := db.Albums.DB.Transaction(func(tx *gorm.DB) error {
			for _, au := range albumUpdates {
				if err := tx.Model(&Album{}).Where("id = ?", au.AlbumId).Update("profile_id", au.ProfileId).Error; err != nil {
					return fmt.Errorf("update album profile %s: %w", au.AlbumId, err)
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (db *DataStore) PerformCleanup() error {
	log.Println("Performing cleanup of old pictures and albums")
	if err := db.Pictures.RemoveInvalidPictures(); err != nil {
		return err
	}
	if err := db.Albums.RemoveInvalidAlbums(); err != nil {
		return err
	}
	log.Println("Cleanup complete")
	return nil
}
