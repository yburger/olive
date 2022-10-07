package dispatcher

import (
	"errors"

	"github.com/go-olive/olive/engine/enum"
	"github.com/sirupsen/logrus"
)

var SharedManager *Manager

func NewManager(log *logrus.Logger) *Manager {
	return &Manager{
		savers:             make(map[enum.DispatcherTypeID]Dispatcher),
		dispatchFuncSavers: make(map[enum.EventTypeID]Dispatcher),
		log:                log,
	}
}

type Manager struct {
	savers             map[enum.DispatcherTypeID]Dispatcher
	dispatchFuncSavers map[enum.EventTypeID]Dispatcher

	log *logrus.Logger
}

func (m *Manager) Register(ds ...Dispatcher) {
	if m.savers == nil {
		m.savers = map[enum.DispatcherTypeID]Dispatcher{}
	}

	if m.dispatchFuncSavers == nil {
		m.dispatchFuncSavers = map[enum.EventTypeID]Dispatcher{}
	}

	for _, d := range ds {
		_, ok := m.savers[d.DispatcherType()]
		if ok {
			m.log.WithField("type", d).Warn("dispatcher has registered")
		}
		m.savers[d.DispatcherType()] = d

		for _, v := range d.DispatchTypes() {
			m.RegisterFunc(v, d)
		}
	}
}

func (m *Manager) RegisterFunc(typ enum.EventTypeID, d Dispatcher) {
	if m.dispatchFuncSavers == nil {
		m.dispatchFuncSavers = map[enum.EventTypeID]Dispatcher{}
	}
	_, ok := m.dispatchFuncSavers[typ]
	if ok {
		m.log.WithField("type", typ.String()).Warn("dipatch func has registered")
	}
	m.dispatchFuncSavers[typ] = d
}

func (m *Manager) Dispatcher(typ enum.DispatcherTypeID) (Dispatcher, bool) {
	v, ok := m.savers[typ]
	return v, ok
}

func (m *Manager) Dispatch(e *Event) error {
	dispatchFunc, ok := m.dispatchFuncSavers[e.Type]
	if !ok {
		return errors.New("dispatch func does not exist")
	}

	return dispatchFunc.Dispatch(e)
}
