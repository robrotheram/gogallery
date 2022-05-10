package datastore

import (
	"path"
	"sort"
	"strings"

	"github.com/robrotheram/gogallery/config"
	Config "github.com/robrotheram/gogallery/config"
)

type AlbumStrcure = map[string]Album

func Sort(as AlbumStrcure) AlbumStrcure {
	keys := make([]string, 0, len(as))
	for k := range as {
		keys = append(keys, k)
	}
	data := make(map[string]Album)
	sort.Strings(keys)
	for _, k := range keys {
		data[k] = as[k]
	}
	return data
}

func SliceToTree(albms []Album, basepath string) AlbumStrcure {
	newalbms := make(map[string]Album)
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
		if (ab.ParenetPath != basepath) && (ab.Id != Config.GetMD5Hash(basepath)) {
			s := strings.Split(strings.Replace(ab.ParenetPath, basepath, "", 1), "/")
			copy(s, s[1:])
			s = s[:len(s)-1]
			pth := basepath
			var alb Album
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

func FindInAlbumStrcureById(ab Album, id string) Album {
	if ab.Id == id {
		return ab
	}
	for _, v := range ab.Children {
		a := FindInAlbumStrcureById(v, id)
		if a.Id == id {
			return a
		}
	}
	return Album{}
}

func GetAlbumFromStructure(as AlbumStrcure, id string) Album {
	album := Album{}
	for _, v := range as {
		album = FindInAlbumStrcureById(v, id)
		if album.Id != "" {
			return album
		}
	}
	return album
}

func GetAlbumStructure(config config.GalleryConfiguration) AlbumStrcure {
	albums := []Album{}
	for _, alb := range GetAlbums() {
		if !IsAlbumInBlacklist(alb.Name) {
			albums = append(albums, alb)
		}
	}
	newalbms := SliceToTree(albums, config.Basepath)
	return newalbms
}
