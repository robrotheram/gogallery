package config

import ()

type Configuration struct {
	Server   ServerConfiguration
	Database DatabaseConfiguration
	Gallery  GalleryConfiguration
}

var Config Configuration
