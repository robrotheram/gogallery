package templateengine

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

func fileNameFromPath(src string) string {
	fileName := filepath.Base(src)
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
