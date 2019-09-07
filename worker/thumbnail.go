package worker

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"runtime"

	"github.com/disintegration/gift"
)

var thumbnailChan = make(chan string)

func QueSize() int {
	return len(thumbnailChan)
}
func SendToThumbnail(image string) {
	if !CheckCacheFolder(image) {
		thumbnailChan <- image
	}
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

var img image.Image

func loadImage(filename string) image.Image {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("os.Open failed: %v", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		log.Println("image.Decode failed on image: %s with err: %v", filename, err)
		return nil
	}
	return img
}
func saveImage(filename string, img image.Image) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer f.Close()
	err = jpeg.Encode(f, img, nil)
	if err != nil {
		log.Fatalf("png.Encode failed: %v", err)
	}
}

func makeCacheFolder() {
	os.MkdirAll("cache", os.ModePerm)
}

func doesThumbExists(path string, prefix string) bool {
	cachePath := fmt.Sprintf("cache/%s%s.jpg", prefix, GetMD5Hash(path))
	if _, err := os.Stat(cachePath); err == nil {
		return true
	}
	return false
}

//Lets check to see if a cache image has already been made before adding it to the channel
func CheckCacheFolder(path string) bool {
	return doesThumbExists(path, "") && doesThumbExists(path, "large_")
}

func generateThumbnail(path string, size int, prefix string) {
	cachePath := fmt.Sprintf("cache/%s%s.jpg", prefix, GetMD5Hash(path))
	src := loadImage(path)
	if src == nil {
		return
	}
	g := gift.New(gift.Resize(size, 0, gift.LanczosResampling))
	dst := image.NewNRGBA(g.Bounds(src.Bounds()))
	g.Draw(dst, src)
	saveImage(cachePath, dst)

	src = nil
	dst = nil
}

func MakeThumbnail(path string) {
	generateThumbnail(path, 400, "")
}

func MakeLargeThumbnail(path string) {
	generateThumbnail(path, 1200, "large_")
}

func worker(id int, jobs <-chan string) {
	log.Printf("Strarting Worker: %d \n", id)
	makeCacheFolder()
	for j := range jobs {
		MakeThumbnail(j)
		MakeLargeThumbnail(j)
		runtime.GC()
	}
}

func StartWorkers() {
	for w := 1; w <= runtime.NumCPU(); w++ {
		go worker(w, thumbnailChan)
	}
}
