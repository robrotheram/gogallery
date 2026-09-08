package monitor

import (
	"sort"
	"sync"
)

type TasksMonitor struct {
	Tasks map[string]*ProgressStats
	mu    sync.RWMutex
}

func NewMonitor() *TasksMonitor {
	return &TasksMonitor{
		Tasks: make(map[string]*ProgressStats),
	}
}

func (t *TasksMonitor) NewTask(name string, total int) MonitorStat {
	t.mu.Lock()
	defer t.mu.Unlock()
	stat := NewProgressStats(name, total)
	t.Tasks[name] = stat
	return stat
}

func (t *TasksMonitor) GetTasks() []MonitorStat {
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
