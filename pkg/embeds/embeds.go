package embeds

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var ThemeFS embed.FS

func CopyTheme(templatePath string) error {
	return copyEmbeddedTree(".", templatePath)
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

func DoesThemeExist(theme string) bool {
	themes := ListThemes()
	for _, t := range themes {
		if t == theme {
			return true
		}
	}
	return false
}

func CopyThemeAssets(theme string, templatePath string) error {
	root := "themes/" + theme + "/assets"
	return copyEmbeddedTree(root, templatePath)
}

func copyEmbeddedTree(root, destination string) error {
	// #nosec G301 -- copied static-site assets must be readable by a web server.
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	return fs.WalkDir(ThemeFS, root, func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, name)
		if err != nil || relative == ".." || strings.HasPrefix(relative, "../") {
			return fmt.Errorf("invalid embedded path %q", name)
		}
		newPath := filepath.Join(destination, filepath.FromSlash(relative))
		if d.IsDir() {
			// #nosec G301 -- copied static-site assets must be readable by a web server.
			return os.MkdirAll(newPath, 0o755)
		}
		file, err := ThemeFS.ReadFile(name)
		if err != nil {
			return err
		}
		// #nosec G306 -- copied static-site assets must be readable by a web server.
		if err := os.WriteFile(newPath, file, 0o644); err != nil {
			return fmt.Errorf("write embedded file %q: %w", newPath, err)
		}
		return nil
	})
}
