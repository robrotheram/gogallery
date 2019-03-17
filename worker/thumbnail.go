package worker

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/nfnt/resize"
	galleryConfig "github.com/robrotheram/gogallery/config"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

var ThumbnailChan = make(chan string, 100)

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func generateThumbnail(path string, size uint, prefix string) {
	cachePath := fmt.Sprintf("cache/%s%s.jpg", prefix, GetMD5Hash(path))

	if _, err := os.Stat(cachePath); err == nil {
		return
	}
	os.MkdirAll("cache", os.ModePerm)
	file, err := os.Open(path)
	if err != nil {
		//fmt.Println(path)
		return
	}
	// decode jpeg into image.Image
	extension := filepath.Ext(path)
	var img image.Image
	var img_err error
	switch extension {
	case ".jpg":
		img, img_err = jpeg.Decode(file)
	case ".png":
		img, img_err = png.Decode(file)
	}
	if img_err != nil {
		return
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	log.Printf("Creating Thumbnail for user: %s", path)
	m := resize.Resize(size, 0, img, resize.Bilinear)
	out, err := os.Create(cachePath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	jpeg.Encode(out, m, nil)
}

func MakeThumbnail(path string) {
	generateThumbnail(path, 400, "")
}

func MakeLargeThumbnail(path string) {
	generateThumbnail(path, 1200, "large_")
}

func worker(id int, jobs <-chan string) {
	log.Printf("Strarting Worker: %d \n", id)
	for j := range jobs {
		MakeThumbnail(j)
		MakeLargeThumbnail(j)
	}
}

func StartWorkers(conf galleryConfig.ServerConfiguration) {
	for w := 1; w <= conf.Workers; w++ {
		go worker(w, ThumbnailChan)
	}
}
