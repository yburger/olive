package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/go-olive/olive/engine/config"
	"github.com/go-olive/olive/engine/kernel"
	l "github.com/go-olive/olive/engine/log"
	"github.com/sirupsen/logrus"
)

var build = "v0.5.0-src"

// todo(lc)
// olivectl run
// olivectl api
// olivectl admin

func main() {
	log := l.Logger
	log.Infof("Powered by go-olive/olive %s", build)

	if err := run(log); err != nil {
		log.Info(err)
		os.Exit(1)
	}
}

func run(log *logrus.Logger) error {
	var cfg config.Config
	conf.Parse("", &cfg)
	k := kernel.New(log, &cfg, []kernel.Show{
		{
			ID:           "1",
			Platform:     "huya",
			RoomID:       "327541",
			StreamerName: "test",
			Parser:       "flv",
			SaveDir:      "/Users/lucas/github/olive/",
			OutTmpl:      cfg.OutTmpl,
		},
	})

	// =========================================================================
	// Start

	go func() {
		k.Run()
	}()

	// =========================================================================
	// Shutdown

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdown
	log.WithField("signal", sig.String()).
		Info("handle request")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
	defer cancel()
	go func(ctx context.Context) {
		k.Shutdown(ctx)
	}(ctx)

	select {
	case <-ctx.Done():
		return errors.New("timeout, force quit")
	case <-k.Done():
		return nil
	}
}
