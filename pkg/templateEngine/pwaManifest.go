package templateengine

import (
	"encoding/json"
	"io"
	"testingFyne/pkg/config"
)

type Manifest struct {
	ShortName string `json:"short_name"`
	Name      string `json:"name"`
	Icons     []struct {
		Purpose string `json:"purpose"`
		Src     string `json:"src"`
		Sizes   string `json:"sizes"`
		Type    string `json:"type"`
	} `json:"icons"`
	StartURL        string `json:"start_url"`
	Display         string `json:"display"`
	ThemeColor      string `json:"theme_color"`
	BackgroundColor string `json:"background_color"`
}

const mainifestTemplate = `{
	"short_name": "TEST App",
	"name": "Create React App Sample",
	"icons": [
	  {
		"src": "/assets/logos/favicon.ico",
		"sizes": "64x64 32x32 24x24 16x16",
		"type": "image/x-icon"
	  },
	  {
		"src": "/assets/logos/logo192.png",
		"type": "image/png",
		"sizes": "192x192",
		"purpose": "any maskable"
	  },
	  {
		"src": "/assets/logos/logo512.png",
		"type": "image/png",
		"sizes": "512x512"
	  }
	],
	"start_url": "/",
	"display": "standalone",
	"theme_color": "#000000",
	"background_color": "#ffffff"
  }
`
const serviceWorker = `
	const CACHE_NAME = 'gogaller-cache';
	const toCache = [
	'/',
	'/manifest.json',
	'/service-worker.js',
	'/assets/logos/logo192.png',
	'/assets/logos/logo512.png',
	];

	self.addEventListener('install', function(event) {
	event.waitUntil(
		caches.open(CACHE_NAME)
		.then(function(cache) {
			return cache.addAll(toCache)
		})
		.then(self.skipWaiting())
	)
	})

	self.addEventListener('fetch', function(event) {
	event.respondWith(
		fetch(event.request)
		.catch(() => {
			return caches.open(CACHE_NAME)
			.then((cache) => {
				return cache.match(event.request)
			})
		})
	)
	})

	self.addEventListener('activate', function(event) {
	event.waitUntil(
		caches.keys()
		.then((keyList) => {
			return Promise.all(keyList.map((key) => {
			if (key !== CACHE_NAME) {
				console.log('[ServiceWorker] Removing old cache', key)
				return caches.delete(key)
			}
			}))
		})
		.then(() => self.clients.claim())
	)
	})
`

func ManifestWriter(w io.Writer, config *config.GalleryConfiguration) {
	manifest := Manifest{}
	json.Unmarshal([]byte(mainifestTemplate), &manifest)
	manifest.ShortName = config.Name
	manifest.Name = config.Name
	manifest.StartURL = config.Url
	json.NewEncoder(w).Encode(manifest)
}

func ServiceWorkerWriter(w io.Writer) {
	w.Write([]byte(serviceWorker))
}
