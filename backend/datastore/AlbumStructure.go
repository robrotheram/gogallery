package datastore

import (
	"path"
	"sort"
	"strings"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore/models"
)

type AlbumStrcure = map[string]models.Album

func Sort(as AlbumStrcure) AlbumStrcure {
	keys := make([]string, 0, len(as))
	for k := range as {
		keys = append(keys, k)
	}
	data := make(map[string]models.Album)
	sort.Strings(keys)
	for _, k := range keys {
		data[k] = as[k]
	}
	return data
}

func SliceToTree(albms []models.Album, basepath string) AlbumStrcure {
	newalbms := make(map[string]models.Album)
	sort.Slice(albms, func(i, j int) bool {
		return albms[i].ParenetPath < albms[j].ParenetPath
	})
	for _, ab := range albms {
		if ab.ParenetPath == basepath {
			ab.ParenetPath = ""
			newalbms[ab.Name] = ab
		}
	}
	for _, ab := range albms {
		if (ab.ParenetPath != basepath) && (ab.Id != config.GetMD5Hash(basepath)) {
			s := strings.Split(strings.Replace(ab.ParenetPath, basepath, "", 1), "/")
			copy(s, s[1:])
			s = s[:len(s)-1]
			pth := basepath
			var alb models.Album
			for i, p := range s {
				if i == 0 {
					alb = newalbms[p]
				} else {
					alb = alb.Children[p]
				}
				pth = path.Join(pth, p)
				if i == len(s)-1 {
					if alb.Children != nil {
						ab.ParenetPath = ""
						alb.Children[ab.Name] = ab
					}
				}
			}
		}
	}
	return newalbms
}

func FindInAlbumStrcureById(ab models.Album, id string) models.Album {
	if ab.Id == id {
		return ab
	}
	for _, v := range ab.Children {
		a := FindInAlbumStrcureById(v, id)
		if a.Id == id {
			return a
		}
	}
	return models.Album{}
}

func GetAlbumFromStructure(as AlbumStrcure, id string) models.Album {
	album := models.Album{}
	for _, v := range as {
		album = FindInAlbumStrcureById(v, id)
		if album.Id != "" {
			return album
		}
	}
	return album
}

func (a *AlumnCollectioins) GetAlbumStructure(config config.GalleryConfiguration) AlbumStrcure {
	albums := []models.Album{}
	for _, alb := range a.GetAll() {
		if !IsAlbumInBlacklist(alb.Name) {
			albums = append(albums, alb)
		}
	}
	newalbms := SliceToTree(albums, config.Basepath)
	return newalbms
}
