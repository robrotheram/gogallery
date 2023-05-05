package datastore

import (
	"os"
	"sync"

	"github.com/asdine/storm"
	"github.com/robrotheram/gogallery/backend/datastore/models"
)

type PictureCollection struct {
	DB *storm.DB
	sync.Mutex
}

func (p *PictureCollection) Save(pic *models.Picture) {
	p.DB.Save(pic)
}

func (p *PictureCollection) GetAll() []models.Picture {
	p.Lock()
	defer p.Unlock()

	var pics []models.Picture
	err := p.DB.All(&pics)
	if err != nil {
		return []models.Picture{}
	}
	return pics
}

func (p *PictureCollection) FindByID(id string) models.Picture {
	p.Lock()
	defer p.Unlock()
	var alb models.Picture
	p.DB.One("Id", id, &alb)
	return alb
}

func (p *PictureCollection) FindManyFeild(key string, val string) []models.Picture {
	var alb []models.Picture
	p.DB.Find(key, val, &alb)
	return alb
}

func (p *PictureCollection) Delete(picture models.Picture) error {
	p.Lock()
	defer p.Unlock()

	err := os.Remove(picture.Path)
	if err != nil {
		return err
	}
	p.DB.DeleteStruct(&picture)
	return nil
}

func (p *PictureCollection) Exist(id string) bool {
	pic := p.FindByID(id)
	return pic.Id != ""
}

func (p *PictureCollection) GetByAlbumID(id string) []models.Picture {
	return models.SortByTime(p.FindManyFeild("Album", id))
}

func (p *PictureCollection) GetFilteredPictures(admin bool) []models.Picture {
	var filterPics []models.Picture
	for _, pic := range p.GetAll() {
		if admin {
			filterPics = append(filterPics, pic)
		} else if !IsAlbumInBlacklist(pic.Album) && pic.Meta.Visibility == "PUBLIC" {
			cleanpic := models.Picture{
				Id:         pic.Id,
				Name:       pic.Name,
				Caption:    pic.Caption,
				Album:      pic.Album,
				AlbumName:  pic.AlbumName,
				FormatTime: pic.Exif.DateTaken.Format("01-02-2006 15:04:05"),
				Exif:       pic.Exif,
				Meta:       pic.Meta,
				Ext:        pic.Ext,
			}
			filterPics = append(filterPics, cleanpic)
		}
	}
	return models.SortByTime(filterPics)
}

func (p *PictureCollection) GetLatestAlbum() string {
	pics := p.GetFilteredPictures(false)
	latests := pics[0].Exif.DateTaken
	album := pics[0].Album
	for _, p := range pics {
		if p.Exif.DateTaken.After(latests) {
			latests = p.Exif.DateTaken
			album = p.Album
		}
	}
	return album
}
