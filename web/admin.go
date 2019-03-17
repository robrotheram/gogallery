package web

import (
	"crypto/subtle"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/worker"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type Stats struct {
	Photos     int
	Albums     int
	ProcessQue int
	ViewCount  int
}

func registerAdmin(r *mux.Router) {

	if config.Admin.Enable {
		r.HandleFunc("/admin", BasicAuth(renderAdminPage))
		r.HandleFunc("/admin/scan", BasicAuth(scanTask))
		r.HandleFunc("/admin/purge", BasicAuth(purgeTask))
		r.HandleFunc("/admin/clear", BasicAuth(clearTask))
	}
	r.Handle("/metrics", promhttp.Handler())
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {

	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)

	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Info(err)
		return false
	}

	return true

}

// BasicAuth wraps a handler requiring HTTP basic auth for it using the given
// username and password and the specified realm, which shouldn't contain quotes.
//
// Most web browser display a dialog with something like:
//
//    The website says: "<realm>"
//
// Which is really stupid so you may want to set the realm to a message rather than
// an actual realm.
func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	realm := "Please enter your username and password for this site"
	return func(w http.ResponseWriter, r *http.Request) {

		user, pass, ok := r.BasicAuth()
		pwdMatch := comparePasswords(config.Admin.Password, []byte(pass))
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(config.Admin.Username)) != 1 || !pwdMatch {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
		handler(w, r)
	}
}

func scanTask(w http.ResponseWriter, r *http.Request) {
	log.Info("Scanning for new images")
	go func() {
		datastore.ScanPath(config.Gallery.Basepath)
	}()
	http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
}
func purgeTask(w http.ResponseWriter, r *http.Request) {
	log.Info("DeletingDB")
	datastore.Cache.RestDB()
	http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
}
func clearTask(w http.ResponseWriter, r *http.Request) {
	log.Info(r.URL)
	datastore.RemoveContents("cache")
	http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
}

func renderAdminPage(w http.ResponseWriter, r *http.Request) {
	s := Stats{1, 1, 1, 1}
	pictures, err := datastore.Cache.Tables("PICTURE").GetAll()
	if err == nil {
		s.Photos = len(pictures.([]datastore.Picture))
	}
	albums, err := datastore.Cache.Tables("ALBUM").GetAll()
	if err == nil {
		s.Albums = len(albums.([]datastore.Album))
	}
	s.ProcessQue = len(worker.ThumbnailChan)
	s.ViewCount = ViewCount / 2

	renderSettingsTemplate(w, "adminPage", s)
}
