package pipeline

import (
	"runtime"
	"sync"

	"github.com/robrotheram/gogallery/backend/monitor"
)

type BatchProcessing[T any] struct {
	items   []T
	stat    *monitor.ProgressStats
	work    func(T) error
	workers int
}

func (batch *BatchProcessing[T]) Run() {
	batch.stat.Start()
	var wg sync.WaitGroup
	itemCh := make(chan T)

	// Start workers
	for i := 0; i < batch.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range itemCh {
				batch.work(item)
				batch.stat.Update()
			}
		}()
	}

	// Feed items to workers
	for _, item := range batch.items {
		itemCh <- item
	}
	close(itemCh)
	wg.Wait()
	batch.stat.End()
}

func NewBatchProcessing[T any](processing func(T) error, items []T, stat *monitor.ProgressStats) *BatchProcessing[T] {
	stat.Total = len(items)
	workers := runtime.NumCPU() - 1
	if workers < 1 {
		workers = 1
	}
	return &BatchProcessing[T]{
		work:    processing,
		workers: workers,
		items:   items,
		stat:    stat,
	}
}
