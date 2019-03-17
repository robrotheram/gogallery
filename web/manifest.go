package web

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type icon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes"`
	Type  string `json:"type"`
}

type manifest struct {
	Name        string `json:"name"`
	ShortName   string `json:"short_name"`
	StartURL    string `json:"start_url"`
	Display     string `json:"display"`
	Background  string `json:"background_color"`
	Theme_Color string `json:"theme_color"`
	Lang        string `json:"lang"`
	Orientation string `json:"orientation"`
	Description string `json:"description"`
	Icons       []icon `json:"icons"`
}

func getIconList() (icons []icon) {
	file, err := os.Open("themes/default/static/icons/ios")
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer file.Close()

	list, _ := file.Readdirnames(0) // 0 to read all files and folders
	for _, fname := range list {

		name := strings.TrimSuffix(fname, filepath.Ext(fname))
		size := strings.Split(name, "-")

		icons = append(icons, icon{
			"/static/icons/ios/" + fname,
			size[len(size)-1],
			"image/" + strings.Replace(filepath.Ext(fname), ".", "", 1)})
	}
	return
}

func makeManifest() *manifest {
	return &manifest{
		config.Gallery.Name,
		config.Gallery.Name,
		config.Gallery.Url,
		"standalone",
		"#3E4EB8",
		"#2F3BA2",
		"english",
		"any",
		"Gallery",
		getIconList()}

}
