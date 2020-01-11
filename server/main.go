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

	"net/http"

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

	//worker.StartWorkers(&Config.Gallery)
	//go setUpWatchers(Config.Gallery.Basepath)

	datastore.Cache = &datastore.DataStore{}
	datastore.Cache.Open(Config.Database.Baseurl)
	defer datastore.Cache.Close()

	// Let provide a tiny cli to allow users to reset the accound.
	cliPtr := flag.Bool("reset-admin", false, " Help text.")
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

func loadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]
	var picture datastore.Picture
	datastore.Cache.DB.One("Id", name, &picture)

	http.ServeFile(w, r, picture.Path)
}

func Serve() {
	r := mux.NewRouter()

	// PUBLIC ROUTES

	r.HandleFunc("/img/{id}", loadImage)

	r = api.InitApiRoutes(r, Config)
	r = auth.InitAuthRoutes(r)

	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	log.Println("Starting server on port" + Config.Server.Port)
	log.Fatal(http.ListenAndServe(Config.Server.Port, handlers.CORS(headers, methods, origins)(r)))
}
