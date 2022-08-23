package serve

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/auth"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	templateengine "github.com/robrotheram/gogallery/templateEngine"
)

func setupSpaHandler(path string) spaHandler {
	index := filepath.Join(path, "index.html")
	fmt.Println(index)
	t, err := template.ParseFiles(index)
	if err != nil {
		panic("Unable to find ui at: " + path)
	}
	return spaHandler{staticPath: path, indexTemplate: t}
}

func loadImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=604800") // 7 days
	vars := mux.Vars(r)
	id := vars["id"]
	pic, _ := datastore.GetPictureByID(id)
	http.ServeFile(w, r, pic.Path)
}

func Serve(config *config.Configuration) {
	r := mux.NewRouter()

	// PUBLIC ROUTES

	r.HandleFunc("/img/{id}", loadImage)
	r = InitApiRoutes(r, config)
	r = auth.InitAuthRoutes(r)
	r = templateengine.InitTemplateRoutes(r, config)

	// r.Handle("/manifest.json", getManifest)
	// r.Handle("/service-worker", getServiceWorker)
	r.PathPrefix("/dashboard").Handler(http.StripPrefix("/dashboard", setupSpaHandler("ui/dashboard")))

	//	http.Handle("/", fs)

	//r.PathPrefix("/").Handler(handlers.CompressHandler(setupSpaHandler("ui/frontend")))

	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	log.Println("Starting server: " + config.Server.Port)
	log.Fatal(http.ListenAndServe(config.Server.Port, handlers.CORS(headers, methods, origins)(r)))
}
