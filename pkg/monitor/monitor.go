package monitor

type Monitor interface {
	NewTask(name string, total int) MonitorStat
	GetTasks() []MonitorStat
	// Update(name string)
}

type MonitorStat interface {
	Start()
	Update()
	Complete()
	Fail(string)
}
