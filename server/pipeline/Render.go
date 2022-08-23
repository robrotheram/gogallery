package pipeline

import (
	"os"

	"github.com/robrotheram/gogallery/datastore"
)

var root = "/home/robertfletcher/workspace/site"

type RenderPipeline struct {
	AlbumRender *BatchProcessing[datastore.Album]
	PageRender  *BatchProcessing[datastore.Picture]
	ImageRender *BatchProcessing[datastore.Picture]
}

func NewRenderPipeline() *RenderPipeline {
	render := RenderPipeline{}
	render.AlbumRender = NewBatchProcessing(renderAlbumTemplate)
	render.PageRender = NewBatchProcessing(renderPhotoTemplate)
	render.ImageRender = NewBatchProcessing(ImageGenV2)
	return &render
}

func (r *RenderPipeline) CreateDir() {
	os.MkdirAll(root, os.ModePerm)
	os.MkdirAll(imgDir, os.ModePerm)
	os.MkdirAll(photoDir, os.ModePerm)
	os.MkdirAll(albumDir, os.ModePerm)
}

func (r *RenderPipeline) BuildSite() {
	r.CreateDir()
	build()
	renderIndex()
	renderAlbums()
	r.AlbumRender.Run(datastore.GetAlbums())
	r.PageRender.Run(datastore.GetPictures())
	r.ImageRender.Run(datastore.GetPictures())
}
