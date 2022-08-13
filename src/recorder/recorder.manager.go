package recorder

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/go-olive/olive/src/config"
	"github.com/go-olive/olive/src/engine"
	l "github.com/go-olive/olive/src/log"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	mu     sync.RWMutex
	savers map[engine.ID]Recorder
	stop   chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		savers: make(map[engine.ID]Recorder),
		stop:   make(chan struct{}),
	}
}

func (m *Manager) Stop() {
	close(m.stop)
	for _, recorder := range m.savers {
		recorder.Stop()
		<-recorder.Done()
	}
}

func (m *Manager) addRecorder(show engine.Show) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.savers[show.GetID()]; ok {
		return errors.New("exist")
	}
	recorder, err := NewRecorder(show)
	if err != nil {
		return err
	}
	m.savers[show.GetID()] = recorder
	return recorder.Start()
}

func (m *Manager) removeRecorder(show engine.Show) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	recorder, ok := m.savers[show.GetID()]
	if !ok {
		return errors.New("recorder not exist")
	}
	recorder.Stop()
	delete(m.savers, show.GetID())
	return nil
}

type Splitter interface {
	Split()
}

func (m *Manager) Split() {
	isValid := false
	for _, r := range config.APP.Shows {
		if r.SplitRule.IsValid() {
			isValid = true
			break
		}
	}
	if !isValid {
		return
	}

	l.Logger.Info("split program starts...")

	t := time.NewTicker(time.Second * time.Duration(config.APP.SplitRestSeconds))
	defer t.Stop()

	for {
		select {
		case <-m.stop:
			return
		case <-t.C:
			for _, r := range m.savers {
				if r.Show().SatisfySplitRule(r.StartTime(), r.Out()) {
					l.Logger.WithFields(logrus.Fields{
						"pf": r.Show().GetPlatform(),
						"id": r.Show().GetRoomID(),
					}).Info("restart by split program")
					r.Show().RestartRecorder()
				}
			}
		}
	}
}

func (m *Manager) MonitorParserStatus() {
	isValid := false
	for _, r := range config.APP.Shows {
		if r.Parser != "flv" {
			isValid = true
			break
		}
	}
	if !isValid {
		return
	}

	l.Logger.Info("parser-monitor program starts...")

	const cacheSize = 1024
	cache := freecache.NewCache(cacheSize)
	expire := config.APP.ParserMonitorRestSeconds + 1

	t := time.NewTicker(time.Second * time.Duration(config.APP.ParserMonitorRestSeconds))
	defer t.Stop()

	for {
		select {
		case <-m.stop:
			return
		case <-t.C:

			// l.Logger.Debug("---------------------")
			// l.Logger.Debugf("EntryCount = %d", cache.EntryCount())
			// iter := cache.NewIterator()
			// for entry := iter.Next(); entry != nil; entry = iter.Next() {
			// 	l.Logger.Debugf("key = %s, val = %s", entry.Key, entry.Value)
			// }
			// l.Logger.Debug("---------------------")

			for _, r := range m.savers {
				fi, err := os.Stat(r.Out())
				if err != nil {
					continue
				}
				curSize := fi.Size()

				func() {
					defer func() {
						cache.Set([]byte(r.Out()), []byte(strconv.FormatInt(curSize, 10)), int(expire))
					}()

					preSizeBytes, err := cache.Peek([]byte(r.Out()))
					if err != nil {
						// l.Logger.Error(err)
						return
					}

					preSize, err := strconv.ParseInt(string(preSizeBytes), 10, 64)
					if err != nil {
						l.Logger.Error(err)
						return
					}

					if curSize <= preSize {
						l.Logger.WithFields(logrus.Fields{
							"pf": r.Show().GetPlatform(),
							"id": r.Show().GetRoomID(),
						}).Info("restart by parser-monitor program")
						r.Show().RestartRecorder()
					}
					return
				}()

			}
		}
	}
}
