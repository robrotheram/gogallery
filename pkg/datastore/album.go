package datastore

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testingFyne/pkg/config"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Album struct {
	Id         string    `gorm:"primaryKey;size:64" json:"id"`
	Name       string    `gorm:"column:name" json:"name"`
	ModTime    time.Time `gorm:"column:mod_time" json:"mod_time"`
	Parent     string    `gorm:"column:parent" json:"parent"`
	ParentPath string    `gorm:"column:parent_path" json:"parentPath,omitempty"`
	Path       string    `gorm:"column:path" json:"path,omitempty"`
	ProfileId  string    `gorm:"column:profile_id" json:"profile_image"`
}

type AlbumCollection struct {
	DB *gorm.DB
	sync.Mutex
}

func NewAlbumCollection(db *gorm.DB) *AlbumCollection {
	db.AutoMigrate(&Album{}) // Use the correct struct for migration
	albumCollection := &AlbumCollection{DB: db}
	return albumCollection
}

// Save or update an album (upsert by primary key)
func (c *AlbumCollection) Save(album Album) error {
	c.Lock()
	defer c.Unlock()
	return c.DB.Save(&album).Error
}

func (c *AlbumCollection) BatchInsert(albums []Album) error {
	err := c.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}}, // primary key
		UpdateAll: true,                          // update all fields on conflict
	}).CreateInBatches(albums, 100).Error
	return err
}

// Update fields of an album by ID
func (c *AlbumCollection) Update(id string, updates Album) error {
	c.Lock()
	defer c.Unlock()
	return c.DB.Model(&Album{}).Where("id = ?", id).Updates(updates).Error
}

// Get all albums
func (c *AlbumCollection) GetAll() ([]Album, error) {
	var albums []Album
	if err := c.DB.Find(&albums).Error; err != nil {
		return nil, err
	}
	return albums, nil
}

// Find album by ID
func (c *AlbumCollection) FindById(id string) (Album, error) {
	var album Album
	if err := c.DB.First(&album, "id = ?", id).Error; err != nil {
		return album, err
	}
	return album, nil
}

// Find albums by field (simple string fields)
func (c *AlbumCollection) FindByField(field, value string) ([]Album, error) {
	var albums []Album
	if err := c.DB.Where(field+" = ?", value).Find(&albums).Error; err != nil {
		return nil, err
	}
	return albums, nil
}
func (c *AlbumCollection) Reset() error {
	c.Lock()
	defer c.Unlock()
	if err := c.DB.Exec("DELETE FROM albums").Error; err != nil {
		return fmt.Errorf("failed to reset albums: %w", err)
	}
	// Optionally, you can also reset the auto-increment ID
	if err := c.DB.Exec("DELETE FROM sqlite_sequence WHERE name='albums'").Error; err != nil {
		return fmt.Errorf("failed to reset auto-increment ID: %w", err)
	}
	return nil
}

func (c *AlbumCollection) GetAlbumStructure(config config.GalleryConfiguration) AlbumStrcure {
	albums, _ := c.GetAll()
	for _, alb := range albums {
		if IsAlbumInBlacklist(alb.Name) {
			albums = RemoveAlbumFromSlice(albums, alb)
		}
	}
	newalbms := SliceToTree(albums, config.Basepath)
	return newalbms
}

func (c *AlbumCollection) FindLatestInAlbum(id string) (Picture, error) {
	c.Lock()
	defer c.Unlock()
	var pic Picture
	err := c.DB.Model(&Picture{}).
		Where("album = ?", id).
		Order("date_taken DESC").
		First(&pic).Error
	if err != nil {
		return pic, err
	}
	return pic, nil
}

func (c *AlbumCollection) GetLatestAlbums() ([]Album, error) {
	albums, err := c.GetAll()
	if err != nil {
		return nil, err
	}
	for i, album := range albums {
		if pic, err := c.FindLatestInAlbum(album.Id); err == nil {
			albums[i].ModTime = pic.DateTaken
		} else {
			albums[i].ModTime = time.Time{} // Set to zero time if no picture found
		}
	}
	sort.Slice(albums, func(i, j int) bool {
		return albums[i].ModTime.After(albums[j].ModTime)
	})
	return albums, err
}

func (a *AlbumCollection) MovePictureToAlbum(picture Picture, newAlbum string) error {
	album, _ := a.FindById(newAlbum)
	newName := fmt.Sprintf("%s/%s/%s%s", album.ParentPath, album.Name, picture.Name, filepath.Ext(picture.Path))
	err := os.Rename(picture.Path, newName)
	if err != nil {
		return err
	}
	picture.Path = newName
	picture.Album = newAlbum
	return nil
}

// RemoveAlbumFromSlice removes the first occurrence of the album with the same ID from the slice
func RemoveAlbumFromSlice(albums []Album, target Album) []Album {
	for i, alb := range albums {
		if alb.Id == target.Id {
			return append(albums[:i], albums[i+1:]...)
		}
	}
	return albums
}
