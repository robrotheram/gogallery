package datastore

import (
	"fmt"
	"gogallery/pkg/config"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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
	return &AlbumCollection{DB: db}
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
	field = strings.ToLower(strings.TrimSpace(field))
	allowed := map[string]struct{}{
		"id": {}, "name": {}, "parent": {}, "parent_path": {}, "path": {},
	}
	if _, ok := allowed[field]; !ok {
		return nil, fmt.Errorf("unsupported album field %q", field)
	}
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

func (c *AlbumCollection) GetAlbumStructure(config config.GalleryConfiguration) (AlbumStructure, error) {
	albums, err := c.GetAll()
	if err != nil {
		return nil, err
	}
	profiles, publicIDs, err := c.publicProfiles()
	if err != nil {
		return nil, err
	}
	filtered := make([]Album, 0, len(albums))
	for _, alb := range albums {
		if !IsAlbumInBlacklist(alb.Name) && pathWithinRoot(alb.Path, config.Basepath) {
			if _, public := publicIDs[alb.ProfileId]; !public {
				alb.ProfileId = profiles[alb.Id].Id
			}
			filtered = append(filtered, alb)
		}
	}
	newalbms := SliceToTree(filtered, config.Basepath)
	return newalbms, nil
}

func (c *AlbumCollection) publicProfiles() (map[string]Picture, map[string]struct{}, error) {
	var pictures []Picture
	if err := c.DB.Select("id", "name", "path", "album", "album_name", "date_taken", "visibility").
		Where("visibility = ?", "PUBLIC").Order("date_taken DESC").Find(&pictures).Error; err != nil {
		return nil, nil, err
	}
	profiles := make(map[string]Picture)
	publicIDs := make(map[string]struct{})
	for _, picture := range pictures {
		if !IsPicturePublishable(picture) {
			continue
		}
		publicIDs[picture.Id] = struct{}{}
		if _, exists := profiles[picture.Album]; !exists {
			profiles[picture.Album] = picture
		}
	}
	return profiles, publicIDs, nil
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
	profiles, publicIDs, err := c.publicProfiles()
	if err != nil {
		return nil, err
	}
	visible := make([]Album, 0, len(albums))
	for _, album := range albums {
		if IsAlbumInBlacklist(album.Name) {
			continue
		}
		pic, exists := profiles[album.Id]
		if !exists {
			continue
		}
		if _, public := publicIDs[album.ProfileId]; !public {
			album.ProfileId = pic.Id
		}
		album.ModTime = pic.DateTaken
		visible = append(visible, album)
	}
	sort.Slice(visible, func(i, j int) bool {
		return visible[i].ModTime.After(visible[j].ModTime)
	})
	return visible, nil
}

func (a *AlbumCollection) MovePictureToAlbum(picture Picture, newAlbum string) error {
	album, err := a.FindById(newAlbum)
	if err != nil {
		return fmt.Errorf("find destination album: %w", err)
	}
	if !pathWithinRoot(album.Path, config.Config.Gallery.Basepath) ||
		!pathWithinRoot(picture.Path, config.Config.Gallery.Basepath) {
		return fmt.Errorf("source or destination is outside the gallery root")
	}
	newPath := filepath.Join(album.Path, filepath.Base(picture.Path))
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("destination picture already exists: %s", newPath)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(picture.Path, newPath); err != nil {
		return err
	}
	if err := a.DB.Model(&Picture{}).Where("id = ?", picture.Id).Updates(map[string]any{
		"path":       newPath,
		"album":      newAlbum,
		"album_name": album.Name,
	}).Error; err != nil {
		_ = os.Rename(newPath, picture.Path)
		return fmt.Errorf("update moved picture: %w", err)
	}
	return nil
}

func pathWithinRoot(path, root string) bool {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(root) == "" {
		return false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	if resolvedRoot, err := filepath.EvalSymlinks(absRoot); err == nil {
		absRoot = resolvedRoot
	} else {
		return false
	}
	if resolvedPath, err := filepath.EvalSymlinks(absPath); err == nil {
		absPath = resolvedPath
	} else {
		resolvedParent, parentErr := filepath.EvalSymlinks(filepath.Dir(absPath))
		if parentErr != nil {
			return false
		}
		absPath = filepath.Join(resolvedParent, filepath.Base(absPath))
	}
	relative, err := filepath.Rel(absRoot, absPath)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
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

func (c *AlbumCollection) RemoveInvalidAlbums() error {
	c.Lock()
	defer c.Unlock()

	albums, err := c.GetAll()
	if err != nil {
		return fmt.Errorf("load albums for cleanup: %w", err)
	}

	for _, album := range albums {
		info, statErr := os.Lstat(album.Path)
		if statErr != nil && !os.IsNotExist(statErr) {
			return fmt.Errorf("inspect album %q: %w", album.Path, statErr)
		}
		invalid := os.IsNotExist(statErr) || info.Mode()&os.ModeSymlink != 0 ||
			!pathWithinRoot(album.Path, config.Config.Gallery.Basepath)
		if invalid {
			if err := c.DB.Delete(&Album{}, "id = ?", album.Id).Error; err != nil {
				return fmt.Errorf("delete invalid album %q: %w", album.Path, err)
			}
		}
	}
	return nil
}
