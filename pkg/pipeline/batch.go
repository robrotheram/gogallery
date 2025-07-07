package pipeline

import (
	"gogallery/pkg/monitor"
	"runtime"
	"sync"
)

type BatchProcessing[T any] struct {
	items   []T
	stat    monitor.MonitorStat
	work    func(T) error
	workers int
}

func (batch *BatchProcessing[T]) Run() {
	var wg sync.WaitGroup
	itemCh := make(chan T)
	defer batch.stat.Complete()
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
}

func NewBatchProcessing[T any](processing func(T) error, items []T, stat monitor.MonitorStat) *BatchProcessing[T] {
	// stat.Total = len(items)
	workers := runtime.NumCPU() - 2
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
