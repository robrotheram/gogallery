package datastore

import (
	"gogallery/pkg/config"
	"path"
	"sort"
	"strings"
)

type AlbumNode struct {
	Album
	Children AlbumStrcure `json:"children"`
}

func (a Album) ToAlbumNode() AlbumNode {
	return AlbumNode{
		Album:    a,
		Children: make(AlbumStrcure),
	}
}

type AlbumStrcure = map[string]AlbumNode

func SliceToTree(albms []Album, basepath string) AlbumStrcure {
	newalbms := initializeAlbumNodes(albms, basepath)
	processChildAlbums(albms, basepath, newalbms)
	setParentProfileImages(newalbms)
	return newalbms
}

func initializeAlbumNodes(albms []Album, basepath string) map[string]AlbumNode {
	newalbms := make(map[string]AlbumNode)
	sort.Slice(albms, func(i, j int) bool {
		return albms[i].ParentPath < albms[j].ParentPath
	})
	for _, ab := range albms {
		if ab.ParentPath == basepath {
			ab.ParentPath = ""
			newalbms[ab.Name] = ab.ToAlbumNode()
		}
	}
	return newalbms
}

func processChildAlbums(albms []Album, basepath string, newalbms map[string]AlbumNode) {
	for _, ab := range albms {
		if (ab.ParentPath != basepath) && (ab.Id != config.GetMD5Hash(basepath)) {
			updateAlbumHierarchy(ab, basepath, newalbms)
		}
	}
}

func updateAlbumHierarchy(ab Album, basepath string, newalbms map[string]AlbumNode) {
	s := strings.Split(strings.Replace(ab.ParentPath, basepath, "", 1), "/")
	copy(s, s[1:])
	s = s[:len(s)-1]
	pth := basepath
	var alb AlbumNode
	for i, p := range s {
		if i == 0 {
			alb = newalbms[p]
		} else {
			alb = alb.Children[p]
		}
		pth = path.Join(pth, p)
		if i == len(s)-1 {
			if alb.Children != nil {
				ab.ParentPath = ""
				alb.Children[ab.Name] = ab.ToAlbumNode()
			}
		}
	}
}

func FindInAlbumStrcureById(ab AlbumNode, Id string) AlbumNode {
	if ab.Id == Id {
		return ab
	}
	for _, v := range ab.Children {
		a := FindInAlbumStrcureById(v, Id)
		if a.Id == Id {
			return a
		}
	}
	return AlbumNode{}
}

func GetAlbmusFromTree(as AlbumStrcure) []AlbumNode {
	albumList := make([]AlbumNode, 0)
	for _, v := range as {
		albumList = append(albumList, v)
	}
	sort.Slice(albumList, func(i, j int) bool {
		return strings.ToLower(albumList[i].Name) < strings.ToLower(albumList[j].Name)
	})
	return albumList
}

func GetAlbumFromStructure(as AlbumStrcure, Id string) AlbumNode {
	album := AlbumNode{}
	for _, v := range as {
		album = FindInAlbumStrcureById(v, Id)
		if album.Id != "" {
			return album
		}
	}
	return album
}

func SortByTime(albs []Album) []Album {
	sort.Slice(albs, func(i, j int) bool {
		return albs[i].ModTime.After(albs[j].ModTime)
	})
	return albs
}

// Recursively set profile image for parent albums if not set, using a child album's profile image
func setParentProfileImages(tree AlbumStrcure) {
	for key, node := range tree {
		if node.ProfileId == "" && len(node.Children) > 0 {
			node.ProfileId = setProfileImageRecursive(&node)
			tree[key] = node // Update the node in the map
		}
	}
}

func setProfileImageRecursive(node *AlbumNode) string {
	// If already has a profile image, return it
	if node.ProfileId != "" {
		return node.ProfileId
	}
	// Try to get from children
	for _, child := range node.Children {
		childProfile := setProfileImageRecursive(&child)
		if childProfile != "" {
			node.ProfileId = childProfile
			return childProfile
		}
	}
	return ""
}
