package monitor

import (
	"errors"
	"sync"

	"github.com/go-olive/olive/engine/config"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	mu     sync.RWMutex
	savers map[config.ID]Monitor

	log *logrus.Logger
	cfg *config.Config
}

func NewManager(log *logrus.Logger, cfg *config.Config) *Manager {
	return &Manager{
		savers: make(map[config.ID]Monitor),

		log: log,
		cfg: cfg,
	}
}

func (m *Manager) Stop() {
	for _, monitor := range m.savers {
		monitor.Stop()
		<-monitor.Done()
	}
}

func (m *Manager) addMonitor(bout config.Bout) error {
	bout.RemoveRecorder()

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.savers[bout.GetID()]; ok {
		return errors.New("exist")
	}
	monitor := NewMonitor(m.log, bout, m.cfg)
	m.savers[bout.GetID()] = monitor
	return monitor.Start()
}

func (m *Manager) removeMonitor(bout config.Bout) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	monitor, ok := m.savers[bout.GetID()]
	if !ok {
		return errors.New("monitor not exist")
	}
	monitor.Stop()
	delete(m.savers, bout.GetID())
	return nil
}
