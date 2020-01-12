package main

import (
	"github.com/gobuffalo/packr/v2"
	"net/http"

	"path/filepath"
)

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath *packr.Box
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	//path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, e := h.staticPath.Find(path)
	if e != nil {
		//if os.IsNotExist(err) {
		// file does not exist, serve index.html+
		index, _ := h.staticPath.Find(h.indexPath)
		w.Write(index)
		return
	}
	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(h.staticPath).ServeHTTP(w, r)
}
