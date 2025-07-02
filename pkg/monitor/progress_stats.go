package monitor

import (
	"sync"
	"time"
)

type ProssesState string

const (
	INACTIVE = ProssesState("inactive")
	RUNNING  = ProssesState("running")
	STOPPED  = ProssesState("stopped")
	COMPLETE = ProssesState("complete")
	ERROR    = ProssesState("error")
)

type ProgressStats struct {
	Name      string        `json:"name"`
	StartTime time.Time     `json:"start"`
	EndTime   time.Time     `json:"end"`
	Duration  time.Duration `json:"duration"`
	Total     int           `json:"total"`
	Proceesed int           `json:"processed"`
	State     ProssesState  `json:"state"`
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
	p.StartTime = time.Now()
	p.Duration = time.Duration(0)
	p.State = RUNNING
}

func (p *ProgressStats) Fail(msg string) {
	p.State = ERROR
	p.EndTime = time.Now()
	p.Duration = p.EndTime.Sub(p.StartTime)
	p.Message = msg
}

func (s *ProgressStats) Update() {
	s.mu.Lock()
	if s.State != RUNNING {
		s.Start()
	}
	s.Proceesed++
	s.mu.Unlock()
}

func (s *ProgressStats) GetProcessed() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Proceesed
}

func (p *ProgressStats) Complete() {
	p.EndTime = time.Now()
	p.Duration = p.EndTime.Sub(p.StartTime)
	p.State = COMPLETE
	p.Proceesed = p.Total // Ensure processed is set to total on completion
}

func (p *ProgressStats) Percent() float64 {
	if p.Total == 0 {
		return 100
	}
	return ((float64(p.Proceesed) / float64(p.Total)) * float64(100))
}
