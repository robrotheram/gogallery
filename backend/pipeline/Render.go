package pipeline

import (
	"os"
	"path/filepath"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/datastore/models"
)

var root = ""
var imgDir string
var photoDir string
var albumsDir string
var albumDir string

type RenderPipeline struct {
	AlbumRender *BatchProcessing[models.Album]
	PageRender  *BatchProcessing[models.Picture]
	ImageRender *BatchProcessing[models.Picture]
	monitor     *TasksMonitor
	db          *datastore.DataStore
	config      *config.GalleryConfiguration
}

func NewRenderPipeline(config *config.GalleryConfiguration, db *datastore.DataStore, monitor *TasksMonitor) *RenderPipeline {
	root = config.Destpath
	imgDir = filepath.Join(root, "img")
	photoDir = filepath.Join(root, "photo")
	albumsDir = filepath.Join(root, "albums")
	albumDir = filepath.Join(root, "album")

	render := RenderPipeline{
		db:          db,
		AlbumRender: NewBatchProcessing(renderAlbumTemplate(db)),
		PageRender:  NewBatchProcessing(renderPhotoTemplate(db)),
		ImageRender: NewBatchProcessing(ImageGenV2),
		monitor:     monitor,
		config:      config,
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
	db := r.db
	r.CreateDir()
	build()
	renderIndex(db, r.config)
	renderAlbums(db)
	r.AlbumRender.Run(db.Albums.GetAll(), r.monitor.NewTask("albums"))
	r.PageRender.Run(db.Pictures.GetAll(), r.monitor.NewTask("photos"))
	r.ImageRender.Run(db.Pictures.GetAll(), r.monitor.NewTask("images"))
}
