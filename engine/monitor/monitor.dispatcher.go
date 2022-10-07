package monitor

import (
	"github.com/go-olive/olive/engine/config"
	"github.com/go-olive/olive/engine/dispatcher"
	"github.com/go-olive/olive/engine/enum"
	"github.com/sirupsen/logrus"
)

func (m *Manager) Dispatch(event *dispatcher.Event) error {
	bout := event.Object.(config.Bout)

	m.log.WithFields(logrus.Fields{
		"pf": bout.GetPlatform(),
		"id": bout.GetRoomID(),
	}).Info("dispatch ", event.Type)

	switch event.Type {
	case enum.EventType.AddMonitor:
		return m.addMonitor(bout)
	case enum.EventType.RemoveMonitor:
		return m.removeMonitor(bout)
	}
	return nil
}

func (m *Manager) DispatcherType() enum.DispatcherTypeID {
	return enum.DispatcherType.Monitor
}

func (m *Manager) DispatchTypes() []enum.EventTypeID {
	return []enum.EventTypeID{
		enum.EventType.AddMonitor,
		enum.EventType.RemoveMonitor,
	}
}
