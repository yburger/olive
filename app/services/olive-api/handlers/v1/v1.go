// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/go-olive/olive/app/services/olive-api/handlers/v1/showgrp"
	"github.com/go-olive/olive/app/services/olive-api/handlers/v1/testgrp"
	"github.com/go-olive/olive/business/core/show"
	"github.com/go-olive/olive/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log *zap.SugaredLogger
	DB  *sqlx.DB
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	// Register show management and authentication endpoints.
	sgh := showgrp.Handlers{
		Show: show.NewCore(cfg.Log, cfg.DB),
	}
	app.Handle(http.MethodGet, version, "/shows/:id", sgh.QueryByID)
	app.Handle(http.MethodPost, version, "/shows", sgh.Create)
	app.Handle(http.MethodPut, version, "/shows/:id", sgh.Update)
	app.Handle(http.MethodDelete, version, "/shows/:id", sgh.Delete)

	// Register test endpoints.
	tgh := testgrp.Handlers{
		Log: cfg.Log,
	}
	app.Handle(http.MethodGet, version, "/test", tgh.Test)
}
