// Package monitor monitors streams.
package monitor

import (
	"sync/atomic"
	"time"

	"github.com/go-olive/olive/engine/config"
	"github.com/go-olive/olive/engine/dispatcher"
	"github.com/go-olive/olive/engine/enum"
	"github.com/lthibault/jitterbug/v2"
	"github.com/sirupsen/logrus"
)

type Monitor interface {
	Start() error
	Stop()
	Done() <-chan struct{}
}

func NewMonitor(log *logrus.Logger, bout config.Bout, cfg *config.Config) Monitor {
	return &monitor{
		status: enum.Status.Starting,
		bout:   bout,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),

		log: log,
		cfg: cfg,
	}
}

type monitor struct {
	status enum.StatusID
	bout   config.Bout
	stop   chan struct{}
	done   chan struct{}

	log *logrus.Logger
	cfg *config.Config

	roomOn bool
}

func (m *monitor) Start() error {
	if !atomic.CompareAndSwapUint32(&m.status, enum.Status.Starting, enum.Status.Pending) {
		return nil
	}

	m.log.WithFields(logrus.Fields{
		"pf": m.bout.GetPlatform(),
		"id": m.bout.GetRoomID(),
	}).Info("monitor start")

	defer atomic.CompareAndSwapUint32(&m.status, enum.Status.Pending, enum.Status.Running)
	m.refresh()

	go m.run()

	return nil
}

func (m *monitor) Stop() {
	if !atomic.CompareAndSwapUint32(&m.status, enum.Status.Running, enum.Status.Stopping) {
		return
	}
	close(m.stop)
}

func (m *monitor) refresh() {
	if err := m.bout.Snap(); err != nil {
		m.log.WithFields(logrus.Fields{
			"pf": m.bout.GetPlatform(),
			"id": m.bout.GetRoomID(),
		}).Tracef("snap failed, %s", err.Error())
		return
	}
	_, roomOn := m.bout.StreamURL()
	defer func() {
		m.roomOn = roomOn
	}()
	var eventType enum.EventTypeID
	if !m.roomOn && roomOn {
		eventType = enum.EventType.AddRecorder
	} else {
		return
	}

	m.log.WithFields(logrus.Fields{
		"pf":  m.bout.GetPlatform(),
		"id":  m.bout.GetRoomID(),
		"old": m.roomOn,
		"new": roomOn,
	}).Info("live status changed")

	d, ok := dispatcher.SharedManager.Dispatcher(enum.DispatcherType.Recorder)
	if !ok {
		return
	}
	e := dispatcher.NewEvent(eventType, m.bout)
	if err := d.Dispatch(e); err != nil {
		m.log.Error(err)
	}

}

func (m *monitor) run() {
	t := jitterbug.New(
		time.Second*time.Duration(m.cfg.SnapRestSeconds),
		&jitterbug.Norm{Stdev: time.Second * 3},
	)
	defer t.Stop()

	for {
		select {
		case <-m.stop:
			close(m.done)
			m.log.WithFields(logrus.Fields{
				"pf": m.bout.GetPlatform(),
				"id": m.bout.GetRoomID(),
			}).Info("monitor stop")
			return
		case <-t.C:
			m.refresh()
		}
	}
}

func (m *monitor) Done() <-chan struct{} {
	return m.done
}
