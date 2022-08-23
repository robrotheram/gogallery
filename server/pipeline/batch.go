package pipeline

import (
	"runtime"
	"sync"
)

type BatchProcessing[T any] struct {
	wg        sync.WaitGroup
	work      func(T) error
	chunkSize int
}

func chunkSlice[T any](slice []T, chunkSize int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func (batch *BatchProcessing[T]) Run(items []T) {
	for _, chunk := range chunkSlice(items, batch.chunkSize) {
		go batch.processing(chunk)
	}
	batch.wg.Wait()
}

func (poc *BatchProcessing[T]) processing(batch []T) {
	poc.wg.Add(1)
	defer poc.wg.Done()
	for _, pic := range batch {
		poc.work(pic)
	}
}

func NewBatchProcessing[T any](processing func(T) error) *BatchProcessing[T] {
	proc := BatchProcessing[T]{}
	proc.work = processing
	proc.chunkSize = runtime.NumCPU()
	return &proc
}
