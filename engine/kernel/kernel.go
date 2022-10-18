package kernel

import (
	"context"

	"github.com/go-olive/olive/engine/config"
	"github.com/go-olive/olive/engine/dispatcher"
	"github.com/go-olive/olive/engine/monitor"
	"github.com/go-olive/olive/engine/recorder"
	"github.com/go-olive/olive/engine/uploader"
	"github.com/go-olive/olive/foundation/syncmap"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type Kernel struct {
	log     *logrus.Logger
	cfg     *config.Config
	showMap *syncmap.RWMap[string, Show]

	recorderManager *recorder.Manager
	monitorManager  *monitor.Manager
	workerPool      *uploader.WorkerPool

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

	workerPool := uploader.NewWorkerPool(log, cfg.CommanderPoolSize, cfg)

	return &Kernel{
		log:     log,
		cfg:     cfg,
		showMap: showMap,

		recorderManager: recorderManager,
		monitorManager:  monitorManager,
		workerPool:      workerPool,

		done: make(chan struct{}),
	}
}

func (k *Kernel) HandleShow(shows ...Show) {
	// k.log.Println("============================")
	// old, _ := k.showMap.Get(shows[0].ID)
	// k.log.Println(jsoniter.MarshalToString(old))
	// k.log.Println(jsoniter.MarshalToString(shows))
	// k.log.Println("============================")

	for _, show := range shows {
		if show.Enable {
			k.UpdateShow(show)
		} else {
			k.DeleteShow(show)
		}
	}
}

func (k *Kernel) UpdateShow(shows ...Show) {
	for _, show := range shows {
		if _, ok := k.showMap.Get(show.ID); ok {
			k.showMap.Set(show.ID, show)
		} else {
			k.showMap.Set(show.ID, show)
			bout, err := NewBout(show.ID, k.showMap, k.cfg)
			if err != nil {
				k.log.Error(err)
				k.showMap.Delete(show.ID)
				continue
			}
			bout.AddMonitor()
		}
	}
}

func (k *Kernel) DeleteShow(shows ...Show) {
	for _, show := range shows {
		bout, err := NewBout(show.ID, k.showMap, k.cfg)
		if err != nil {
			k.log.Error(err)
			continue
		}
		bout.RemoveMonitor()
		bout.RemoveRecorder()
		k.showMap.Delete(show.ID)
	}
}

func (k *Kernel) UpdateConfig(key, value string) {
	switch key {
	case config.CoreConfigKey:
		var cfg config.Config
		if err := jsoniter.UnmarshalFromString(value, &cfg); err == nil {
			*k.cfg = cfg
		}
	}
}

func (k *Kernel) IsValidPortalUser(un, pw string) bool {
	return k.cfg.PortalUsername == un && k.cfg.PortalPassword == pw
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

	if k.cfg.BiliupEnable {
		k.workerPool.BiliupPrerun()
	}
	k.workerPool.Run()
}

func (k *Kernel) Shutdown(ctx context.Context) {
	k.recorderManager.Stop()
	k.monitorManager.Stop()
	close(k.done)
}

func (k *Kernel) Done() <-chan struct{} {
	return k.done
}
