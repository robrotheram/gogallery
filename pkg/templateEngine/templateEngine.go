package templateengine

import (
	"bytes"
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/embeds"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

const HomeTemplate = "index"
const AlbumTemplate = "albums"
const CollectionTemplate = "collections"
const PhotoTemplate = "photo"
const PaginationTemplate = "pagination"

func (te *TemplateEngine) LoadFromEmbed(theme string) error {
	te.loadMu.Lock()
	defer te.loadMu.Unlock()
	source := "embed:" + theme
	if te.loadedSource == source {
		return nil
	}

	cache := newTemplateCache()
	path := "themes/" + theme
	base, err := template.New("default.tmpl.html").Funcs(template.FuncMap{
		"ImgSizes": func() map[string]ImgSize { return ImageSizes },
	}).ParseFS(embeds.ThemeFS, path+"/default.tmpl.html")
	if err != nil {
		return fmt.Errorf("parse embedded theme %q: %w", theme, err)
	}
	base, err = base.ParseFS(embeds.ThemeFS, path+"/partials/*.html")
	if err != nil {
		return fmt.Errorf("parse embedded theme %q partials: %w", theme, err)
	}

	items, err := embeds.ThemeFS.ReadDir(path + "/pages")
	if err != nil {
		return fmt.Errorf("read embedded theme %q pages: %w", theme, err)
	}
	for _, item := range items {
		if item.IsDir() || !strings.HasSuffix(item.Name(), ".tmpl.html") {
			continue
		}
		name := strings.TrimSuffix(item.Name(), ".tmpl.html")
		pageTemplate, err := base.Clone()
		if err != nil {
			return fmt.Errorf("clone embedded theme %q for page %q: %w", theme, name, err)
		}
		pageTemplate, err = pageTemplate.ParseFS(embeds.ThemeFS, path+"/pages/"+item.Name())
		if err != nil {
			return fmt.Errorf("parse embedded theme %q page %q: %w", theme, name, err)
		}
		cache.Add(name, pageTemplate)
	}
	te.setCache(cache)
	te.loadedSource = source
	return nil
}

func (te *TemplateEngine) AssetServer(theme string, assetPath string) http.Handler {
	assetPrefix := "/assets/"
	embedPath := "themes/" + theme + "/" + assetPath
	fs := http.FS(embeds.ThemeFS)
	return http.StripPrefix(assetPrefix, http.FileServer(http.FS(&embedSubFS{fs, embedPath})))
}

// embedSubFS restricts access to a subdirectory of an http.FS.
type embedSubFS struct {
	fs     http.FileSystem
	subDir string
}

func (e *embedSubFS) Open(name string) (fs.File, error) {
	clean := strings.TrimPrefix(path.Clean("/"+name), "/")
	full := path.Join(e.subDir, clean)
	return e.fs.Open(full)
}

func (te *TemplateEngine) Load(basePath string) error {
	basePath = config.NormalizeTheme(basePath)
	if embeds.DoesThemeExist(basePath) {
		return te.LoadFromEmbed(basePath)
	}
	return te.LoadFromPath(basePath)
}

func (te *TemplateEngine) LoadFromPath(basePath string) error {
	te.loadMu.Lock()
	defer te.loadMu.Unlock()

	cache := newTemplateCache()
	pagePath := "pages"
	baseFile := filepath.Join(basePath, "default.tmpl.html")
	base, err := template.New(filepath.Base(baseFile)).Funcs(template.FuncMap{
		"ImgSizes": func() map[string]ImgSize { return ImageSizes },
	}).ParseFiles(baseFile)
	if err != nil {
		return fmt.Errorf("parse theme %q: %w", basePath, err)
	}
	partials, err := filepath.Glob(filepath.Join(basePath, "partials/*.tmpl.html"))
	if err != nil {
		return fmt.Errorf("find theme %q partials: %w", basePath, err)
	}
	if len(partials) > 0 {
		base, err = base.ParseFiles(partials...)
		if err != nil {
			return fmt.Errorf("parse theme %q partials: %w", basePath, err)
		}
	}
	items, err := os.ReadDir(filepath.Join(basePath, pagePath))
	if err != nil {
		return fmt.Errorf("read theme %q pages: %w", basePath, err)
	}
	for _, item := range items {
		if item.IsDir() || !strings.HasSuffix(item.Name(), ".tmpl.html") {
			continue
		}
		name := strings.TrimSuffix(item.Name(), ".tmpl.html")
		pageTemplate, err := base.Clone()
		if err != nil {
			return fmt.Errorf("clone theme %q for page %q: %w", basePath, name, err)
		}
		pageTemplate, err = pageTemplate.ParseFiles(filepath.Join(basePath, pagePath, item.Name()))
		if err != nil {
			return fmt.Errorf("parse theme %q page %q: %w", basePath, name, err)
		}
		cache.Add(name, pageTemplate)
	}
	te.setCache(cache)
	te.loadedSource = "path:" + basePath
	return nil
}

type TemplateEngine struct {
	Cache   *TemplateCache
	m       *minify.M
	cacheMu sync.RWMutex
	loadMu  sync.Mutex
	// loadedSource avoids reparsing immutable embedded templates for every
	// preview request. Filesystem themes are deliberately reloaded for live edits.
	loadedSource string
}

var Templates = NewTemplateEngine()

func NewTemplateEngine() *TemplateEngine {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	return &TemplateEngine{
		Cache: newTemplateCache(),
		m:     m,
	}
}

func (te *TemplateEngine) setCache(cache *TemplateCache) {
	te.cacheMu.Lock()
	te.Cache = cache
	te.cacheMu.Unlock()
}

func (te *TemplateEngine) RenderPage(w io.Writer, pageName string, data Page) error {
	te.cacheMu.RLock()
	cache := te.Cache
	te.cacheMu.RUnlock()
	if cache == nil {
		return fmt.Errorf("render page %q: templates are not loaded", pageName)
	}
	pageTemplate := cache.Get(pageName)
	if pageTemplate == nil {
		return fmt.Errorf("render page %q: template does not exist", pageName)
	}

	var tpl bytes.Buffer
	if err := pageTemplate.Execute(&tpl, data); err != nil {
		return fmt.Errorf("render page %q: %w", pageName, err)
	}
	b, err := te.m.Bytes("text/html", tpl.Bytes())
	if err != nil {
		return fmt.Errorf("minify page %q: %w", pageName, err)
	}
	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("write page %q: %w", pageName, err)
	}
	return nil
}
