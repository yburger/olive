package main

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
	"github.com/go-olive/olive/business/core/config"
	"github.com/go-olive/olive/business/core/show"
	"github.com/go-olive/olive/business/sys/database"
	"github.com/go-olive/olive/engine/kernel"
	l "github.com/go-olive/olive/engine/log"
	"github.com/go-olive/olive/foundation/logger"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// Construct the application logger.
	log, err := logger.New("OLIVE-API")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	// Perform the startup and shutdown sequence.
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

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
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:localhost"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:0"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "Copyright 2022 luxcgo",
		},
	}

	const prefix = "OLIVE"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

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

	db, err := database.Open(database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Infow("shutdown", "status", "stopping database support", "host", cfg.DB.Host)
		db.Close()
	}()

	// =========================================================================
	// Start Engine

	log.Infow("startup", "status", "initializing olive engine")
	l.Logger.Infof("Powered by go-olive/olive %s", build)

	configCore := config.NewCore(log, db)
	ctx1, cancel := context.WithTimeout(context.Background(), cfg.Web.ReadTimeout)
	defer cancel()
	engineConfig, err := configCore.QueryEngineConfig(ctx1)
	if err != nil {
		return fmt.Errorf("query engine config: %w", err)
	}

	showCore := show.NewCore(log, db)
	ctx2, cancel := context.WithTimeout(context.Background(), cfg.Web.ReadTimeout)
	defer cancel()
	showsEnabled, err := showCore.QueryAllEnabled(ctx2)
	if err != nil {
		return fmt.Errorf("query shows enabled: %w", err)
	}

	k := kernel.New(l.Logger, engineConfig, showsEnabled)
	go func() {
		k.Run()
	}()

	// todo(lc): timing is somewhat wrong
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()
		k.Shutdown(ctx)
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
