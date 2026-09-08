package uiMonitor

import (
	"gogallery/pkg/config"
	"gogallery/pkg/monitor"
	"sort"
	"sync"

	"fyne.io/fyne/v2"
)

type TaskUpdateListener func()

// UIMonitor is a Monitor that notifies listeners on task changes
// for UI updates.
type UIMonitor struct {
	Tasks     map[string]*monitor.ProgressStats
	listeners []TaskUpdateListener
	mu        sync.RWMutex
}

func NewUIMonitor() *UIMonitor {
	return &UIMonitor{
		Tasks: make(map[string]*monitor.ProgressStats),
	}
}

func (m *UIMonitor) NewTask(name string, total int) monitor.MonitorStat {
	m.mu.Lock()
	stat := monitor.NewProgressStats(name, total)
	m.Tasks[name] = stat
	m.mu.Unlock()
	m.notifyListeners()
	return &uiProgressStat{ProgressStats: stat, parent: m}
}

func (m *UIMonitor) GetTasks() []monitor.MonitorStat {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.Tasks))
	for k := range m.Tasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	values := make([]monitor.MonitorStat, 0, len(m.Tasks))
	for _, k := range keys {
		values = append(values, m.Tasks[k])
	}
	return values
}

// RegisterListener adds a callback to be called on task updates
func (m *UIMonitor) RegisterListener(listener TaskUpdateListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

func (m *UIMonitor) notifyListeners() {
	m.mu.RLock()
	listeners := append([]TaskUpdateListener(nil), m.listeners...)
	m.mu.RUnlock()
	for _, l := range listeners {
		go l() // call in goroutine to avoid blocking
	}
}

// uiProgressStat wraps ProgressStats to notify parent monitor on update/complete
// Implements MonitorStat

type uiProgressStat struct {
	*monitor.ProgressStats
	parent *UIMonitor
}

func (u *uiProgressStat) Start() {
	u.ProgressStats.Start()
	u.parent.notifyListeners()
}

func (u *uiProgressStat) Update() {
	u.ProgressStats.Update()
	u.parent.notifyListeners()
}

func (u *uiProgressStat) Complete() {
	u.ProgressStats.Complete()
	snapshot := u.ProgressStats.Snapshot()
	// Send Fyne notification on task complete
	if config.Config.UI.Notification && snapshot.State == monitor.COMPLETE {
		fyne.Do(func() {
			app := fyne.CurrentApp()
			if app == nil {
				return
			}
			app.SendNotification(&fyne.Notification{
				Title:   "Task Complete",
				Content: snapshot.Name + " finished successfully.",
			})
		})
	}
	u.parent.notifyListeners()
}

func (u *uiProgressStat) Fail(message string) {
	u.ProgressStats.Fail(message)
	u.parent.notifyListeners()
}
