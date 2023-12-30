package pipeline

import (
	"runtime"
	"sync"

	"github.com/robrotheram/gogallery/backend/monitor"
)

type BatchProcessing[T any] struct {
	items     []T
	stat      *monitor.ProgressStats
	work      func(T) error
	chunkSize int
}

func chunkSlice[T any](slice []T, nchunks int) [][]T {
	var chunks [][]T
	chunkSize := (len(slice) / nchunks)
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func (batch *BatchProcessing[T]) Run() {
	batch.stat.Start()
	var wg sync.WaitGroup
	chunks := chunkSlice(batch.items, batch.chunkSize)
	for _, chunk := range chunks {
		wg.Add(1)
		go batch.processing(chunk, &wg)
	}
	wg.Wait()
	batch.stat.End()
}

func (poc *BatchProcessing[T]) processing(batch []T, wg *sync.WaitGroup) {
	for _, pic := range batch {
		poc.work(pic)
		poc.stat.Update()
	}
	wg.Done()
}

func NewBatchProcessing[T any](processing func(T) error, items []T, stat *monitor.ProgressStats) *BatchProcessing[T] {
	stat.Total = len(items)
	//save 1 core for the system
	chunsize := runtime.NumCPU() - 1
	if chunsize < 1 {
		chunsize = 1
	}
	return &BatchProcessing[T]{
		work:      processing,
		chunkSize: chunsize,
		items:     items,
		stat:      stat,
	}
}
