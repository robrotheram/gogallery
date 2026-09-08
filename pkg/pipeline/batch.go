package pipeline

import (
	"fmt"
	"runtime"
	"sync"

	"gogallery/pkg/monitor"
)

type BatchProcessing[T any] struct {
	items   []T
	stat    monitor.MonitorStat
	work    func(T) error
	workers int
}

func (batch *BatchProcessing[T]) Run() error {
	var wg sync.WaitGroup
	var errorMu sync.Mutex
	var firstError error
	failures := 0
	itemCh := make(chan T)
	batch.stat.Start()
	// Start workers
	for i := 0; i < batch.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range itemCh {
				if err := batch.work(item); err != nil {
					errorMu.Lock()
					failures++
					if firstError == nil {
						firstError = err
					}
					errorMu.Unlock()
				}
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
	if failures > 0 {
		err := fmt.Errorf("%d item(s) failed; first error: %w", failures, firstError)
		batch.stat.Fail(err.Error())
		return err
	}
	batch.stat.Complete()
	return nil
}

func NewBatchProcessing[T any](processing func(T) error, items []T, stat monitor.MonitorStat) *BatchProcessing[T] {
	// Image decoding is memory-heavy, so scaling one worker per CPU can make a
	// large photo library consume gigabytes without improving UI responsiveness.
	workers := runtime.NumCPU() / 2
	if workers < 1 {
		workers = 1
	}
	if workers > 4 {
		workers = 4
	}
	return &BatchProcessing[T]{
		work:    processing,
		workers: workers,
		items:   items,
		stat:    stat,
	}
}
