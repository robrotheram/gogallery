package datastore

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"gogallery/pkg/config"

	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PictureCollection struct {
	DB *gorm.DB
	sync.Mutex
}

type Picture struct {
	Id         string `gorm:"primaryKey;size:64" json:"id"`
	Name       string `gorm:"size:255" json:"name"`
	Caption    string `gorm:"size:255" json:"caption"`
	Tags       string `gorm:"size:1024" json:"tags,omitempty"`
	Path       string `gorm:"size:1024" json:"path,omitempty"`
	Ext        string `gorm:"size:32" json:"extension,omitempty"`
	FormatTime string `gorm:"size:64" json:"format_time"`
	Album      string `gorm:"size:64" json:"album"`
	AlbumName  string `gorm:"size:255" json:"album_name"`
	RootPath   string `gorm:"size:1024" json:"root_path,omitempty"`
	// Flattened Exif fields
	FStop        string    `gorm:"size:32" json:"f_stop"`
	FocalLength  string    `gorm:"size:32" json:"focal_length"`
	ShutterSpeed string    `gorm:"size:32" json:"shutter_speed"`
	ISO          string    `gorm:"size:32" json:"iso"`
	Dimension    string    `gorm:"size:32" json:"dimension"`
	AspectRatio  float32   `json:"aspect_ratio"`
	Camera       string    `gorm:"size:255" json:"camera"`
	LensModel    string    `gorm:"size:255" json:"lens_model"`
	DateTaken    time.Time `json:"date_taken"`
	// GPS coordinates
	GPSLat       float64   `json:"gps_latitude"`
	GPSLng       float64   `json:"gps_longitude"`
	GPSAltitude  float64   `json:"gps_altitude,omitempty"`
	GPSTimestamp time.Time `json:"gps_timestamp,omitempty"`
	// Additional metadata
	FileFormat string `gorm:"size:32" json:"file_format"`
	Software   string `gorm:"size:255" json:"software"`
	ColorSpace string `gorm:"size:32" json:"color_space"`

	MeteringMode string `gorm:"size:32" json:"metering_mode"`
	WhiteBalance string `gorm:"size:32" json:"white_balance,omitempty"`
	Saturation   string `gorm:"size:32" json:"saturation,omitempty"`
	Contrast     string `gorm:"size:32" json:"contrast,omitempty"`
	Sharpness    string `gorm:"size:32" json:"sharpness,omitempty"`
	Temperature  string `gorm:"size:32" json:"temperature,omitempty"`

	// Flattened Meta fields
	PostedToIG   bool      `json:"posted_to_IG,omitempty"`
	Visibility   string    `gorm:"size:32" json:"visibility,omitempty"`
	DateAdded    time.Time `json:"date_added,omitempty"`
	DateModified time.Time `json:"date_modified,omitempty"`
}

func NewPictureCollection(db *gorm.DB) *PictureCollection {
	return &PictureCollection{DB: db}
}

func (p *PictureCollection) Save(pic Picture) error {
	p.Lock()
	defer p.Unlock()
	return p.DB.Create(pic).Error
}

func (p *PictureCollection) Reset() error {
	p.Lock()
	defer p.Unlock()
	// Delete all pictures from the database
	if err := p.DB.Exec("DELETE FROM pictures").Error; err != nil {
		return fmt.Errorf("failed to reset pictures: %w", err)
	}
	// Optionally, you can also reset the auto-increment ID
	if err := p.DB.Exec("DELETE FROM sqlite_sequence WHERE name='pictures'").Error; err != nil {
		return fmt.Errorf("failed to reset auto-increment ID: %w", err)
	}
	return nil
}

// Update fields of an album by ID
func (p *PictureCollection) Update(id string, updates Picture) error {
	p.Lock()
	defer p.Unlock()
	if strings.EqualFold(updates.Ext, ".jpg") || strings.EqualFold(updates.Ext, ".jpeg") {
		if err := updates.UpdateExifTags(); err != nil {
			log.Printf("Could not update EXIF metadata for %s: %v", updates.Path, err)
		}
	}
	return p.DB.Model(&Picture{}).Where("id = ?", id).Select("*").Updates(updates).Error
}

func (p *PictureCollection) BatchInsert(pics []Picture) error {
	err := p.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}}, // primary key
		UpdateAll: false,                         // update all fields on conflict
	}).CreateInBatches(pics, 100).Error
	return err
}

// GetAll returns all pictures as domain models
func (p *PictureCollection) GetAll() ([]Picture, error) {
	var dbModels []Picture
	if err := p.DB.Order("date_taken desc").Find(&dbModels).Error; err != nil {
		return nil, err
	}
	return dbModels, nil
}

func (p *PictureCollection) IDs() ([]string, error) {
	var ids []string
	if err := p.DB.Model(&Picture{}).Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// FindByID returns a picture by its ID as a domain model
func (p *PictureCollection) FindByID(id string) (Picture, error) {
	var dbModel Picture
	if err := p.DB.First(&dbModel, "id = ?", id).Error; err != nil {
		return dbModel, err
	}
	return dbModel, nil
}

func (p *PictureCollection) FindLatestInAlbum(album string) (Picture, error) {
	var pic Picture
	err := p.DB.Where("Album = ?", album).Order("date_taken desc").First(&pic).Error
	return pic, err
}

func (p *PictureCollection) FindPublicByAlbum(album string) ([]Picture, error) {
	var pictures []Picture
	if err := p.DB.Where("album = ? AND visibility = ?", album, "PUBLIC").
		Order("date_taken desc").Find(&pictures).Error; err != nil {
		return nil, err
	}
	return filterPublicPictures(pictures), nil
}

// FindByField returns all pictures where a field matches a value (simple string fields)
func (p *PictureCollection) FindByField(field, value string) ([]Picture, error) {
	field = strings.ToLower(strings.TrimSpace(field))
	allowed := map[string]struct{}{
		"album": {}, "album_name": {}, "visibility": {}, "name": {},
	}
	if _, ok := allowed[field]; !ok {
		return nil, fmt.Errorf("unsupported picture field %q", field)
	}
	var dbModels []Picture
	if err := p.DB.Where(field+" = ?", value).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	return dbModels, nil
}

func (p *PictureCollection) GetFilteredPictures(admin bool) ([]Picture, error) {
	var filterPics []Picture
	pictures, err := p.GetAll()
	if err != nil {
		return nil, err
	}
	for _, pic := range pictures {
		if admin {
			filterPics = append(filterPics, pic)
		} else if IsPicturePublishable(pic) {
			filterPics = append(filterPics, pic)
		}
	}
	return filterPics, nil
}

func IsPicturePublic(pic Picture) bool {
	return strings.EqualFold(pic.Visibility, "PUBLIC") &&
		!IsAlbumInBlacklist(pic.AlbumName) &&
		!IsPictureInBlacklist(pic.Name)
}

func IsPicturePublishable(pic Picture) bool {
	return IsPicturePublic(pic) && pathWithinRoot(pic.Path, config.Config.Gallery.Basepath)
}

func filterPublicPictures(pictures []Picture) []Picture {
	filtered := make([]Picture, 0, len(pictures))
	for _, picture := range pictures {
		if IsPicturePublishable(picture) {
			filtered = append(filtered, picture)
		}
	}
	return filtered
}

func (p *PictureCollection) Delete(picture Picture) error {
	p.Lock()
	defer p.Unlock()
	if err := os.Remove(picture.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete picture file: %w", err)
	}
	if err := p.DB.Delete(&picture).Error; err != nil {
		return fmt.Errorf("delete picture record: %w", err)
	}
	return nil
}

func (p *Picture) Load() (image.Image, error) {
	f, err := os.Open(p.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	imageConfig, _, err := image.DecodeConfig(f)
	if err != nil {
		return nil, fmt.Errorf("image %s, header decode failed: %v", p.Path, err)
	}
	const maxImagePixels int64 = 100_000_000
	pixels := int64(imageConfig.Width) * int64(imageConfig.Height)
	if imageConfig.Width <= 0 || imageConfig.Height <= 0 || pixels > maxImagePixels {
		return nil, fmt.Errorf("image %s has unsupported dimensions %dx%d", p.Path, imageConfig.Width, imageConfig.Height)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("rewind image %s: %w", p.Path, err)
	}
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("image %s, decode failed: %v", p.Path, err)
	}
	return img, nil
}

// TagList returns normalized tags stored in the comma-separated database field.
func (p Picture) TagList() []string {
	tags := make([]string, 0)
	seen := make(map[string]struct{})
	for _, value := range strings.Split(p.Tags, ",") {
		tag := strings.TrimSpace(value)
		key := strings.ToLower(tag)
		if tag == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		tags = append(tags, tag)
	}
	return tags
}

func (p *PictureCollection) RemoveInvalidPictures() error {
	var invalidPics []Picture
	pictures, err := p.GetAll()
	if err != nil {
		return fmt.Errorf("load pictures for cleanup: %w", err)
	}
	for _, pic := range pictures {
		info, err := os.Lstat(pic.Path)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("inspect picture %q: %w", pic.Path, err)
		}
		if os.IsNotExist(err) || info.Mode()&os.ModeSymlink != 0 || !pathWithinRoot(pic.Path, config.Config.Gallery.Basepath) {
			invalidPics = append(invalidPics, pic)
		}
	}
	if len(invalidPics) > 0 {
		p.Lock()
		defer p.Unlock()
		if err := p.DB.Delete(&invalidPics).Error; err != nil {
			return fmt.Errorf("failed to remove invalid pictures: %w", err)
		}
	}
	return nil
}
