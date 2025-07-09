package embeds

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var ThemeFS embed.FS

func CopyTheme(templatePath string) {
	os.MkdirAll(templatePath, os.ModePerm)
	fs.WalkDir(ThemeFS, ".", func(path string, d fs.DirEntry, err error) error {
		newPath := filepath.Join(templatePath, path)
		if d.IsDir() {
			os.MkdirAll(newPath, os.ModePerm)
		} else {
			file, _ := ThemeFS.ReadFile(path)
			os.WriteFile(newPath, file, os.ModePerm)
		}
		return nil
	})

}

func ListThemes() []string {
	items, err := ThemeFS.ReadDir("themes")
	if err != nil {
		return nil
	}
	var pages []string
	for _, item := range items {
		name := strings.TrimSuffix(item.Name(), filepath.Ext(item.Name()))
		pages = append(pages, name)
	}
	return pages
}

func DoesThmeExist(theme string) bool {
	themes := ListThemes()
	for _, t := range themes {
		if t == theme {
			return true
		}
	}
	return false
}

func CopyThemeAssets(theme string, templatePath string) {
	os.MkdirAll(templatePath, os.ModePerm)
	root := "themes/" + theme + "/assets"
	fs.WalkDir(ThemeFS, root, func(path string, d fs.DirEntry, err error) error {
		newPath := filepath.Join(templatePath, strings.Replace(path, root, "", -1))
		if d.IsDir() {
			os.MkdirAll(newPath, os.ModePerm)
		} else {
			file, _ := ThemeFS.ReadFile(path)
			os.WriteFile(newPath, file, os.ModePerm)
		}
		return nil
	})

}
