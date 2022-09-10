package pipeline

import (
	"sort"
	"sync"
	"time"
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

func (t *TasksMonitor) NewTask(name string) *ProgressStats {
	stat := NewProgressStats(name)
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

type ProgressStats struct {
	Name      string        `json:"name"`
	Done      bool          `json:"done"`
	StartTime time.Time     `json:"start"`
	EndTime   time.Time     `json:"end"`
	Duration  time.Duration `json:"duration"`
}

func NewProgressStats(name string) *ProgressStats {
	return &ProgressStats{Name: name}
}

func (p *ProgressStats) Start() {
	p.StartTime = time.Now()
	p.Duration = time.Duration(0)
	p.Done = false
}

func (p *ProgressStats) Update() {
	p.Duration = time.Since(p.StartTime)
}

func (p *ProgressStats) End() {
	p.EndTime = time.Now()
	p.Duration = p.EndTime.Sub(p.StartTime)
	p.Done = true
}
