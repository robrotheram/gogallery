package datastore

import (
	"fmt"
	"gogallery/pkg/config"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var validExtension = []string{"jpg", "png", "gif"}

type FileInfo struct {
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"mod_time"`
	IsDir   bool        `json:"is_dir"`
}

// Helper function to create a local FileInfo struct from os.FileInfo interface.
func FileInfoFromInterface(v os.FileInfo) FileInfo {
	return FileInfo{v.Name(), v.Size(), v.Mode(), v.ModTime(), v.IsDir()}
}

// Node represents a node in a directory tree.
type Node struct {
	FullPath string   `json:"path"`
	Info     FileInfo `json:"info"`
	Children []*Node  `json:"children"`
	Parent   *Node    `json:"-"`
}

func CheckEXT(path string) bool {
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

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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
	if strings.EqualFold(album, "rubish") {
		return true
	}
	for _, n := range config.Config.Gallery.AlbumBlacklist {
		if strings.EqualFold(album, n) {
			return true
		}
	}
	return false
}

func IsPictureInBlacklist(name string) bool {
	for _, n := range config.Config.Gallery.PictureBlacklist {
		if strings.EqualFold(name, n) {
			return true
		}
	}
	return false
}

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}
