package uploader

import (
	"os/exec"
	"path/filepath"

	"github.com/go-olive/olive/engine/config"
	"github.com/sirupsen/logrus"
)

func (wp *WorkerPool) BiliupPrerun() {
	files, err := filepath.Glob(filepath.Join(wp.cfg.SaveDir, "*.flv"))
	if err != nil {
		return
	}
	tasks := make([]*TaskGroup, len(files))
	for i, filepath := range files {
		tasks[i] = &TaskGroup{
			Filepath: filepath,
			PostCmds: []*exec.Cmd{
				{Path: olivebiliup},
				{Path: olivetrash},
			},
			cfg: wp.cfg,
		}
	}
	wp.AddTask(tasks...)
}

type WorkerPool struct {
	log         *logrus.Logger
	cfg         *config.Config
	concurrency uint
	workers     []*worker
	uploadTasks chan *TaskGroup
	stopChan    chan struct{}
}

func NewWorkerPool(log *logrus.Logger, concurrency uint, cfg *config.Config) *WorkerPool {
	wp := &WorkerPool{
		log:         log,
		cfg:         cfg,
		concurrency: concurrency,
		uploadTasks: make(chan *TaskGroup, 1024),
		stopChan:    make(chan struct{}),
	}
	for i := uint(0); i < wp.concurrency; i++ {
		w := newWorker(log, cfg, i)
		wp.workers = append(wp.workers, w)
	}
	return wp
}

func (wp *WorkerPool) AddTask(tasks ...*TaskGroup) {
	for _, t := range tasks {
		select {
		case <-wp.stopChan:
			return
		default:
			wp.uploadTasks <- t
		}
	}
}

func (wp *WorkerPool) Run() {
	for _, worker := range wp.workers {
		go worker.start(wp.uploadTasks)
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.stopChan)
	close(wp.uploadTasks)
	for _, worker := range wp.workers {
		worker.stop()
		<-worker.done()
	}
}
