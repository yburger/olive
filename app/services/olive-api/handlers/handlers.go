// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/go-olive/olive/app/services/olive-api/handlers/debug/checkgrp"
	v1 "github.com/go-olive/olive/app/services/olive-api/handlers/v1"
	"github.com/go-olive/olive/business/web/v1/mid"
	"github.com/go-olive/olive/engine/kernel"
	"github.com/go-olive/olive/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// DebugStandardLibraryMux registers all the debug routes from the standard library
// into a new mux bypassing the use of the DefaultServerMux. Using the
// DefaultServerMux would be a security risk since a dependency could inject a
// handler into our service without us knowing it.
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// DebugMux registers all the debug standard library routes and then custom
// debug application routes for the service. This bypassing the use of the
// DefaultServerMux. Using the DefaultServerMux would be a security risk since
// a dependency could inject a handler into our service without us knowing it.
func DebugMux(build string, log *zap.SugaredLogger, db *sqlx.DB) http.Handler {
	mux := DebugStandardLibraryMux()

	// Register debug check endpoints.
	cgh := checkgrp.Handlers{
		Build: build,
		Log:   log,
		DB:    db,
	}
	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux
}

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	DB       *sqlx.DB
	K        *kernel.Kernel
}

// APIMux constructs an http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) *web.App {

	// Construct the web.App which holds all routes.
	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Panics(),
	)

	// Load the v1 routes.
	v1.Routes(app, v1.Config{
		Log: cfg.Log,
		DB:  cfg.DB,
		K:   cfg.K,
	})

	return app
}
