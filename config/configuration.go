package config

import (
	"crypto/md5"
	"encoding/hex"
)

type Configuration struct {
	Server   ServerConfiguration
	Database DatabaseConfiguration
	Gallery  GalleryConfiguration
}

var Config Configuration

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
