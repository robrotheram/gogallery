package datastore

import (
	"fmt"
	"sort"
	"time"

	"github.com/robrotheram/gogallery/config"
)

/**

New Utility Functions for getting photos form the cache.
TODO: Move API to use this common lib

**/
func GetPictureByID(id string) (Picture, error) {
	var pic Picture
	Cache.DB.One("Id", id, &pic)
	if pic.Id == "" {
		return pic, fmt.Errorf("no picture with id %s cound be found", id)
	}
	alb, err := GetAlbumByID(pic.Album)
	if err != nil {
		return pic, err
	}
	pic.AlbumName = alb.Name
	return pic, nil
}

func GetPicturesByAlbumID(id string) []Picture {
	var pic []Picture
	Cache.DB.Find("Album", id, &pic)
	return pic
}

func GetAlbumByID(id string) (Album, error) {
	var album Album
	Cache.DB.One("Id", id, &album)
	if album.Id == "" {
		return album, fmt.Errorf("no album with id %s cound be found", id)
	}
	return album, nil
}

func GetPictures() []Picture {
	var pics []Picture
	Cache.DB.All(&pics)
	return pics
}

func GetAlbums() []Album {
	var albums []Album
	Cache.DB.All(&albums)
	return albums
}

func GetAlbumStructure(config config.GalleryConfiguration) AlbumStrcure {
	newalbms := SliceToTree(GetAlbums(), config.Basepath)
	return newalbms
}

func GetFilteredPictures() []Picture {
	var filterPics []Picture
	for _, pic := range GetPictures() {
		if !IsAlbumInBlacklist(pic.Album) {
			if pic.Meta.Visibility == "PUBLIC" {
				var album Album
				Cache.DB.One("Id", pic.Album, &album)
				cleanpic := Picture{
					Id:         pic.Id,
					Name:       pic.Name,
					Caption:    pic.Caption,
					Album:      pic.Album,
					AlbumName:  album.Name,
					FormatTime: pic.Exif.DateTaken.Format("01-02-2006 15:04:05"),
					Exif:       pic.Exif,
					Meta:       pic.Meta,
				}
				filterPics = append(filterPics, cleanpic)
			}
		}
	}
	sort.Slice(filterPics, func(i, j int) bool {
		return filterPics[i].Exif.DateTaken.Sub(filterPics[j].Exif.DateTaken) > 0
	})
	return filterPics
}

func GetLatestPhotoDate() time.Time {
	pics := GetPictures()
	latests := pics[0].Exif.DateTaken
	for _, p := range pics {
		if p.Exif.DateTaken.After(latests) {
			latests = p.Exif.DateTaken
		}
	}
	return latests
}

func GetPhotosByDate(yourDate time.Time) []Picture {
	pics := GetPictures()
	latests := []Picture{}
	for _, p := range pics {
		if DateEqual(p.Exif.DateTaken, yourDate) {
			latests = append(latests, p)
		}
	}
	return latests
}

func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
