package pipeline

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	templateengine "gogallery/pkg/templateEngine"
)

type RenderPipeline struct {
	AlbumRender *BatchProcessing[datastore.Album]
	PageRender  *BatchProcessing[datastore.Picture]
	ImageRender *BatchProcessing[datastore.Picture]
	Thumbnails  *BatchProcessing[datastore.Picture]
	config      *config.GalleryConfiguration
	root        string
	imgDir      string
	photoDir    string
	albumsDir   string
	albumDir    string
	*datastore.DataStore
}

func NewRenderPipeline(cfg *config.GalleryConfiguration, db *datastore.DataStore) *RenderPipeline {
	root := strings.TrimSpace(cfg.Destpath)
	return &RenderPipeline{
		DataStore: db,
		config:    cfg,
		root:      root,
		imgDir:    filepath.Join(root, "img"),
		photoDir:  filepath.Join(root, "photo"),
		albumsDir: filepath.Join(root, "albums"),
		albumDir:  filepath.Join(root, "album"),
	}
}

func (r *RenderPipeline) validateDestination() error {
	if r.root == "" {
		return fmt.Errorf("gallery destination path is not configured")
	}
	abs, err := filepath.Abs(r.root)
	if err != nil {
		return fmt.Errorf("resolve gallery destination: %w", err)
	}
	unsafePaths := []string{filepath.VolumeName(abs) + string(os.PathSeparator)}
	if home, err := os.UserHomeDir(); err == nil {
		unsafePaths = append(unsafePaths, home)
	}
	if workingDirectory, err := os.Getwd(); err == nil {
		unsafePaths = append(unsafePaths, workingDirectory)
	}
	for _, unsafePath := range unsafePaths {
		if containsPath(abs, unsafePath) {
			return fmt.Errorf("refusing unsafe gallery destination %q", abs)
		}
	}
	if strings.TrimSpace(r.config.Basepath) != "" {
		galleryRoot, err := filepath.Abs(r.config.Basepath)
		if err != nil {
			return fmt.Errorf("resolve gallery source: %w", err)
		}
		if containsPath(abs, galleryRoot) || containsPath(galleryRoot, abs) {
			return fmt.Errorf("gallery destination must be separate from source path %q", galleryRoot)
		}
	}
	return nil
}

func containsPath(parent, child string) bool {
	relative, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(child))
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

func (r *RenderPipeline) CreateDir() error {
	if err := r.validateDestination(); err != nil {
		return err
	}
	for _, path := range []string{r.root, r.imgDir, r.photoDir, r.albumDir} {
		// #nosec G301 -- generated website directories must be readable by a web server.
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create output directory %q: %w", path, err)
		}
	}
	return nil
}

func (r *RenderPipeline) DeleteSite() error {
	if err := r.validateDestination(); err != nil {
		return err
	}
	if err := os.RemoveAll(r.root); err != nil {
		return fmt.Errorf("delete generated site: %w", err)
	}
	return nil
}

func (r *RenderPipeline) GenerateThumbnails() error {
	images, err := r.Pictures.GetAll()
	if err != nil {
		return fmt.Errorf("load pictures for thumbnails: %w", err)
	}
	thumbnails := NewBatchProcessing(r.generateThumbnails(), images, r.NewTask("Optimizing thumbnails", len(images)))
	return thumbnails.Run()
}

func (r *RenderPipeline) BuildSite() error {
	if err := r.CreateDir(); err != nil {
		return err
	}
	theme := config.NormalizeTheme(r.config.Theme)
	if err := templateengine.Templates.Load(theme); err != nil {
		return err
	}
	if err := r.assets(); err != nil {
		return err
	}
	if err := r.renderIndex(); err != nil {
		return err
	}
	if err := r.renderAlbums(); err != nil {
		return err
	}

	albums, err := r.Albums.GetAll()
	if err != nil {
		return fmt.Errorf("load albums: %w", err)
	}
	publicAlbums := albums[:0]
	for _, album := range albums {
		if !datastore.IsAlbumInBlacklist(album.Name) {
			publicAlbums = append(publicAlbums, album)
		}
	}
	albums = publicAlbums
	images, err := r.Pictures.GetFilteredPictures(false)
	if err != nil {
		return fmt.Errorf("load public pictures: %w", err)
	}

	albumRender := NewBatchProcessing(r.renderAlbumTemplate(), albums, r.NewTask("Building albums", len(albums)))
	pageRender := NewBatchProcessing(r.renderPhotoTemplate(), images, r.NewTask("Building pages", len(images)))
	imageRender := NewBatchProcessing(r.imageGenV2, images, r.NewTask("Building images", len(images)))

	if err := albumRender.Run(); err != nil {
		return err
	}
	if err := pageRender.Run(); err != nil {
		return err
	}
	return imageRender.Run()
}

func writeGeneratedFile(path string, render func(io.Writer) error) (err error) {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".gogallery-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = temporary.Close()
		if err != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err = temporary.Chmod(0o644); err != nil {
		return err
	}
	if err = render(temporary); err != nil {
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	if err = os.Rename(temporaryPath, path); err != nil {
		return err
	}
	return nil
}
