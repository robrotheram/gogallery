package datastore

import (
	"fmt"
	"image"
	"os"
	"sync"
	"time"

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
	Path       string `gorm:"size:1024" json:"path,omitempty"`
	Ext        string `gorm:"size:32" json:"extention,omitempty"`
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
	db.AutoMigrate(&Picture{}) // Use the correct struct for migration
	pictureCollection := &PictureCollection{DB: db}
	return pictureCollection
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
	return p.DB.Model(&Picture{}).Where("id = ?", id).Updates(updates).Error
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

// FindByID returns a picture by its ID as a domain model
func (p *PictureCollection) FindById(id string) (Picture, error) {
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

// FindByField returns all pictures where a field matches a value (simple string fields)
func (p *PictureCollection) FindByField(field, value string) ([]Picture, error) {
	var dbModels []Picture
	if err := p.DB.Where(field+" = ?", value).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	return dbModels, nil
}

func (p *PictureCollection) GetFilteredPictures(admin bool) []Picture {
	var filterPics []Picture
	pictures, _ := p.GetAll()
	for _, pic := range pictures {
		if admin {
			filterPics = append(filterPics, pic)
		} else if !IsAlbumInBlacklist(pic.Album) && pic.Visibility == "PUBLIC" {
			filterPics = append(filterPics, pic)
		}
	}
	return (filterPics)
}

func (p *PictureCollection) Delete(picture Picture) error {
	p.Lock()
	defer p.Unlock()
	os.Remove(picture.Path)
	p.DB.Delete(picture)
	return nil
}

func (p *Picture) Load() (image.Image, error) {
	f, err := os.Open(p.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("image %s, decode failed: %v", p.Path, err)
	}
	return img, nil
}
