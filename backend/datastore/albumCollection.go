package datastore

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/asdine/storm"
	"github.com/robrotheram/gogallery/backend/datastore/models"
)

type AlumnCollectioins struct {
	DB *storm.DB
	sync.Mutex
}

func (a *AlumnCollectioins) Save(pic *models.Album) {
	a.DB.Save(pic)
}

func (a *AlumnCollectioins) Update(alb *models.Album) {
	album := a.FindByID(alb.Id)
	album.Update(*alb)
	a.Save(&album)
}

func (a *AlumnCollectioins) GetAll() []models.Album {
	a.Lock()
	defer a.Unlock()

	var albums []models.Album
	a.DB.All(&albums)
	return albums
}

func (a *AlumnCollectioins) FindByID(id string) models.Album {
	var alb models.Album
	a.DB.One("Id", id, &alb)
	return alb
}

func (a *AlumnCollectioins) FindManyFeild(id string) []models.Album {
	var alb []models.Album
	a.FindByFeild("id", id, alb)
	return alb
}

func (a *AlumnCollectioins) FindByFeild(key string, val any, to any) {
	a.Lock()
	defer a.Unlock()
	a.DB.Find(key, val, &to)
}

func (a *AlumnCollectioins) UpdateField(id string, key string, val any) {
	a.Lock()
	defer a.Unlock()
	a.DB.UpdateField(&models.Album{Id: id}, key, val)
}

func (a *AlumnCollectioins) Delete(albums models.Album) error {
	a.Lock()
	defer a.Unlock()

	err := os.Remove(albums.Path)
	if err != nil {
		return err
	}
	a.DB.DeleteStruct(&albums)
	return nil
}

func (a *AlumnCollectioins) MovePictureToAlbum(picture *models.Picture, newAlbum string) error {
	album := a.FindByID(newAlbum)
	newName := fmt.Sprintf("%s/%s/%s%s", album.ParenetPath, album.Name, picture.Name, filepath.Ext(picture.Path))
	err := os.Rename(picture.Path, newName)
	if err != nil {
		return err
	}
	picture.Path = newName
	picture.Album = newAlbum
	return nil
}
