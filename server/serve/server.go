package serve

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/auth"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	templateengine "github.com/robrotheram/gogallery/templateEngine"
)

func setupSpaHandler(path string) spaHandler {

	return spaHandler{staticPath: path}
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
	r.HandleFunc("/img/{id}", loadImage)
	r.HandleFunc("/img/{id}/{size}.{ext}", loadImage)
	r = InitApiRoutes(r, config)
	r = auth.InitAuthRoutes(r)
	r = templateengine.InitTemplateRoutes(r, config)
	r.PathPrefix("/dashboard").Handler(http.StripPrefix("/dashboard", setupSpaHandler("ui/dashboard")))
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})
	log.Println("Starting server: " + config.Server.Port)
	log.Fatal(http.ListenAndServe(config.Server.Port, handlers.CORS(headers, methods, origins)(r)))
}
