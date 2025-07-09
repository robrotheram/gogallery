package pipeline

import (
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	templateengine "gogallery/pkg/templateEngine"
	"os"
	"path/filepath"
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
	Thumbnails  *BatchProcessing[datastore.Picture]
	config      *config.GalleryConfiguration
	*datastore.DataStore
}

func NewRenderPipeline(config *config.GalleryConfiguration, db *datastore.DataStore) *RenderPipeline {
	root = config.Destpath
	imgDir = filepath.Join(root, "img")
	photoDir = filepath.Join(root, "photo")
	albumsDir = filepath.Join(root, "albums")
	albumDir = filepath.Join(root, "album")

	render := RenderPipeline{
		DataStore: db,
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

func (r *RenderPipeline) GenTumbnails() {
	images, _ := r.Pictures.GetAll()
	thumbnails := NewBatchProcessing(r.generateThumbnails(), images, r.NewTask("Optomizing thumbnails", len(images)))
	thumbnails.Run()
}

func (r *RenderPipeline) BuildSite() {
	r.CreateDir()
	err := templateengine.Templates.Load(config.Config.Gallery.Theme)
	if err != nil {
		fmt.Println(err)
	}
	Assets()
	r.renderIndex()
	r.renderAlbums()

	albums, _ := r.Albums.GetAll()
	images, _ := r.Pictures.GetAll()

	AlbumRender := NewBatchProcessing(r.renderAlbumTemplate(), albums, r.NewTask("Building albums", len(albums)))
	PageRender := NewBatchProcessing(r.renderPhotoTemplate(), images, r.NewTask("Building pages", len(images)))
	ImageRender := NewBatchProcessing(ImageGenV2, images, r.NewTask("Building images", len(images)))

	AlbumRender.Run()
	PageRender.Run()
	ImageRender.Run()
}
