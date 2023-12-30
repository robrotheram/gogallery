package monitor

type Monitor interface {
	NewTask(name string, total int) *ProgressStats
	GetTasks() []ProgressStats
}
