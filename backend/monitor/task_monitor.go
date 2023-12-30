package monitor

import (
	"sort"
	"sync"
)

type TasksMonitor struct {
	Tasks map[string]*ProgressStats
	*sync.Mutex
}

func NewMonitor() *TasksMonitor {
	return &TasksMonitor{
		Tasks: make(map[string]*ProgressStats),
	}
}

func (t *TasksMonitor) NewTask(name string, total int) *ProgressStats {
	stat := NewProgressStats(name, total)
	t.Tasks[name] = stat
	return stat
}

func (t *TasksMonitor) GetTasks() []ProgressStats {
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
