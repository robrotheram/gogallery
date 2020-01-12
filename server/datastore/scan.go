package datastore

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	Config "github.com/robrotheram/gogallery/config"
)

var validExtension = []string{"jpg", "png", "gif"}
var gConfig *Config.GalleryConfiguration
var IsScanning bool

//albumInBlacklist []string

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

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
		if strings.ToLower(filepath.Ext(path)) == "."+ext {
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
	if strings.EqualFold(album, "temp") {
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
	err := Cache.DB.One("Id", p.Id, &Picture{})
	return err == nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ScanPath(path string, g_config *Config.GalleryConfiguration) (map[string]*Node, error) {
	gConfig = g_config
	rubishPath := fmt.Sprintf("%s/%s", g_config.Basepath, "rubish")
	if _, err := os.Stat(rubishPath); os.IsNotExist(err) {
		os.Mkdir(rubishPath, 0755)
	}
	if !contains(gConfig.AlbumBlacklist, "rubish") {
		gConfig.AlbumBlacklist = append(gConfig.AlbumBlacklist, "rubish")
	}

	log.Println("Scanning Folders at:" + path)
	IsScanning = true

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
					Id:       GetMD5Hash(path),
					Name:     picName,
					Path:     path,
					Album:    albumName,
					Exif:     Exif{},
					RootPath: g_config.Basepath,
					Meta: PictureMeta{
						PostedToIG:   false,
						Visibility:   "PUBLIC",
						DateAdded:    time.Now(),
						DateModified: time.Now()}}
				p.CreateExif()
				if !doesPictureExist(p) {
					Cache.DB.Save(&p)
				}
				Cache.DB.UpdateField(&Album{Id: GetMD5Hash(filepath.Dir(path))}, "ProfileID", p.Id)
				//worker.SendToThumbnail(path)
			}
		}

		if info.IsDir() {
			if !IsAlbumInBlacklist(info.Name()) {
				if filepath.Dir(path) == g_config.Basepath {
					info := fileInfoFromInterface(info)
					Cache.DB.Save(&Album{
						Id:      GetMD5Hash(path),
						Name:    info.Name,
						ModTime: info.ModTime,
						Parent:  filepath.Base(filepath.Dir(path))})
				}
			}
		}
		return nil
	}
	err = filepath.Walk(absRoot, walkFunc)
	log.Println("Scanning Complete")
	IsScanning = false
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
