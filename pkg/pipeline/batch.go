package pipeline

import (
	"errors"
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
	// Each worker owns its item's slot. Keep every failure in input order
	// so a completed batch identifies all affected files consistently.
	itemErrors := make([]error, len(batch.items))
	itemCh := make(chan int)
	batch.stat.Start()
	// Start workers
	for i := 0; i < batch.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range itemCh {
				itemErrors[index] = batch.work(batch.items[index])
				batch.stat.Update()
			}
		}()
	}

	// Feed items to workers
	for index := range batch.items {
		itemCh <- index
	}
	close(itemCh)
	wg.Wait()
	failures := make([]error, 0)
	for _, err := range itemErrors {
		if err != nil {
			failures = append(failures, err)
		}
	}
	if len(failures) > 0 {
		itemLabel := "items"
		if len(failures) == 1 {
			itemLabel = "item"
		}
		err := fmt.Errorf("%d %s failed:\n%w", len(failures), itemLabel, errors.Join(failures...))
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
