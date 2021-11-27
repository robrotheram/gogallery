package datastore

import (
	"path"
	"sort"
	"strings"

	Config "github.com/robrotheram/gogallery/config"
)

type AlbumStrcure = map[string]Album

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
		return FindInAlbumStrcureById(v, id)
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