package command

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-olive/olive/engine/config"
	"github.com/go-olive/olive/engine/kernel"
	l "github.com/go-olive/olive/engine/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var _ cmder = (*runCmd)(nil)

type runCmd struct {
	cfgFilepath string
	roomURL     string
	cookie      string

	*baseBuilderCmd
}

func (b *commandsBuilder) newRunCmd() *runCmd {
	cc := &runCmd{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start the engine.",
		Long:  `Start the engine.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run()
		},
	}
	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	cmd.Flags().StringVarP(&cc.cfgFilepath, "filepath", "f", "", "set config.toml filepath")
	cmd.Flags().StringVarP(&cc.roomURL, "cookie", "c", "", "room url")
	cmd.Flags().StringVarP(&cc.cookie, "url", "u", "", "http cookie")

	return cc
}

func (c *runCmd) run() error {
	viper.SetConfigFile(c.cfgFilepath)
	var cfg = struct {
		config.Config
		Shows []kernel.Show
	}{}
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}
	cfg.Config.CheckAndFix()

	log := l.InitLogger(cfg.LogDir)
	k := kernel.New(log, &cfg.Config, cfg.Shows)

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
