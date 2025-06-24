package pipeline

import (
	"os"
	"path/filepath"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/monitor"
)

var root = ""
var imgDir string
var photoDir string
var albumsDir string
var albumDir string

type RenderPipeline struct {
	AlbumRender *BatchProcessing[datastore.Album]
	PageRender  *BatchProcessing[datastore.Picture]
	ImageRender *BatchProcessing[datastore.Picture]
	monitor     monitor.Monitor
	config      *config.GalleryConfiguration
	*datastore.DataStore
}

func NewRenderPipeline(config *config.GalleryConfiguration, db *datastore.DataStore, monitor monitor.Monitor) *RenderPipeline {
	root = config.Destpath
	imgDir = filepath.Join(root, "img")
	photoDir = filepath.Join(root, "photo")
	albumsDir = filepath.Join(root, "albums")
	albumDir = filepath.Join(root, "album")

	render := RenderPipeline{
		DataStore: db,
		monitor:   monitor,
		config:    config,
	}
	return &render
}

func (r *RenderPipeline) CreateDir() {
	os.MkdirAll(root, os.ModePerm)
	os.MkdirAll(imgDir, os.ModePerm)
	os.MkdirAll(photoDir, os.ModePerm)
	os.MkdirAll(albumDir, os.ModePerm)
}

func (r *RenderPipeline) DeleteSite() {
	os.RemoveAll(root)
}

func (r *RenderPipeline) BuildSite() {
	r.CreateDir()
	build()
	r.renderIndex()
	r.renderAlbums()

	albums, _ := r.Albums.GetAll()
	images, _ := r.Pictures.GetAll()

	AlbumRender := NewBatchProcessing(r.renderAlbumTemplate(), albums, r.monitor.NewTask("rendering albums", len(albums)))
	PageRender := NewBatchProcessing(r.renderPhotoTemplate(), images, r.monitor.NewTask("rendering pages", len(images)))
	ImageRender := NewBatchProcessing(ImageGenV2, images, r.monitor.NewTask("optomizing images", len(images)))

	AlbumRender.Run()
	PageRender.Run()
	ImageRender.Run()
}
