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

var validExtension = []string{"jpg", "jpeg", "png", "gif", "webp"}

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
	clean, err := safeRemovalDirectory(dir)
	if err != nil {
		return err
	}
	d, err := os.Open(clean) // #nosec G304 -- clean was resolved and checked against sensitive roots
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(clean, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func safeRemovalDirectory(dir string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", fmt.Errorf("refusing to empty an unspecified directory")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	unsafe := []string{filepath.VolumeName(abs) + string(os.PathSeparator)}
	if home, err := os.UserHomeDir(); err == nil {
		unsafe = append(unsafe, home)
	}
	if workingDirectory, err := os.Getwd(); err == nil {
		unsafe = append(unsafe, workingDirectory)
	}
	for _, path := range unsafe {
		if abs == filepath.Clean(path) {
			return "", fmt.Errorf("refusing to empty unsafe directory %q", abs)
		}
	}
	return abs, nil
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
	if !pathWithinRoot(sourcePath, config.Config.Gallery.Basepath) ||
		!pathWithinRoot(destPath, config.Config.Gallery.Basepath) {
		return fmt.Errorf("source or destination is outside the gallery root")
	}
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("destination file already exists: %s", destPath)
	} else if !os.IsNotExist(err) {
		return err
	}
	inputFile, err := os.Open(sourcePath) // #nosec G304 -- source is constrained to the configured gallery root
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	defer inputFile.Close()
	outputFile, err := os.CreateTemp(filepath.Dir(destPath), ".gogallery-move-*") // #nosec G304 -- destination is constrained above
	if err != nil {
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	temporaryPath := outputFile.Name()
	defer os.Remove(temporaryPath)
	if _, err = io.Copy(outputFile, inputFile); err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	if err := outputFile.Sync(); err != nil {
		return err
	}
	if err := outputFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, destPath); err != nil {
		return err
	}
	// The copy was successful, so now delete the original file
	if err = os.Remove(sourcePath); err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}
