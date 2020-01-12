package worker

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/disintegration/gift"
	galleryConfig "github.com/robrotheram/gogallery/config"
)

var thumbnailChan = make(chan string, 1000)
var Config *galleryConfig.GalleryConfiguration

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
		log.Printf("image.Decode failed on image: %s with err: %v \n", filename, err)
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
		log.Fatalf("image encode failed: %v", err)
	}
}

func sendToCommand(path string, size int, prefix string) {
	cachePath := fmt.Sprintf("cache/%s%s.jpg", prefix, GetMD5Hash(path))
	_, err := exec.Command("convert", path, "-trim", "-resize", "800", cachePath).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Println("Command Thumbnail generatet: " + path)
	// output := string(out[:])
	// fmt.Println(output)

}

func generateThumbnail(path string, size int, prefix string) {
	cachePath := fmt.Sprintf("cache/%s%s.jpg", GetMD5Hash(path), prefix)
	src := loadImage(path)
	if src == nil {
		return
	}
	g := gift.New(gift.Resize(size, 0, gift.LanczosResampling))
	dst := image.NewNRGBA(g.Bounds(src.Bounds()))
	g.Draw(dst, src)
	saveImage(cachePath, dst)
	fmt.Println("Internal Thumbnail generate: " + path)
}

func makeCacheFolder() {
	os.MkdirAll("cache", os.ModePerm)
}

func doesThumbExists(path string, prefix string) bool {
	cachePath := fmt.Sprintf("cache/%s%s.jpg", GetMD5Hash(path), prefix)
	if _, err := os.Stat(cachePath); err == nil {
		return true
	}
	return false
}

// CheckCacheFolder Lets check to see if a cache image has already been made before adding it to the channel
func CheckCacheFolder(path string) bool {
	return doesThumbExists(path, "") && doesThumbExists(path, "_tiny")
}

func MakeThumbnail(path string) {
	if(!CheckCacheFolder(path)){
		if Config.Renderer == "imagemagick" {
			sendToCommand(path, 1200, "")
			sendToCommand(path, 450, "_tiny")
		} else {
			generateThumbnail(path, 1200, "")
			generateThumbnail(path, 450, "_tiny")
		}
	}
}

func worker(id int, jobs <-chan string) {
	log.Printf("Strarting Worker: %d \n", id)
	makeCacheFolder()
	for j := range jobs {
		MakeThumbnail(j)
	}
}

func StartWorkers(config *galleryConfig.GalleryConfiguration) {
	Config = config
	for w := 1; w <= runtime.NumCPU(); w++ {
		go worker(w, thumbnailChan)
	}
}
