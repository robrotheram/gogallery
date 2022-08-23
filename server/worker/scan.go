package worker

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
)

var validExtension = []string{"jpg", "png", "gif"}
var IsScanning bool
var gConfig = config.Config.Gallery

//var ImagePipeline = pipeline.NewImageRenderPipeline()

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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ScanPath(path string) error {
	rubishPath := fmt.Sprintf("%s/%s", gConfig.Basepath, "rubish")
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
		return err
	}
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if checkEXT(path) && !info.IsDir() {
			albumName := filepath.Base(filepath.Dir(path))
			picName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			if !datastore.IsAlbumInBlacklist(albumName) && !datastore.IsPictureInBlacklist(picName) {
				p := datastore.Picture{
					Id:       config.GetMD5Hash(path),
					Name:     picName,
					Path:     path,
					Ext:      filepath.Ext(path),
					Album:    config.GetMD5Hash(filepath.Dir(path)),
					Exif:     datastore.Exif{},
					RootPath: gConfig.Basepath,
					Meta: datastore.PictureMeta{
						PostedToIG:   false,
						Visibility:   "PUBLIC",
						DateAdded:    time.Now(),
						DateModified: time.Now()}}
				p.CreateExif()
				if !datastore.DoesPictureExist(p) {
					p.Save()
				}
				datastore.Cache.DB.UpdateField(&datastore.Album{Id: config.GetMD5Hash(filepath.Dir(path))}, "ProfileID", p.Id)
			}
		}

		if info.IsDir() {
			if !datastore.IsAlbumInBlacklist(info.Name()) {
				if filepath.Base(filepath.Dir(path)) != gConfig.Basepath {
					info := fileInfoFromInterface(info)
					album := datastore.Album{}
					datastore.Cache.DB.One("Id", config.GetMD5Hash(path), &album)
					album.Update(datastore.Album{
						Id:          config.GetMD5Hash(path),
						Name:        info.Name,
						ModTime:     info.ModTime,
						Parent:      filepath.Base(filepath.Dir(path)),
						ParenetPath: (filepath.Dir(path))})
					datastore.Cache.DB.Save(&album)
				}
			}
		}
		return nil
	}
	err = filepath.Walk(absRoot, walkFunc)
	log.Println("Scanning Complete")
	IsScanning = false
	return err
}
