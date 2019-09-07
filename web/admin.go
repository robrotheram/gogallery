package web

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/fatih/structs"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	galleryConfig "github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/worker"
	"golang.org/x/crypto/bcrypt"
)

type Stats struct {
	Photos     int
	Albums     int
	ProcessQue int
	ViewCount  int
}

func MakeStats() Stats {
	s := Stats{1, 1, 1, 1}
	var pics []datastore.Picture
	var albms []datastore.Album
	datastore.Cache.DB.All(&pics)
	datastore.Cache.DB.All(&albms)
	s.Photos = len(pics)
	s.Albums = len(albms)
	s.ProcessQue = worker.QueSize()
	s.ViewCount = ViewCount / 2
	return s
}

type AdminModel struct {
	Stats  Stats
	Albums []datastore.Album
	Config map[string]interface{}
	IG     galleryConfig.InstagramConfiguration
}

func registerAdmin(r *mux.Router) {

	if config.Admin.Enable {
		r.HandleFunc("/admin", BasicAuth(renderAdminPage))
		r.HandleFunc("/admin/scan", BasicAuth(scanTask))
		r.HandleFunc("/admin/UpdateImage", BasicAuth(updateImage))
		r.HandleFunc("/admin/IGLogin", BasicAuth(IGLogin))
		r.HandleFunc("/admin/purge", BasicAuth(purgeTask))
		r.HandleFunc("/admin/clear", BasicAuth(clearTask))
		r.HandleFunc("/admin/backup", BasicAuth(backupTask))
		r.HandleFunc("/admin/upload", BasicAuth(uploadTask))
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
		datastore.ScanPath(config.Gallery.Basepath, &config.Gallery)
	}()
	http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
}

func updateImage(w http.ResponseWriter, r *http.Request) {
	log.Info("Updating Image")
	fmt.Println(r.FormValue("id"))
	fmt.Println(r.FormValue("name"))
	fmt.Println(r.FormValue("caption"))
	fmt.Println(r.FormValue("instagram"))
	if len(r.FormValue("id")) > 0 {
		var p datastore.Picture
		err := datastore.Cache.DB.One("Id", r.FormValue("id"), &p)
		if err != nil {
			return
		}
		if len(r.FormValue("caption")) > 0 {
			p.Caption = r.FormValue("caption")
		}
		if len(r.FormValue("name")) > 0 {
			p.Name = r.FormValue("name")
		}
		if !p.PostedToIG && (r.FormValue("instagram") == "on") && config.IG.Enable {
			p.PostedToIG = true
			fmt.Println("Sending IG")
			p.Caption = p.Caption + " 🖼️ " + config.Gallery.Url + "/pic/" + p.Name
			datastore.IG.UploadPhoto(p.Path, p.Caption)
		}
		datastore.Cache.DB.Save(&p)
		fmt.Println("Saved Image")
	}
}

func IGLogin(w http.ResponseWriter, r *http.Request) {
	log.Info("IG Login Image")
	fmt.Println(r.FormValue("username"))
	if len(r.FormValue("username")) > 0 {
		datastore.IG = &datastore.Instagram{GalleryPath: config.Gallery.Basepath}
		if datastore.IG.Connect(r.FormValue("username"), r.FormValue("password")) == nil {
			datastore.IG.SetUpAlbum()
			config.IG.Enable = true
			config.IG.Username = r.FormValue("username")
		} else {
			fmt.Println("IG Login Failed")
		}
	}
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

type backup struct {
	Albums    []datastore.Album   `json:"albums"`
	Pictures  []datastore.Picture `json:"pictures"`
	Instagram []goinsta.Item      `json:"instagram"`
}

func uploadTask(w http.ResponseWriter, r *http.Request) {
	bk := backup{}
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("fileToUpload")
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		fmt.Println(err)
		return
	}
	json.Unmarshal(buf.Bytes(), &bk)
	for _, p := range bk.Pictures {
		datastore.Cache.DB.Save(&p)
	}
	for _, a := range bk.Albums {
		datastore.Cache.DB.Save(&a)
	}
	if config.IG.Enable {
		for _, ig := range bk.Instagram {
			datastore.IG.SavePost(ig)
		}
	}
	http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
}

func backupTask(w http.ResponseWriter, r *http.Request) {
	bk := backup{}
	datastore.Cache.DB.All(&bk.Pictures)
	datastore.Cache.DB.All(&bk.Albums)

	if config.IG.Enable {
		igCache, _ := datastore.IG.GetAllPosts()
		bk.Instagram = igCache
	}
	w.Header().Set("Content-Disposition", "attachment; filename=Gallery-Backup.json")
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	json.NewEncoder(w).Encode(bk)
}

func renderAdminPage(w http.ResponseWriter, r *http.Request) {

	var albms []datastore.Album
	datastore.Cache.DB.All(&albms)

	for i := range albms {
		album := &albms[i]
		var apics []datastore.Picture
		datastore.Cache.DB.Find("Album", album.Name, &apics)
		album.Images = apics
		album.Key = strings.Replace(album.Name, " ", "", -1)
	}

	renderSettingsTemplate(w, "adminPage", AdminModel{Stats: MakeStats(), Albums: albms, IG: config.IG, Config: structs.Map(config)})
}
