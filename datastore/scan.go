package datastore

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	Config "github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/worker"
)

var validExtension = []string{"jpg", "png", "gif"}
var gConfig *Config.GalleryConfiguration

//albumInBlacklist []string

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

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func IsAlbumInBlacklist(album string) bool {
	if strings.EqualFold(album, "instagram") {
		return true
	}
	if strings.EqualFold(album, "images") {
		return true
	}
	for _, n := range gConfig.AlbumBlacklist {
		if strings.EqualFold(album, n) {
			return true
		}
	}
	return false
}

func IsPictureInBlacklist(pic string) bool {
	for _, n := range gConfig.PictureBlacklist {
		if strings.EqualFold(pic, n) {
			return true
		}
	}
	return false
}
func doesPictureExist(p Picture) bool {
	pics, err := Cache.Tables("PICTURE").Query("Id", p.Id, 0)
	if err != nil {
		return false
	}
	return len(pics.([]Picture)) > 0
}

func ScanPath(path string, g_config *Config.GalleryConfiguration) (map[string]*Node, error) {
	log.Println("Scanning Folders at:" + path)
	gConfig = g_config
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
			albumName := filepath.Base(filepath.Dir(path))
			picName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			if !IsAlbumInBlacklist(albumName) && !IsPictureInBlacklist(picName) {
				p := Picture{
					Id:    path,
					Name:  picName,
					Path:  path,
					Album: albumName,
					Exif:  Exif{}}
				p.CreateExif()
				if !doesPictureExist(p) {
					Cache.Tables("PICTURE").Save(p)
				}

				a, _ := Cache.Tables("ALBUM").Get(filepath.Dir(path))
				album := a.(Album)
				if album.ProfileIMG == nil {
					album.ProfileIMG = &p
					Cache.Tables("ALBUM").Edit(album)
				}
				worker.ThumbnailChan <- path
			}
		}

		if info.IsDir() {
			if !IsAlbumInBlacklist(info.Name()) {
				info := fileInfoFromInterface(info)
				Cache.Tables("ALBUM").Save(Album{
					Id:      path,
					Name:    info.Name,
					ModTime: info.ModTime,
					Parent:  filepath.Base(filepath.Dir(path))})
			}
		}
		return nil
	}
	err = filepath.Walk(absRoot, walkFunc)
	log.Println("Scanning Complete")
	return parents, err
}

func NewTree(path string) (result *Node, err error) {
	var root = &Node{}
	paths, err := ScanPath(path, nil)
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
