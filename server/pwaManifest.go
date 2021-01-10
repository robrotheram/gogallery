package main

import (
	"encoding/json"
	"net/http"
)

type Manifest struct {
	ShortName string `json:"short_name"`
	Name      string `json:"name"`
	Icons     []struct {
		Src   string `json:"src"`
		Sizes string `json:"sizes"`
		Type  string `json:"type"`
	} `json:"icons"`
	StartURL        string `json:"start_url"`
	Display         string `json:"display"`
	ThemeColor      string `json:"theme_color"`
	BackgroundColor string `json:"background_color"`
}

var mainifestTemplate = `{
	"short_name": "TEST App",
	"name": "Create React App Sample",
	"icons": [
	  {
		"src": "favicon.ico",
		"sizes": "64x64 32x32 24x24 16x16",
		"type": "image/x-icon"
	  },
	  {
		"src": "logo192.png",
		"type": "image/png",
		"sizes": "192x192"
	  },
	  {
		"src": "logo512.png",
		"type": "image/png",
		"sizes": "512x512"
	  }
	],
	"start_url": ".",
	"display": "standalone",
	"theme_color": "#000000",
	"background_color": "#ffffff"
  }
`

var getManifest = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	manifest := Manifest{}
	json.Unmarshal([]byte(mainifestTemplate), &manifest)

	manifest.ShortName = Config.Gallery.Name
	manifest.Name = Config.Gallery.Name

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(manifest)

})
