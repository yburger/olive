package recorder

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/go-olive/olive/engine/config"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	mu     sync.RWMutex
	savers map[config.ID]Recorder
	stop   chan struct{}

	log *logrus.Logger
	cfg *config.Config
}

func NewManager(log *logrus.Logger, cfg *config.Config) *Manager {
	return &Manager{
		savers: make(map[config.ID]Recorder),
		stop:   make(chan struct{}),
		log:    log,
		cfg:    cfg,
	}
}

func (m *Manager) Stop() {
	close(m.stop)
	for _, recorder := range m.savers {
		recorder.Stop()
		<-recorder.Done()
	}
}

func (m *Manager) addRecorder(bout config.Bout) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.savers[bout.GetID()]; ok {
		return errors.New("exist")
	}
	recorder, err := NewRecorder(m.log, bout)
	if err != nil {
		return err
	}
	m.savers[bout.GetID()] = recorder
	return recorder.Start()
}

func (m *Manager) removeRecorder(bout config.Bout) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	recorder, ok := m.savers[bout.GetID()]
	if !ok {
		return errors.New("recorder not exist")
	}
	recorder.Stop()
	delete(m.savers, bout.GetID())
	return nil
}

type Splitter interface {
	Split()
}

func (m *Manager) Split() {
	m.log.Info("split program starts...")

	t := time.NewTicker(time.Second * time.Duration(m.cfg.SplitRestSeconds))
	defer t.Stop()

	for {
		select {
		case <-m.stop:
			return
		case <-t.C:
			for _, r := range m.savers {
				if r.Bout().SatisfySplitRule(r.StartTime(), r.Out()) {
					m.log.WithFields(logrus.Fields{
						"pf": r.Bout().GetPlatform(),
						"id": r.Bout().GetRoomID(),
					}).Info("restart by split program")
					r.Bout().RestartRecorder()
				}
			}
		}
	}
}

func (m *Manager) MonitorParserStatus() {
	m.log.Info("parser-monitor program starts...")

	const cacheSize = 1024
	cache := freecache.NewCache(cacheSize)

	t := time.NewTicker(time.Second * time.Duration(m.cfg.ParserMonitorRestSeconds))
	defer t.Stop()

	for {
		select {
		case <-m.stop:
			return
		case <-t.C:

			// m.log.Debug("---------------------")
			// m.log.Debugf("EntryCount = %d", cache.EntryCount())
			// iter := cache.NewIterator()
			// for entry := iter.Next(); entry != nil; entry = iter.Next() {
			// 	m.log.Debugf("key = %s, val = %s", entry.Key, entry.Value)
			// }
			// m.log.Debug("---------------------")

			for _, r := range m.savers {
				fi, err := os.Stat(r.Out())
				if err != nil {
					continue
				}
				curSize := fi.Size()

				func() {
					defer func() {
						cache.Set([]byte(r.Out()), []byte(strconv.FormatInt(curSize, 10)), int(m.cfg.ParserMonitorRestSeconds+1))
					}()

					preSizeBytes, err := cache.Peek([]byte(r.Out()))
					if err != nil {
						// m.log.Error(err)
						return
					}

					preSize, err := strconv.ParseInt(string(preSizeBytes), 10, 64)
					if err != nil {
						m.log.Error(err)
						return
					}

					if curSize <= preSize {
						m.log.WithFields(logrus.Fields{
							"pf": r.Bout().GetPlatform(),
							"id": r.Bout().GetRoomID(),
						}).Info("restart by parser-monitor program")
						r.Bout().RestartRecorder()
					}
				}()

			}
		}
	}
}
