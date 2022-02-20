package monitor

import (
	"errors"
	"sync"

	"github.com/luxcgo/lifesaver/engine"
)

type Manager struct {
	mu     sync.RWMutex
	savers map[engine.ID]Monitor
}

func NewManager() *Manager {
	return &Manager{
		savers: make(map[engine.ID]Monitor),
	}
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, monitor := range m.savers {
		monitor.Stop()
		delete(m.savers, id)
	}
}

func (m *Manager) addMonitor(show engine.Show) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.savers[show.GetID()]; ok {
		return errors.New("exist")
	}
	monitor := NewMonitor(show)
	m.savers[show.GetID()] = monitor
	return monitor.Start()
}
