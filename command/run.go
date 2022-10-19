package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-olive/olive/engine/config"
	"github.com/go-olive/olive/engine/kernel"
	l "github.com/go-olive/olive/engine/log"
	"github.com/go-olive/olive/foundation/olivetv"
	jsoniter "github.com/json-iterator/go"
	"github.com/pelletier/go-toml/v2"
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
	cmd.Flags().StringVarP(&cc.roomURL, "url", "u", "", "room url")
	cmd.Flags().StringVarP(&cc.cookie, "cookie", "c", "", "http cookie")

	return cc
}

type CompositeConfig struct {
	Config config.Config
	Shows  []kernel.Show
}

func (cfg *CompositeConfig) checkAndFix() {
	cfg.Config.CheckAndFix()
	for _, show := range cfg.Shows {
		show.CheckAndFix(&cfg.Config)
	}
}

func (cfg *CompositeConfig) autosave() error {
	bytes, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	appTomlFile, err := os.Create(fmt.Sprintf("config-%d.toml", time.Now().Unix()))
	if err != nil {
		return err
	}
	defer appTomlFile.Close()

	appTomlFile.Write(bytes)
	return nil
}

func (c *runCmd) run() error {
	if c.roomURL != "" {
		return c.runWithURL()
	}
	viper.SetConfigFile(c.cfgFilepath)
	cfg := new(CompositeConfig)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return err
	}

	cfg.checkAndFix()

	log := l.InitLogger(cfg.Config.LogDir)
	k := kernel.New(log, &cfg.Config, cfg.Shows)

	// =========================================================================
	// Watch config change
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("config file[%s] is changed", e.Name)

		compoCfg := new(CompositeConfig)
		viper.Unmarshal(compoCfg)
		compoCfg.checkAndFix()

		cfgStr, _ := jsoniter.MarshalToString(compoCfg.Config)
		k.UpdateConfig(config.CoreConfigKey, cfgStr)
		for _, show := range compoCfg.Shows {
			k.UpdateShow(show)
		}
	})
	viper.WatchConfig()

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

func (c *runCmd) runWithURL() error {
	cc, err := newCompositeConfig(c.roomURL, c.cookie)
	if err != nil {
		return err
	}
	cc.autosave()

	log := l.InitLogger(cc.Config.LogDir)
	k := kernel.New(log, &cc.Config, cc.Shows)

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

func newCompositeConfig(roomURL, cookie string) (*CompositeConfig, error) {

	// initialize Shows
	if cookie != "" {
		config.DefaultConfig.DouyinCookie = cookie
		config.DefaultConfig.KuaishouCookie = cookie
	}

	var shows []kernel.Show
	tv, err := olivetv.NewWithURL(roomURL, olivetv.SetCookie(cookie))
	if err != nil {
		return nil, err
	}
	site, _ := olivetv.Sniff(tv.SiteID)
	show := kernel.Show{
		StreamerName: site.Name(),
		Platform:     tv.SiteID,
		RoomID:       tv.RoomID,
	}
	shows = []kernel.Show{show}

	// initialize CompositeConfigConfig
	cc := &CompositeConfig{
		Config: config.DefaultConfig,
		Shows:  shows,
	}
	cc.checkAndFix()

	return cc, nil
}
