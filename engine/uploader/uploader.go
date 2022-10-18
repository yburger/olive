// Package uploader handles post cmds when show ends.
package uploader

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/go-olive/olive/engine/config"
	"github.com/sirupsen/logrus"
)

type Uploader interface {
	proc()
	stop()
	done() <-chan struct{}
}

type TaskGroup struct {
	Filepath string
	PostCmds []*exec.Cmd

	cfg *config.Config
}

type uploader struct {
	log       *logrus.Logger
	cfg       *config.Config
	taskGroup *TaskGroup
	closeOnce sync.Once
	stopChan  chan struct{}
	doneChan  chan struct{}
}

func NewUploader(log *logrus.Logger, cfg *config.Config, taskGroup *TaskGroup) Uploader {
	return &uploader{
		log:       log,
		cfg:       cfg,
		taskGroup: taskGroup,
		stopChan:  make(chan struct{}),
		doneChan:  make(chan struct{}),
	}
}

func (u *uploader) proc() {
	defer close(u.doneChan)

	for _, postCmd := range u.taskGroup.PostCmds {
		select {
		case <-u.stopChan:
			return
		default:
			u.log.WithFields(logrus.Fields{
				"postCmdPath": postCmd.Path,
				"postCmdArgs": strings.Join(postCmd.Args, " "),
				"filepath":    u.taskGroup.Filepath,
			}).Info("cmd start running")
			handler := DefaultTaskMux.MustGetHandler(postCmd.Path)
			err := handler.Process(
				&Task{
					log:      u.log,
					cfg:      u.cfg,
					Filepath: u.taskGroup.Filepath,
					StopChan: u.stopChan,
					Cmd:      postCmd,
				},
			)
			if err != nil {
				u.log.WithFields(logrus.Fields{
					"postCmdPath": postCmd.Path,
					"postCmdArgs": strings.Join(postCmd.Args, " "),
					"filepath":    u.taskGroup.Filepath,
				}).Error(err)
				return
			}
		}
	}
}

func (u *uploader) stop() {
	u.closeOnce.Do(func() {
		close(u.stopChan)
	})
}

func (u *uploader) done() <-chan struct{} {
	return u.doneChan
}
