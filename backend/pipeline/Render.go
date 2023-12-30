package pipeline

import (
	"os"
	"path/filepath"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/datastore/models"
	"github.com/robrotheram/gogallery/backend/monitor"
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
	monitor     monitor.Monitor
	db          *datastore.DataStore
	config      *config.GalleryConfiguration
}

func NewRenderPipeline(config *config.GalleryConfiguration, db *datastore.DataStore, monitor monitor.Monitor) *RenderPipeline {
	root = config.Destpath
	imgDir = filepath.Join(root, "img")
	photoDir = filepath.Join(root, "photo")
	albumsDir = filepath.Join(root, "albums")
	albumDir = filepath.Join(root, "album")

	albums := db.Albums.GetAll()
	images := db.Pictures.GetAll()

	render := RenderPipeline{
		db:          db,
		AlbumRender: NewBatchProcessing(renderAlbumTemplate(db), albums, monitor.NewTask("rendering albums", len(albums))),
		PageRender:  NewBatchProcessing(renderPhotoTemplate(db), images, monitor.NewTask("rendering pages", len(images))),
		ImageRender: NewBatchProcessing(ImageGenV2, images, monitor.NewTask("optomizing images", len(images))),
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
	r.AlbumRender.Run()
	r.PageRender.Run()
	r.ImageRender.Run()
}
