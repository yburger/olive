package kernel

import (
	"context"

	"github.com/go-olive/olive/engine/config"
	"github.com/go-olive/olive/engine/dispatcher"
	"github.com/go-olive/olive/engine/monitor"
	"github.com/go-olive/olive/engine/recorder"
	"github.com/go-olive/olive/foundation/syncmap"
	"github.com/sirupsen/logrus"
)

type Kernel struct {
	log     *logrus.Logger
	cfg     *config.Config
	showMap *syncmap.RWMap[string, Show]

	recorderManager *recorder.Manager
	monitorManager  *monitor.Manager

	done chan struct{}
}

func New(log *logrus.Logger, cfg *config.Config, shows []Show) *Kernel {
	showMap := syncmap.NewRWMap[string, Show](len(shows))
	for _, show := range shows {
		showMap.Set(show.ID, show)
	}

	recorderManager := recorder.NewManager(log, cfg)
	monitorManager := monitor.NewManager(log, cfg)
	dispatcher.SharedManager = dispatcher.NewManager(log)
	dispatcher.SharedManager.Register(recorderManager, monitorManager)

	return &Kernel{
		log:     log,
		cfg:     cfg,
		showMap: showMap,

		recorderManager: recorderManager,
		monitorManager:  monitorManager,

		done: make(chan struct{}),
	}
}

func (k *Kernel) UpdateShow(...Show) {
	// todo(lc)
}

func (k *Kernel) UpdateConfig(cfg config.Config) {
	*k.cfg = cfg
}

func (k *Kernel) Run() {
	k.showMap.Each(func(showID string, _ Show) bool {
		bout, err := NewBout(showID, k.showMap, k.cfg)
		if err != nil {
			k.log.Error(err)
		}
		bout.AddMonitor()
		return true
	})

	go k.recorderManager.Split()
	go k.recorderManager.MonitorParserStatus()
}

func (k *Kernel) Shutdown(ctx context.Context) {
	k.recorderManager.Stop()
	k.monitorManager.Stop()
	close(k.done)
}

func (k *Kernel) Done() <-chan struct{} {
	return k.done
}
