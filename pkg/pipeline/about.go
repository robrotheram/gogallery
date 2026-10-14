package pipeline

import (
	"io"
	"os"
	"path/filepath"

	templateengine "gogallery/pkg/templateEngine"
)

func (r *RenderPipeline) BuildAbout(w io.Writer) error {
	page := templateengine.NewPage(nil)
	page.PagePath = "/about/"
	page.SEO.SetPath(page.PagePath)
	page.SEO.Title = "About"
	if page.Author.Photographer != "" {
		page.SEO.Title = "About " + page.Author.Photographer
	}
	if info, err := os.Stat(page.Author.ProfilePhoto); err == nil && info.Mode().IsRegular() {
		page.Author.ProfilePhoto = "/about/profile" + filepath.Ext(page.Author.ProfilePhoto)
	}
	return templateengine.Templates.RenderPage(w, templateengine.AboutTemplate, page)
}

func (r *RenderPipeline) renderAbout() error {
	dir := filepath.Join(r.root, "about")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	page := templateengine.NewPage(nil)
	if info, err := os.Stat(page.Author.ProfilePhoto); err == nil && info.Mode().IsRegular() {
		source, err := os.Open(page.Author.ProfilePhoto)
		if err != nil {
			return err
		}
		defer source.Close()
		if err := writeGeneratedFile(filepath.Join(dir, "profile"+filepath.Ext(source.Name())), func(w io.Writer) error {
			_, err := io.Copy(w, source)
			return err
		}); err != nil {
			return err
		}
	}
	return writeGeneratedFile(filepath.Join(dir, "index.html"), r.BuildAbout)
}
