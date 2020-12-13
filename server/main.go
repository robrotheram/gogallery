package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/robrotheram/gogallery/api"
	"github.com/robrotheram/gogallery/auth"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/worker"

	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var Config *config.Configuration

func main() {
	Config = config.LoadConfig()
	if _, err := os.Stat(Config.Gallery.Basepath); os.IsNotExist(err) {
		panic("GALLERY DIRECTORY NOT FOUND EXITING!")
	}
	fmt.Printf("%+v\n", Config.Gallery)
	worker.StartWorkers(&Config.Gallery)
	//go setUpWatchers(Config.Gallery.Basepath)

	datastore.Cache = &datastore.DataStore{}
	datastore.Cache.Open(Config.Database.Baseurl)
	defer datastore.Cache.Close()

	// Let provide a tiny cli to allow users to reset the accound.
	cliPtr := flag.Bool("reset-admin", false, "Resets Admin account with a new random password")
	flag.Parse()
	if *cliPtr {
		log.Printf("Resetting Admin Account...")
		datastore.CreateDefaultUser()
		os.Exit(0)
	}

	checkAndCreateAdmin()
	go func() {
		datastore.ScanPath(Config.Gallery.Basepath, &Config.Gallery)
	}()

	Serve()

}

// Check to see if there is a admin account if not create a new one
func checkAndCreateAdmin() {
	var user datastore.User
	datastore.Cache.DB.One("ID", datastore.ADMINID, &user)
	if user.ID != datastore.ADMINID {
		datastore.CreateDefaultUser()
	}

}

func checkAuth(r *http.Request) bool {
	tokenString := r.URL.Query().Get("token")
	if len(tokenString) == 0 {
		return false
	}
	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
	_, err := auth.VerifyToken(tokenString)
	return err == nil
}

func loadImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=604800") // 7 days
	vars := mux.Vars(r)
	name := vars["id"]
	var picture datastore.Picture
	datastore.Cache.DB.One("Id", name, &picture)
	size := r.URL.Query().Get("size")
	if (picture.Meta.Visibility == "PUBLIC") || (picture.Meta.Visibility == "HIDDEN") || (checkAuth(r)) {
		if size == "" {
			cachePath := fmt.Sprintf("cache/%s.jpg", (picture.Id))
			if _, err := os.Stat(cachePath); err == nil {
				http.ServeFile(w, r, cachePath)
				return
			}
		} else if size == "tiny" {
			cachePath := fmt.Sprintf("cache/%s.jpg", (picture.Id))
			if _, err := os.Stat(cachePath); err == nil {
				http.ServeFile(w, r, cachePath)
				return
			}
		} else if size == "original" {
			http.ServeFile(w, r, picture.Path)
			return
		}
	}
	svg := `<svg height="512" viewBox="0 0 512 512" width="512" xmlns="http://www.w3.org/2000/svg"><g><g><path d="m48 105v-81h416v80" fill="#787680"/><path d="m496 124v344a20 20 0 0 1 -20 20h-440a20 20 0 0 1 -20-20v-392a20 20 0 0 1 20-20h108l32 48h300a20 20 0 0 1 20 20z" fill="#acabb1"/><path d="m464 136v308.57l-160-228.57-124.2 177.43 43.8 62.57h-175.6v-320z" fill="#83d8f4"/><circle cx="160" cy="224" fill="#ffda44" r="48"/><path d="m464 444.57v11.43h-240.4l-43.8-62.57 124.2-177.43z" fill="#4e901e"/><path d="m223.6 456h-175.6v-5.14l86-122.86 45.8 65.43z" fill="#91cc04"/></g><g><path d="m476 96h-4v-72a8 8 0 0 0 -8-8h-416a8 8 0 0 0 -8 8v24h-4a28.031 28.031 0 0 0 -28 28v392a28.031 28.031 0 0 0 28 28h440a28.031 28.031 0 0 0 28-28v-344a28.031 28.031 0 0 0 -28-28zm-420-64h400v64h-275.72l-29.62-44.44a8.033 8.033 0 0 0 -6.66-3.56h-88zm432 436a12.01 12.01 0 0 1 -12 12h-440a12.01 12.01 0 0 1 -12-12v-392a12.01 12.01 0 0 1 12-12h103.72l29.62 44.44a8.033 8.033 0 0 0 6.66 3.56h300a12.01 12.01 0 0 1 12 12z"/><path d="m464 128h-416a8 8 0 0 0 -8 8v320a8 8 0 0 0 8 8h416a8 8 0 0 0 8-8v-320a8 8 0 0 0 -8-8zm-404.23 320 74.23-106.05 74.23 106.05zm396.23 0h-228.23l-38.2-54.57 114.43-163.48 152 217.14zm0-28.81-145.45-207.78a8 8 0 0 0 -13.1 0l-117.65 168.07-39.25-56.07a8 8 0 0 0 -13.1 0l-71.45 102.07v-281.48h400z"/><path d="m160 280a56 56 0 1 0 -56-56 56.063 56.063 0 0 0 56 56zm0-96a40 40 0 1 1 -40 40 40.045 40.045 0 0 1 40-40z"/></g></g></svg>`
	w.Write([]byte(svg))
}

func setupSpaHandler(path string) spaHandler {
	index := filepath.Join(path, "index.html")
	fmt.Println(index)
	t, err := template.ParseFiles(index)
	if err != nil {
		panic("Unable to find ui at: " + path)
	}
	return spaHandler{staticPath: path, indexTemplate: t}
}

func Serve() {
	r := mux.NewRouter()

	// PUBLIC ROUTES

	r.HandleFunc("/img/{id}", loadImage)

	r = api.InitApiRoutes(r, Config)
	r = auth.InitAuthRoutes(r)

	r.PathPrefix("/dashboard").Handler(http.StripPrefix("/dashboard", setupSpaHandler("ui/dashboard")))
	r.PathPrefix("/").Handler(setupSpaHandler("ui/frontend"))

	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	log.Println("Starting server on port" + Config.Server.Port)
	log.Fatal(http.ListenAndServe(Config.Server.Port, handlers.CORS(headers, methods, origins)(r)))
}
