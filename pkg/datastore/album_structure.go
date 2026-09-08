package datastore

import (
	"gogallery/pkg/config"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type AlbumNode struct {
	Album
	Children AlbumStructure `json:"children"`
}

func (a Album) ToAlbumNode() AlbumNode {
	return AlbumNode{
		Album:    a,
		Children: make(AlbumStructure),
	}
}

type AlbumStructure = map[string]AlbumNode

func SliceToTree(albums []Album, basepath string) AlbumStructure {
	tree := initializeAlbumNodes(albums, basepath)
	processChildAlbums(albums, basepath, tree)
	setParentProfileImages(tree)
	return tree
}

func initializeAlbumNodes(albums []Album, basepath string) map[string]AlbumNode {
	tree := make(map[string]AlbumNode)
	sort.Slice(albums, func(i, j int) bool {
		return albums[i].ParentPath < albums[j].ParentPath
	})
	for _, album := range albums {
		if filepath.Clean(album.ParentPath) == filepath.Clean(basepath) {
			album.ParentPath = ""
			tree[album.Name] = album.ToAlbumNode()
		}
	}
	return tree
}

func processChildAlbums(albums []Album, basepath string, tree map[string]AlbumNode) {
	for _, album := range albums {
		if filepath.Clean(album.ParentPath) != filepath.Clean(basepath) && album.Id != config.GetMD5Hash(basepath) {
			updateAlbumHierarchy(album, basepath, tree)
		}
	}
}

func updateAlbumHierarchy(album Album, basepath string, tree map[string]AlbumNode) {
	relative, err := filepath.Rel(filepath.Clean(basepath), filepath.Clean(album.ParentPath))
	if err != nil || relative == "." || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return
	}
	parents := strings.Split(relative, string(os.PathSeparator))
	children := tree
	for i, parentName := range parents {
		parent, exists := children[parentName]
		if !exists {
			return
		}
		if i == len(parents)-1 {
			album.ParentPath = ""
			parent.Children[album.Name] = album.ToAlbumNode()
			children[parentName] = parent
			return
		}
		children = parent.Children
	}
}

func FindInAlbumStructureByID(album AlbumNode, id string) AlbumNode {
	if album.Id == id {
		return album
	}
	for _, child := range album.Children {
		found := FindInAlbumStructureByID(child, id)
		if found.Id == id {
			return found
		}
	}
	return AlbumNode{}
}

func GetAlbumsFromTree(tree AlbumStructure) []AlbumNode {
	albumList := make([]AlbumNode, 0)
	for _, v := range tree {
		albumList = append(albumList, v)
	}
	sort.Slice(albumList, func(i, j int) bool {
		return strings.ToLower(albumList[i].Name) < strings.ToLower(albumList[j].Name)
	})
	return albumList
}

func GetAlbumFromStructure(tree AlbumStructure, id string) AlbumNode {
	album := AlbumNode{}
	for _, root := range tree {
		album = FindInAlbumStructureByID(root, id)
		if album.Id != "" {
			return album
		}
	}
	return album
}

func SortByTime(albums []Album) []Album {
	sort.Slice(albums, func(i, j int) bool {
		return albums[i].ModTime.After(albums[j].ModTime)
	})
	return albums
}

// Recursively set profile image for parent albums if not set, using a child album's profile image
func setParentProfileImages(tree AlbumStructure) {
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
