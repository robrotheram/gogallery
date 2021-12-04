package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/worker"
)

type UploadCollection struct {
	Album  string   `json:"album"`
	Photos []string `json:"photos"`
}

var uploadHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var uploadCollection UploadCollection
	_ = json.NewDecoder(r.Body).Decode(&uploadCollection)
	var album datastore.Album
	datastore.Cache.DB.One("Id", uploadCollection.Album, &album)
	for _, photo := range uploadCollection.Photos {
		albumPath := fmt.Sprintf("%s/%s", album.ParenetPath, album.Name)
		newPath := fmt.Sprintf("%s/%s", albumPath, photo)
		oldPath := fmt.Sprintf("./temp/%s", config.GetMD5Hash(photo))
		err := datastore.MoveFile(oldPath, newPath)
		if err == nil {
			picName := strings.TrimSuffix(photo, filepath.Ext(photo))
			p := datastore.Picture{
				Id:       config.GetMD5Hash(newPath),
				Name:     picName,
				Path:     newPath,
				Album:    album.Id,
				Exif:     datastore.Exif{},
				RootPath: Config.Gallery.Basepath,
				Meta: datastore.PictureMeta{
					PostedToIG:   false,
					Visibility:   "PUBLIC",
					DateAdded:    time.Now(),
					DateModified: time.Now()}}
			p.CreateExif()
			datastore.Cache.DB.Save(&p)
			worker.SendToThumbnail(p)
		}
	}
})

var uploadFileHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")
	//photoID := mux.Vars(r)["id"]

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern

	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		os.Mkdir("temp", 0755)
	}

	tfile, err := os.OpenFile("./temp/"+config.GetMD5Hash(handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tfile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	// write this byte array to our temporary file
	tfile.Write(fileBytes)
	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")
})
