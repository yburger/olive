package uploader

import (
	"github.com/go-olive/olive/engine/config"
	"github.com/sirupsen/logrus"
)

type worker struct {
	log      *logrus.Logger
	cfg      *config.Config
	id       uint
	stopChan chan struct{}
	doneChan chan struct{}
	uploader Uploader
}

func (w *worker) done() <-chan struct{} {
	return w.doneChan
}

func newWorker(log *logrus.Logger, cfg *config.Config, id uint) *worker {
	return &worker{
		log:      log,
		cfg:      cfg,
		id:       id,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
}

func (w *worker) start(tasks <-chan *TaskGroup) {
	defer close(w.doneChan)

	for {
		task, ok := <-tasks
		if !ok {
			return
		}
		w.uploader = NewUploader(w.log, w.cfg, task)

		select {
		case <-w.stopChan:
			return
		default:
			w.uploader.proc()
			w.uploader = nil
		}
	}

}

func (w *worker) stop() {
	close(w.stopChan)
	if w.uploader != nil {
		w.uploader.stop()
	}
}
