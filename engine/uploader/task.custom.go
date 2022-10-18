package uploader

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-olive/olive/engine/util"
	"github.com/go-olive/olive/foundation/biliup"
	"github.com/sirupsen/logrus"
)

const (
	olivetrash   = "olivetrash"
	olivearchive = "olivearchive"
	olivebiliup  = "olivebiliup"
	oliveshell   = "oliveshell"
)

var DefaultHandlerFunc = TaskHandlerFunc(OliveDefault)

func init() {
	DefaultTaskMux.RegisterHandler(olivetrash, TaskHandlerFunc(OliveTrash))
	DefaultTaskMux.RegisterHandler(olivearchive, TaskHandlerFunc(OliveArchive))
	DefaultTaskMux.RegisterHandler(olivebiliup, TaskHandlerFunc(OliveBiliup))
	DefaultTaskMux.RegisterHandler(oliveshell, DefaultHandlerFunc)
}

func OliveTrash(t *Task) error {
	return os.Remove(t.Filepath)
}

func OliveArchive(t *Task) error {
	if err := os.MkdirAll("archive", os.ModePerm); err != nil {
		return err
	}
	return t.move("archive")
}

func (t *Task) move(dest string) error {
	if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dest, os.ModePerm)
		return err
	}

	base := filepath.Base(t.Filepath)
	dest = filepath.Join(dest, base)
	return util.MoveFile(t.Filepath, dest)
}

func OliveBiliup(t *Task) error {
	t.log.WithFields(logrus.Fields{
		"filepath": t.Filepath,
	}).Info("upload start")

	biliupConfig := biliup.Config{
		CookieFilepath: t.cfg.CookieFilepath,
		VideoFilepath:  t.Filepath,
		Threads:        t.cfg.Threads,
	}
	err := biliup.New(biliupConfig).Upload()
	if err == nil {
		t.log.WithFields(logrus.Fields{
			"filepath": t.Filepath,
		}).Info("upload succeed")
		return nil
	}

	return err
}

func OliveDefault(t *Task) error {
	doneChan := make(chan struct{})
	defer close(doneChan)

	if t.Cmd == nil || len(t.Cmd.Args) == 0 {
		return nil
	}

	cmd := exec.Command(t.Cmd.Args[0], t.Cmd.Args[1:]...)

	envFilepath := "FILE_PATH=" + t.Filepath
	cmd.Env = append([]string{envFilepath}, t.Cmd.Env...)
	cmd.Dir = t.Cmd.Dir

	go func() {
		select {
		case <-t.StopChan:
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			return
		case <-doneChan:
			return
		}
	}()

	resp, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	t.log.Infof("oliveshell success: %s", resp)
	return nil
}
