package recorder

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
	case enum.EventType.AddRecorder:
		return m.addRecorder(bout)
	case enum.EventType.RemoveRecorder:
		return m.removeRecorder(bout)
	}
	return nil
}

func (m *Manager) DispatcherType() enum.DispatcherTypeID {
	return enum.DispatcherType.Recorder
}

func (m *Manager) DispatchTypes() []enum.EventTypeID {
	return []enum.EventTypeID{
		enum.EventType.AddRecorder,
		enum.EventType.RemoveRecorder,
	}
}
