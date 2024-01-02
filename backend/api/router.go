package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/monitor"
	"github.com/robrotheram/gogallery/backend/pipeline"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
	"golang.org/x/net/html"
)

type GoGalleryAPI struct {
	db      *datastore.DataStore
	config  *config.Configuration
	router  *mux.Router
	monitor *monitor.TasksMonitor
}

func NewGoGalleryAPI(config *config.Configuration, db *datastore.DataStore) *GoGalleryAPI {
	api := &GoGalleryAPI{config: config, db: db, monitor: monitor.NewMonitor()}
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
	size := r.URL.Query().Get("size")
	vars := mux.Vars(r)
	id := vars["id"]
	if len(size) == 0 {
		size = vars["size"]
	}
	pic := api.db.Pictures.FindByID(id)
	//Is image in cache
	if file, err := api.db.ImageCache.Get(pic.Id, size); err == nil {
		io.Copy(w, file)
		return
	}
	src, err := pic.Load()
	if err != nil {
		return
	}
	cache, _ := api.db.ImageCache.Writer(pic.Id, size)
	writer := io.MultiWriter(w, cache)
	if size, ok := templateengine.ImageSizes[size]; ok {
		pipeline.ProcessImage(src, size, writer)
		return
	}
	pipeline.ProcessImage(src, templateengine.ImageSizes["xsmall"], writer)
}

func (api *GoGalleryAPI) DashboardAPI() {
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	api.router.PathPrefix("/preview-build").Handler(&home{
		base: api.config.Gallery.Destpath,
	})
	spa := SPAHandler{StaticPath: "frontend/dist", IndexPath: "index.html"}
	api.router.PathPrefix("/").Handler(spa)

	log.Println("Starting api server on port: http://" + api.config.Server.GetLocalAddr())
	log.Fatal(http.ListenAndServe(api.config.Server.GetLocalAddr(), handlers.CORS(origins, headers, methods)(api.router)))
}

func (api *GoGalleryAPI) Serve() {
	fs := http.FileServer(http.Dir(api.config.Gallery.Destpath))
	http.Handle("/", fs)
	log.Println("Starting server on port: http://" + api.config.Server.GetAddr())
	log.Fatal(http.ListenAndServe(api.config.Server.GetAddr(), nil))
}

type home struct {
	base string
}

func getAttribute(n *html.Node, attributeName string) string {
	for _, attr := range n.Attr {
		if attr.Key == attributeName {
			return attr.Val
		}
	}
	return ""
}

func updateLinks(n *html.Node) {
	if n.Type == html.ElementNode {
		if n.Data == "a" || n.Data == "img" || (n.Data == "link" && getAttribute(n, "rel") == "stylesheet") {
			updateHrefAttribute(n, "href")
			updateHrefAttribute(n, "src")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		updateLinks(c)
	}
}

func updateHrefAttribute(n *html.Node, attributeName string) {
	if hrefAttr := findAttribute(n, attributeName); hrefAttr != nil {
		href := hrefAttr.Val
		if href != "" && !strings.HasPrefix(href, "http") {
			// Update the href attribute value by adding the "/dev/" prefix
			hrefAttr.Val = "/preview-build" + href
		}
	}
}

func findAttribute(n *html.Node, attributeName string) *html.Attribute {
	for i := range n.Attr {
		if n.Attr[i].Key == attributeName {
			return &n.Attr[i]
		}
	}
	return nil
}

func GetFileContentType(ouput *os.File) (string, error) {
	buf := make([]byte, 512)
	_, err := ouput.Read(buf)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buf)
	if contentType == "text/plain; charset=utf-8" && filepath.Ext(ouput.Name()) == ".svg" {
		contentType = "image/svg+xml; charset=utf-8"
	} else if contentType == "text/plain; charset=utf-8" && filepath.Ext(ouput.Name()) == ".css" {
		contentType = "text/css; charset=utf-8"
	}
	ouput.Seek(0, 0)
	return contentType, nil
}

func IsHtml(constentType string) bool {
	return strings.Contains(constentType, "text/html")
}

func (h *home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := strings.Replace(r.URL.Path, "/preview-build", "", -1)
	data = path.Join(h.base, data)
	fileInfo, err := os.Stat(data)
	if err != nil {
		return
	}
	if fileInfo.IsDir() {
		data = path.Join(data, "index.html")
	}
	file, err := os.Open(data)
	if err != nil {
		return
	}
	contentType, _ := GetFileContentType(file)
	if IsHtml(contentType) {
		doc, err := html.Parse(file)
		if err != nil {
			fmt.Println(err)
		}
		updateLinks(doc)
		html.Render(w, doc)
	} else {
		w.Header().Set("Content-Type", contentType)
		io.Copy(w, file)
	}
}
