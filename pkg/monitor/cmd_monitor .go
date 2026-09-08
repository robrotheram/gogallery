package monitor

import (
	"fmt"
	"sort"
	"sync"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

type CmdMonitor struct {
	Tasks map[string]*CmdTask
	done  chan bool
	mu    sync.RWMutex
}

type CmdTask struct {
	*ProgressStats
	*uiprogress.Bar
}

func NewCmdTask(name string, total int) *CmdTask {
	stat := NewProgressStats(name, total)
	bar := uiprogress.AddBar(total).AppendCompleted()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(name, 20)
	})
	bar.AppendFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("%d/%d", uint(b.Current()), uint(b.Total))
	})
	return &CmdTask{
		ProgressStats: stat,
		Bar:           bar,
	}
}
func (t *CmdTask) Start() {
	t.ProgressStats.Start()
	if t.Bar != nil {
		_ = t.Bar.Set(t.GetProcessed())
	}
}
func (t *CmdTask) Update() {
	t.ProgressStats.Update()
	if t.Bar != nil {
		_ = t.Bar.Set(t.GetProcessed())
	}
}

func (t *CmdTask) Fail(message string) {
	t.ProgressStats.Fail(message)
	if t.Bar != nil {
		_ = t.Bar.Set(t.GetProcessed())
	}
}

func (t *CmdTask) Complete() {
	t.ProgressStats.Complete()
	if t.Bar != nil {
		_ = t.Bar.Set(t.GetProcessed())
	}
}

func NewCMDMonitor() *CmdMonitor {
	return &CmdMonitor{
		Tasks: make(map[string]*CmdTask),
		done:  make(chan bool),
	}
}

func (t *CmdMonitor) NewTask(name string, total int) MonitorStat {
	t.mu.Lock()
	defer t.mu.Unlock()
	if stat, exists := t.Tasks[name]; exists {
		return stat
	}
	stat := NewCmdTask(name, total)
	t.Tasks[name] = stat
	return stat
}

func (t *CmdMonitor) GetTasks() []MonitorStat {
	t.mu.RLock()
	defer t.mu.RUnlock()
	keys := make([]string, 0, len(t.Tasks))
	values := make([]MonitorStat, 0, len(t.Tasks))
	for k := range t.Tasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		values = append(values, t.Tasks[k])
	}
	return values
}

func (t *CmdMonitor) StartUpdater() {
	uiprogress.Start() // Start the renderer in a separate goroutine
}
func (t *CmdMonitor) StopUpdater() {
	t.done <- true
}
