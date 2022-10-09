// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/go-olive/olive/app/services/olive-api/handlers/v1/configgrp"
	"github.com/go-olive/olive/app/services/olive-api/handlers/v1/showgrp"
	"github.com/go-olive/olive/app/services/olive-api/handlers/v1/testgrp"
	"github.com/go-olive/olive/app/services/olive-api/handlers/v1/usrgrp"
	"github.com/go-olive/olive/business/core/config"
	"github.com/go-olive/olive/business/core/show"
	"github.com/go-olive/olive/engine/kernel"
	"github.com/go-olive/olive/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log *zap.SugaredLogger
	DB  *sqlx.DB
	K   *kernel.Kernel
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	// Register show management and authentication endpoints.
	sgh := showgrp.Handlers{
		Show: show.NewCore(cfg.Log, cfg.DB),
		K:    cfg.K,
	}
	app.Handle(http.MethodGet, version, "/shows/:pageIndex/:pageSize", sgh.Query)
	app.Handle(http.MethodGet, version, "/shows/:id", sgh.QueryByID)
	app.Handle(http.MethodPost, version, "/shows", sgh.Create)
	app.Handle(http.MethodPut, version, "/shows/:id", sgh.Update)
	app.Handle(http.MethodDelete, version, "/shows/:id", sgh.Delete)

	// Register test endpoints.
	tgh := testgrp.Handlers{
		Log: cfg.Log,
	}
	app.Handle(http.MethodGet, version, "/test", tgh.Test)

	// Register user endpoints.
	ugh := usrgrp.Handlers{
		Log: cfg.Log,
		K:   cfg.K,
	}
	app.Handle(http.MethodPost, version, "/user/login", ugh.Login)
	app.Handle(http.MethodGet, version, "/user/logout", ugh.Logout)

	// Register config endpoints.
	cgh := configgrp.Handlers{
		Config: config.NewCore(cfg.Log, cfg.DB),
		K:      cfg.K,
	}
	app.Handle(http.MethodGet, version, "/configs/:key", cgh.QueryByKey)
	app.Handle(http.MethodPost, version, "/configs", cgh.Create)
	app.Handle(http.MethodPut, version, "/configs/:key", cgh.Update)
	app.Handle(http.MethodDelete, version, "/configs/:key", cgh.Delete)
}
