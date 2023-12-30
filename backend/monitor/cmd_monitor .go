package monitor

import (
	"fmt"
	"sort"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

type CmdMonitor struct {
	Tasks  map[string]*ProgressStats
	Bars   map[string]*uiprogress.Bar
	done   chan bool
	ticker *time.Ticker
}

func NewCMDMonitor() *CmdMonitor {
	return &CmdMonitor{
		Tasks: make(map[string]*ProgressStats),
		Bars:  make(map[string]*uiprogress.Bar),
		done:  make(chan bool),
	}
}

func (t *CmdMonitor) NewTask(name string, total int) *ProgressStats {
	stat := NewProgressStats(name, total)
	t.Tasks[name] = stat
	t.Bars[name] = uiprogress.AddBar(total).AppendCompleted()
	t.Bars[name].PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(name, 20)
	})
	t.Bars[name].AppendFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("%d/%d", uint(b.Current()), uint(b.Total))
	})
	return stat
}

func (t *CmdMonitor) GetTasks() []ProgressStats {
	keys := make([]string, 0, len(t.Tasks))
	values := make([]ProgressStats, 0, len(t.Tasks))
	for k := range t.Tasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		values = append(values, *t.Tasks[k])
	}
	return values
}
func (t *CmdMonitor) StartUpdater() {
	t.ticker = time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-t.done:
				t.UpdateProgress()
				return
			case <-t.ticker.C:
				t.UpdateProgress()
			}
		}
	}()
}
func (t *CmdMonitor) StopUpdater() {
	t.ticker.Stop()
	t.done <- true
}
func (t *CmdMonitor) UpdateProgress() {
	for name, stat := range t.Tasks {
		t.Bars[name].Set(stat.Proceesed)
	}
}
