package monitor

import (
	"sync"
	"time"
)

type ProcessState string

const (
	INACTIVE = ProcessState("inactive")
	RUNNING  = ProcessState("running")
	STOPPED  = ProcessState("stopped")
	COMPLETE = ProcessState("complete")
	ERROR    = ProcessState("error")
)

type ProgressStats struct {
	Name      string        `json:"name"`
	StartTime time.Time     `json:"start"`
	EndTime   time.Time     `json:"end"`
	Duration  time.Duration `json:"duration"`
	Total     int           `json:"total"`
	Processed int           `json:"processed"`
	State     ProcessState  `json:"state"`
	Message   string        `json:"message,omitempty"`
	mu        sync.Mutex    `json:"-"`
}

func NewProgressStats(name string, total int) *ProgressStats {
	return &ProgressStats{
		Name:  name,
		Total: total,
		State: INACTIVE,
	}
}

func (p *ProgressStats) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.startLocked()
}

func (p *ProgressStats) startLocked() {
	p.StartTime = time.Now()
	p.Duration = time.Duration(0)
	p.State = RUNNING
}

func (p *ProgressStats) Fail(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.State = ERROR
	p.EndTime = time.Now()
	p.Duration = p.EndTime.Sub(p.StartTime)
	p.Message = msg
}

func (s *ProgressStats) Update() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.State != RUNNING {
		s.startLocked()
	}
	s.Processed++
}

func (s *ProgressStats) GetProcessed() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Processed
}

func (p *ProgressStats) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.State == ERROR || p.State == COMPLETE {
		return
	}
	p.EndTime = time.Now()
	p.Duration = p.EndTime.Sub(p.StartTime)
	p.State = COMPLETE
	p.Processed = p.Total
}

func (p *ProgressStats) Percent() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Total == 0 {
		return 100
	}
	return float64(p.Processed) / float64(p.Total) * 100
}

// Snapshot returns a consistent copy that is safe to inspect from another goroutine.
func (p *ProgressStats) Snapshot() ProgressStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return ProgressStats{
		Name:      p.Name,
		StartTime: p.StartTime,
		EndTime:   p.EndTime,
		Duration:  p.Duration,
		Total:     p.Total,
		Processed: p.Processed,
		State:     p.State,
		Message:   p.Message,
	}
}
