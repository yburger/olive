package command

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/go-olive/olive/app/services/olive-api/handlers"
	"github.com/go-olive/olive/app/tooling/olive-admin/commands"
	"github.com/go-olive/olive/business/core/config"
	"github.com/go-olive/olive/business/core/show"
	"github.com/go-olive/olive/business/sys/database"
	"github.com/go-olive/olive/engine/kernel"
	l "github.com/go-olive/olive/engine/log"
	"github.com/go-olive/olive/foundation/logger"
	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

var _ cmder = (*serverCmd)(nil)

type serverCmd struct {
	Web
	DB

	logDir  string
	saveDir string

	*baseBuilderCmd
}

type Web struct {
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	APIHost         string
	DebugHost       string
}

type DB struct {
	User         string
	Password     string `conf:"mask"`
	Host         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

func (b *commandsBuilder) newServerCmd() *serverCmd {
	cc := &serverCmd{}
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Server provides olive-api support.",
		Long:  "Server provides olive-api support.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run()
		},
	}
	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	cmd.Flags().DurationVar(&cc.ReadTimeout, "web-read-timeout", 5*time.Second, "")
	cmd.Flags().DurationVar(&cc.ReadTimeout, "web-write-timeout", 10*time.Second, "")
	cmd.Flags().DurationVar(&cc.ReadTimeout, "web-idle-timeout", 120*time.Second, "")
	cmd.Flags().DurationVar(&cc.ReadTimeout, "web-shutdown-timeout", 20*time.Second, "")
	cmd.Flags().StringVar(&cc.APIHost, "web-api-host", "0.0.0.0:3000", "")
	cmd.Flags().StringVar(&cc.DebugHost, "web-debug-host", "0.0.0.0:4000", "")

	cmd.Flags().StringVar(&cc.User, "db-user", "postgres", "")
	cmd.Flags().StringVar(&cc.Password, "db-password", "postgres", "")
	cmd.Flags().StringVar(&cc.Host, "db-host", "localhost", "")
	cmd.Flags().StringVar(&cc.Name, "db-name", "postgres", "")
	cmd.Flags().IntVar(&cc.MaxIdleConns, "db-max-idle-conns", 0, "")
	cmd.Flags().IntVar(&cc.MaxOpenConns, "db-max-open-conns", 0, "")
	cmd.Flags().BoolVar(&cc.DisableTLS, "db-disable-tls", true, "")

	cmd.Flags().StringVarP(&cc.logDir, "logdir", "l", "", "log file directory")
	cmd.Flags().StringVarP(&cc.saveDir, "savedir", "s", "", "video file directory")

	return cc
}

func (c *serverCmd) run() error {
	// Construct the application logger.
	log, err := logger.New("OLIVE-API")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	// Perform the startup and shutdown sequence.
	cfg := cfg{
		Web: c.Web,
		DB:  c.DB,
	}
	if err := c.serve(log, cfg); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
	return nil
}

type cfg struct {
	Web
	DB
}

func (c *serverCmd) serve(log *zap.SugaredLogger, cfg cfg) (err error) {

	// =========================================================================
	// GOMAXPROCS

	// Want to see what maxprocs reports.
	opt := maxprocs.Logger(log.Infof)

	// Set the correct number of threads for the service
	// based on what is available either by the machine or quotas.
	if _, err := maxprocs.Set(opt); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// =========================================================================
	// App Starting

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	expvar.NewString("build").Set(build)

	// =========================================================================
	// Database Support

	// Create connectivity to the database.
	log.Infow("startup", "status", "initializing database support", "host", cfg.DB.Host)

	dbConfig := database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	}
	db, err := database.Open(dbConfig)
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Infow("shutdown", "status", "stopping database support", "host", cfg.DB.Host)
		db.Close()
	}()

	// =========================================================================
	// Database Admin Operation

	if err := commands.Migrate(dbConfig); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	if err := commands.Seed(dbConfig); err != nil {
		return fmt.Errorf("seeding database: %w", err)
	}

	// =========================================================================
	// Start Engine
	log.Infow("startup", "status", "initializing olive engine")

	configCore := config.NewCore(log, db)
	ctx1, cancel := context.WithTimeout(context.Background(), cfg.Web.ReadTimeout)
	defer cancel()
	engineConfig, err := configCore.QueryEngineConfig(ctx1)
	if err != nil {
		return fmt.Errorf("query engine config: %w", err)
	}
	if c.logDir != "" {
		engineConfig.LogDir = c.logDir
	}
	if c.saveDir != "" {
		engineConfig.SaveDir = c.saveDir
	}
	engineConfig.CheckAndFix()
	engineLogger := l.InitLogger(engineConfig.LogDir)
	engineLogger.Infof("Powered by go-olive/olive %s", build)

	showCore := show.NewCore(log, db)
	ctx2, cancel := context.WithTimeout(context.Background(), cfg.Web.ReadTimeout)
	defer cancel()
	showsEnabled, err := showCore.QueryAllEnabled(ctx2)
	if err != nil {
		return fmt.Errorf("query shows enabled: %w", err)
	}

	k := kernel.New(engineLogger, engineConfig, showsEnabled)
	go func() {
		k.Run()
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()
		go func(ctx context.Context) {
			k.Shutdown(ctx)
		}(ctx)

		select {
		case <-ctx.Done():
			// trick, can change the err value even the funciton has returned.
			newErr := errors.New("engine timeout, force quit")
			if err != nil {
				err = fmt.Errorf("%v\n%v", err, newErr)
			} else {
				err = newErr
			}
		case <-k.Done():
		}
	}()

	// =========================================================================
	// Start Debug Service

	log.Infow("startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. This includes the standard library endpoints.

	// Construct the mux for the debug calls.
	debugMux := handlers.DebugMux(build, log, db)

	// Start the service listening for debug requests.
	// Not concerned with shutting this down with load shedding.
	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	// =========================================================================
	// Start API Service

	log.Infow("startup", "status", "initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Construct the mux for the API calls.
	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Log:      log,
		DB:       db,
		K:        k,
	})

	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		engineLogger.WithField("signal", sig.String()).
			Info("handle request")

		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
