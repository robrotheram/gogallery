package datastore

import (
	"github.com/robrotheram/gogallery/worker"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var validExtension = []string{"jpg", "png", "gif"}

// FileInfo is a struct created from os.FileInfo interface for serialization.
type FileInfo struct {
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"mod_time"`
	IsDir   bool        `json:"is_dir"`
}

// Helper function to create a local FileInfo struct from os.FileInfo interface.
func fileInfoFromInterface(v os.FileInfo) FileInfo {
	return FileInfo{v.Name(), v.Size(), v.Mode(), v.ModTime(), v.IsDir()}
}

// Node represents a node in a directory tree.
type Node struct {
	FullPath string   `json:"path"`
	Info     FileInfo `json:"info"`
	Children []*Node  `json:"children"`
	Parent   *Node    `json:"-"`
}

func checkEXT(path string) bool {
	chk := false
	for _, ext := range validExtension {
		if filepath.Ext(path) == "."+ext {
			chk = true
		}
	}
	return chk
}

func ScanPath(path string) (map[string]*Node, error) {
	log.Println("Scanning Folders at:" + path)
	absRoot, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	parents := make(map[string]*Node)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if checkEXT(path) && !info.IsDir() {
			p := Picture{
				Id:    path,
				Name:  strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
				Path:  path,
				Album: filepath.Base(filepath.Dir(path)),
				Exif:  Exif{}}
			p.CreateExif()
			Cache.Tables("PICTURE").Save(p)

			a, _ := Cache.Tables("ALBUM").Get(filepath.Dir(path))
			album := a.(Album)
			if album.ProfileIMG == nil {
				album.ProfileIMG = &p
				Cache.Tables("ALBUM").Edit(album)
			}
			worker.ThumbnailChan <- path
		}

		if info.IsDir() {
			info := fileInfoFromInterface(info)
			Cache.Tables("ALBUM").Save(Album{
				Id:      path,
				Name:    info.Name,
				ModTime: info.ModTime,
				Parent:  filepath.Base(filepath.Dir(path))})
		}
		return nil
	}
	err = filepath.Walk(absRoot, walkFunc)
	return parents, err
}

func NewTree(path string) (result *Node, err error) {
	var root = &Node{}
	paths, err := ScanPath(path)
	if err != nil {
		return nil, err
	}
	for path, node := range paths {
		parentPath := filepath.Dir(path)
		parent, exists := paths[parentPath]
		if !exists { // If a parent does not exist, this is the root.
			root = node
		} else {
			node.Parent = parent
			parent.Children = append(parent.Children, node)

		}
	}
	//GalleryCache.AddAlbum()
	return root, nil
}
