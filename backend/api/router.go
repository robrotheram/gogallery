package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/pipeline"
)

type GoGalleryAPI struct {
	db      *datastore.DataStore
	config  *config.Configuration
	router  *mux.Router
	monitor *pipeline.TasksMonitor
}

func NewGoGalleryAPI(config *config.Configuration, db *datastore.DataStore) *GoGalleryAPI {
	api := &GoGalleryAPI{config: config, db: db, monitor: pipeline.NewMonitor()}
	api.router = mux.NewRouter()
	api.setupDashboardRoutes()
	return api
}

func (api *GoGalleryAPI) setupDashboardRoutes() {
	fmt.Println("Setting up API")
	api.router.HandleFunc("/img/{id}", api.ImgHandler)
	api.router.HandleFunc("/img/{id}/{size}.{ext}", api.ImgHandler)

	api.router.HandleFunc("/api/admin/photos", api.GetPicturesHandler).Methods("GET")
	api.router.HandleFunc("/api/admin/photo/{id}", api.GetPictureHandler).Methods("GET")
	api.router.HandleFunc("/api/admin/photo/{id}", api.UpdatePictureHandler).Methods("POST")
	api.router.HandleFunc("/api/admin/photo/{id}", api.DeletePictureHandler).Methods("DELETE")

	api.router.HandleFunc("/api/admin/collection/move", api.moveCollectionHandler).Methods("POST")
	api.router.HandleFunc("/api/admin/collection/uploadFile", api.uploadFileHandler).Methods("POST")
	api.router.HandleFunc("/api/admin/collection/upload", api.uploadHandler).Methods("POST")

	api.router.HandleFunc("/api/admin/collections", api.getCollectionsHandler).Methods("GET")
	api.router.HandleFunc("/api/admin/collection", api.createCollectionHandler).Methods("POST")
	api.router.HandleFunc("/api/admin/collection/{id}/photos", api.getCollectionPhotosHandler).Methods("GET")
	api.router.HandleFunc("/api/admin/collection/{id}", api.getCollectionHandler).Methods("GET")
	api.router.HandleFunc("/api/admin/collection/{id}", api.updateCollectionHandler).Methods("POST")

	api.router.HandleFunc("/api/admin/settings/stats", api.statsHandler).Methods("GET")
	api.router.HandleFunc("/api/admin/settings/gallery", api.getGallerySettings).Methods("GET")
	api.router.HandleFunc("/api/admin/settings/gallery", api.setGallerySettings).Methods("POST")
	api.router.HandleFunc("/api/admin/settings/profile", api.getProfileInfo).Methods("GET")
	api.router.HandleFunc("/api/admin/settings/profile", api.setProfileInfo).Methods("POST")

	api.router.HandleFunc("/api/admin/tasks", api.getTasks).Methods("GET")
	api.router.HandleFunc("/api/admin/tasks/purge", api.purgeTaskHandler).Methods("GET")
	api.router.HandleFunc("/api/admin/tasks/rescan", api.rescanTaskHandler).Methods("GET")
	api.router.HandleFunc("/api/admin/tasks/backup", api.backupTaskHandler).Methods("GET")
	api.router.HandleFunc("/api/admin/tasks/upload", api.uploadTaskHandler).Methods("POST")

	api.router.HandleFunc("/api/admin/tasks/build", api.buildTaskHandler).Methods("POST")
	api.router.HandleFunc("/api/admin/tasks/publish", api.deployTaskHandler).Methods("POST")
}

func (api *GoGalleryAPI) ImgHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=604800") // 7 days
	vars := mux.Vars(r)
	id := vars["id"]
	pic := api.db.Pictures.FindByID(id)
	w.Write(pipeline.ProcessImage(pic))
	//http.ServeFile(w, r, pic.Path)
}

func (api *GoGalleryAPI) Serve() {
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	log.Println("Starting server on port: http://" + api.config.Server.Port)
	log.Fatal(http.ListenAndServe(api.config.Server.Port, handlers.CORS(origins, headers, methods)(api.router)))
}
