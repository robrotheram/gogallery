package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/robrotheram/gogallery/api"
	"github.com/robrotheram/gogallery/auth"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/worker"

	"html/template"
	"net/http"
	"strings"

	"github.com/gobuffalo/packr/v2"
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
	tokenString := r.Header.Get("Authorization")
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

	if (picture.Meta.Visibility == "PUBLIC") || (checkAuth(r)) {
		if size == "" {
			cachePath := fmt.Sprintf("cache/%s.jpg", worker.GetMD5Hash(picture.Path))
			if _, err := os.Stat(cachePath); err == nil {
				http.ServeFile(w, r, cachePath)
				return
			}
		} else if size == "tiny" {
			cachePath := fmt.Sprintf("cache/%s.jpg", worker.GetMD5Hash(picture.Path))
			if _, err := os.Stat(cachePath); err == nil {
				http.ServeFile(w, r, cachePath)
				return
			}
		} else if size == "original" {
			http.ServeFile(w, r, picture.Path)
			return
		}
	}
	index, _ := pbox.Find("placeholder.png")
	w.Write(index)
}

var pbox *packr.Box

func setupSpaHandler(box *packr.Box) spaHandler {
	index, _ := box.FindString("index.html")
	t := template.New("T")
	t.Parse(index)
	return spaHandler{staticPath: box, indexTemplate: t}
}

func Serve() {
	r := mux.NewRouter()

	// PUBLIC ROUTES

	fbox := packr.New("Frontend Box", "./ui/frontend")
	bbox := packr.New("Dashboard Box", "./ui/dashboard")
	pbox = packr.New("Placeholder Box", "./ui/placeholders")

	r.HandleFunc("/img/{id}", loadImage)

	r = api.InitApiRoutes(r, Config)
	r = auth.InitAuthRoutes(r)

	r.PathPrefix("/dashboard").Handler(http.StripPrefix("/dashboard", setupSpaHandler(bbox)))
	r.PathPrefix("/").Handler(setupSpaHandler(fbox))

	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	log.Println("Starting server on port" + Config.Server.Port)
	log.Fatal(http.ListenAndServe(Config.Server.Port, handlers.CORS(headers, methods, origins)(r)))
}
