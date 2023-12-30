package monitor

import "time"

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

func (p *ProgressStats) Update() {
	p.Proceesed = p.Proceesed + 1
	p.Duration = time.Since(p.StartTime)
}

func (p *ProgressStats) End() {
	p.EndTime = time.Now()
	p.Duration = p.EndTime.Sub(p.StartTime)
	p.State = COMPLETE
}

func (p *ProgressStats) Percent() float64 {
	if p.Total == 0 {
		return 100
	}
	return ((float64(p.Proceesed) / float64(p.Total)) * float64(100))
}
