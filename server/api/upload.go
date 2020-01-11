package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/robrotheram/gogallery/datastore"
)

var uploadHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var uploadCollection datastore.UploadCollection
	_ = json.NewDecoder(r.Body).Decode(&uploadCollection)
	for _, photo := range uploadCollection.Photos {
		albumPath := fmt.Sprintf("%s/%s", Config.Gallery.Basepath, uploadCollection.Album)
		os.Rename(fmt.Sprintf("./temp/%s", datastore.GetMD5Hash(photo)), fmt.Sprintf("%s/%s", albumPath, photo))
		datastore.ScanPath(albumPath, &Config.Gallery)
	}
	fmt.Println(uploadCollection)
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

	tfile, err := os.OpenFile("./temp/"+datastore.GetMD5Hash(handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
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
