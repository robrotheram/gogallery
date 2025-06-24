package datastore

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/robrotheram/gogallery/backend/config"
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
