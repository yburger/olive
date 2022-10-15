// Package recorder records streams.
package recorder

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/go-olive/olive/engine/config"
	"github.com/go-olive/olive/engine/enum"
	"github.com/go-olive/olive/engine/parser"
	"github.com/sirupsen/logrus"
)

type Recorder interface {
	Start() error
	Stop()
	StartTime() time.Time
	Done() <-chan struct{}
	Out() string
	Bout() config.Bout
}

type recorder struct {
	status    enum.StatusID
	bout      config.Bout
	stop      chan struct{}
	startTime time.Time
	parser    parser.Parser
	done      chan struct{}
	out       string
	log       *logrus.Logger
}

func NewRecorder(log *logrus.Logger, bout config.Bout) (Recorder, error) {
	return &recorder{
		status:    enum.Status.Starting,
		bout:      bout,
		stop:      make(chan struct{}),
		startTime: time.Now(),
		done:      make(chan struct{}),
		log:       log,
	}, nil
}

func (r *recorder) Start() error {
	if !atomic.CompareAndSwapUint32(&r.status, enum.Status.Starting, enum.Status.Pending) {
		return nil
	}
	defer atomic.CompareAndSwapUint32(&r.status, enum.Status.Pending, enum.Status.Running)
	go r.run()

	r.log.WithFields(logrus.Fields{
		"pf": r.bout.GetPlatform(),
		"id": r.bout.GetRoomID(),
	}).Info("recorder start")

	return nil
}

func (r *recorder) Stop() {
	if !atomic.CompareAndSwapUint32(&r.status, enum.Status.Running, enum.Status.Stopping) {
		return
	}
	close(r.stop)
	if r.parser != nil {
		r.parser.Stop()
	}
}

func (r *recorder) StartTime() time.Time {
	return r.startTime
}

func (r *recorder) Out() string {
	return r.out
}

func (r *recorder) Bout() config.Bout {
	return r.bout
}

func (r *recorder) record() error {
	newParser, exist := parser.SharedManager.Parser(r.bout.GetParser())
	if !exist {
		return fmt.Errorf("parser[%s] does not exist", r.bout.GetParser())
	}
	r.parser = newParser.New()

	var out string
	defer func() {
		fi, err := os.Stat(out)
		if err != nil {
			r.log.Errorf("rm small file failed(stat): %+v", err)
			return
		}
		const oneMB = 1e6
		if fi.Size() < oneMB {
			if err := os.Remove(out); err != nil {
				r.log.WithFields(logrus.Fields{
					"filename": fi.Name(),
					"filesize": fi.Size(),
				}).Errorf("rm small file failed: %+v", err)
			}
			return
		}

		r.SubmitUploadTask(out, r.bout.GetPostCmds())
	}()

	const retry = 3
	var streamURL string
	var ok bool
	for i := 0; i < retry; i++ {
		err := r.bout.Snap()
		if err == nil {
			if streamURL, ok = r.bout.StreamURL(); ok {
				break
			} else {
				err = errors.New("empty stream url")
			}
		}
		r.log.WithFields(logrus.Fields{
			"pf":  r.bout.GetPlatform(),
			"id":  r.bout.GetRoomID(),
			"cnt": i + 1,
		}).Errorf("snap failed, %s", err.Error())

		if i == retry-1 {
			return err
		}
		time.Sleep(5 * time.Second)
	}

	roomName, _ := r.bout.RoomName()
	out = r.bout.GetOutFilename()

	r.log.WithFields(logrus.Fields{
		"pf": r.bout.GetPlatform(),
		"id": r.bout.GetRoomID(),
		"rn": roomName,
	}).Info("record start")

	saveDir := r.bout.GetSaveDir()
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		r.log.WithFields(logrus.Fields{
			"pf": r.bout.GetPlatform(),
			"id": r.bout.GetRoomID(),
		}).Errorf("mkdir failed: %s", err.Error())
		return nil
	}

	out = filepath.Join(saveDir, out)

	switch r.parser.Type() {
	case "yt-dlp":
		ext := filepath.Ext(out)
		out = out[0:len(out)-len(ext)] + ".mp4"
	default:
		ext := filepath.Ext(out)
		out = out[0:len(out)-len(ext)] + ".flv"
	}

	r.startTime = time.Now()
	r.out = out

	err := r.parser.Parse(streamURL, out)

	r.log.WithFields(logrus.Fields{
		"pf": r.bout.GetPlatform(),
		"id": r.bout.GetRoomID(),
	}).Infof("record stop: %+v", err)

	return nil
}

func (r *recorder) run() {
	r.bout.RemoveMonitor()

	defer func() {
		select {
		case <-r.stop:
		default:
			r.bout.AddMonitor()
		}
	}()

	for {
		select {
		case <-r.stop:
			close(r.done)
			r.log.WithFields(logrus.Fields{
				"pf": r.bout.GetPlatform(),
				"id": r.bout.GetRoomID(),
			}).Info("recorder stop")
			return
		default:
			if err := r.record(); err != nil {
				return
			}
		}
	}
}

func (r *recorder) Done() <-chan struct{} {
	return r.done
}

func (r *recorder) SubmitUploadTask(filepath string, cmds []*exec.Cmd) {
	// todo(lc)
	// if len(cmds) > 0 && filepath != "" {
	// 	uploader.UploaderWorkerPool.AddTask(&uploader.TaskGroup{
	// 		Filepath: filepath,
	// 		PostCmds: cmds,
	// 	})
	// }
}
