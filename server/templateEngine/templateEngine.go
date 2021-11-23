package templateengine

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mailgun/raymond/v2"
)

func fileNameFromPath(src string) string {
	fileName := filepath.Base(src)
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

type TemplateEngine struct {
	partialSrc map[string]string
	pagesSrc   map[string]string

	Base  *raymond.Template
	Pages map[string]*raymond.Template
}

func (te *TemplateEngine) loadPartialSrc(filePath string) error {
	name := fileNameFromPath(filePath)
	fmt.Println(name, filePath, "LOADING partials")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	te.partialSrc[name] = string(data)
	return nil
}

func (te *TemplateEngine) loadPageSrc(name string, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	te.pagesSrc[name] = string(data)
	return nil
}

func (te *TemplateEngine) walk(root string) error {
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			fileExtension := filepath.Ext(path)
			if fileExtension != ".hbs" {
				return nil
			}

			pattern := root + string(filepath.Separator) + "*"
			matched, err := filepath.Match(pattern, path)
			subpath := strings.Replace(path, root+"/", "", -1)

			if matched {
				if fileNameFromPath(path) != "default" {
					fmt.Println(path, "LOADING PAGE")
					te.loadPageSrc(fileNameFromPath(path), path)
				}
			} else if strings.HasPrefix(subpath, "partials") {
				te.loadPartialSrc(path)
			}
			return nil
		})
	return err
}

func (te *TemplateEngine) loadPartials(tpl *raymond.Template) {
	for partialName, partialSource := range te.partialSrc {
		tpl.RegisterPartial(partialName, partialSource)
	}
}

func (te *TemplateEngine) Load(templatePath string) error {
	//find all partial sorce
	err := te.walk(templatePath)
	if err != nil {
		return err
	}

	for pageName, pageSource := range te.pagesSrc {
		tpl, err := raymond.Parse(pageSource)
		if err != nil {
			break
		}
		te.loadPartials(tpl)
		te.Pages[pageName] = tpl
	}

	//Check and load base template
	basePath := templatePath + "/default.hbs"
	if fileExists(basePath) {
		data, err := os.ReadFile(basePath)
		if err != nil {
			return nil
		}
		te.Base, err = raymond.Parse(string(data))
		if err != nil {
			return err
		}
		te.loadPartials(te.Base)
		fmt.Println("LOADING BASE")
	}
	return nil
}

func (te *TemplateEngine) ListPages() []string {
	pages := make([]string, len(te.Pages))
	i := 0
	for k := range te.Pages {
		pages[i] = k
		i++
	}
	return pages
}

func (te *TemplateEngine) RenderPage(pageName string, data Page) string {
	page := te.Pages[pageName]
	body := page.MustExec(data)

	if te.Base == nil {
		return body
	}
	data.Body = body
	return te.Base.MustExec(data)
}

func NewTemplateEgine() *TemplateEngine {
	return &TemplateEngine{
		partialSrc: make(map[string]string),
		pagesSrc:   make(map[string]string),
		Pages:      make(map[string]*raymond.Template),
	}
}

// var te = NewTemplateEgine()

// func HomeHandler(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	page := vars["page"]
// 	if te.Pages[page] == nil {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	}
// 	ctx := map[string]string{
// 		"title": "My New Post: " + page,
// 		"body":  "This is my first post!",
// 	}
// 	w.Write([]byte(te.RenderPage(page, ctx)))
// }

// func main() {

// 	te.Load("templates/beta")
// 	fmt.Println(te.ListPages())

// 	r := mux.NewRouter()
// 	//r.HandleFunc("/", HomeHandler)
// 	r.HandleFunc("/{page}/{id}", HomeHandler)
// 	r.HandleFunc("/{page}/{id}", HomeHandler)
// 	//r.HandleFunc("/articles", HomeHandler)

// 	srv := &http.Server{
// 		Handler: r,
// 		Addr:    "0.0.0.0:8000",
// 		// Good practice: enforce timeouts for servers you create!
// 		WriteTimeout: 15 * time.Second,
// 		ReadTimeout:  15 * time.Second,
// 	}

// 	log.Fatal(srv.ListenAndServe())

// 	//fmt.Println()
// }
