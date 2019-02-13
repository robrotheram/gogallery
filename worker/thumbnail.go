package worker

import (
	"fmt"
	"github.com/nfnt/resize"
	"github.com/robrotheram/gogallery/config"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

var ThumbnailChan = make(chan string, 100)

func MakeThumbnail(path string) {
	cachePath := fmt.Sprintf("cache/%s.jpg", config.GetMD5Hash(path))

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
		fmt.Println(path)
		return
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(600, 0, img, resize.Bilinear)
	out, err := os.Create(cachePath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	jpeg.Encode(out, m, nil)
}

func worker(id int, jobs <-chan string) {
	fmt.Printf("Strarting Worker: %d \n", id)
	for j := range jobs {
		MakeThumbnail(j)
	}
}

func StartWorkers() {
	for w := 1; w <= config.Config.Server.Workers; w++ {
		go worker(w, ThumbnailChan)
	}
}
