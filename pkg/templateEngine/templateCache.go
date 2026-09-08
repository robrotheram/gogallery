package templateengine

import (
	"html/template"
	"sync"
)

type TemplateCache struct {
	cache map[string]*template.Template
	sync.RWMutex
}

func newTemplateCache() *TemplateCache {
	return &TemplateCache{
		cache: make(map[string]*template.Template),
	}
}
func (te *TemplateCache) Get(key string) *template.Template {
	te.RLock()
	defer te.RUnlock()
	if val, ok := te.cache[key]; ok {
		return val
	}
	return nil
}

func (te *TemplateCache) Exists(key string) bool {
	te.RLock()
	defer te.RUnlock()
	if _, ok := te.cache[key]; ok {
		return ok
	}
	return false
}

func (te *TemplateCache) Add(key string, val *template.Template) {
	te.Lock()
	defer te.Unlock()
	te.cache[key] = val
}

func (te *TemplateCache) Reset() {
	te.Lock()
	defer te.Unlock()
	te.cache = make(map[string]*template.Template)
}
