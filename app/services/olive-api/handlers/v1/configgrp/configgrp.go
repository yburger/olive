// Package configgrp maintains the group of handlers for config access.
package configgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-olive/olive/business/core/config"
	v1Web "github.com/go-olive/olive/business/web/v1"
	"github.com/go-olive/olive/business/web/v1/mid"
	"github.com/go-olive/olive/engine/kernel"
	"github.com/go-olive/olive/foundation/web"
)

// Handlers manages the set of config endpoints.
type Handlers struct {
	Config config.Core
	K      *kernel.Kernel
}

// Create adds a new config to the system.
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var newConfig config.NewConfig
	if err := web.Decode(r, &newConfig); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	s, err := h.Config.Create(ctx, newConfig, v.Now)
	if err != nil {
		return fmt.Errorf("config[%+v]: %w", &s, err)
	}

	return mid.Respond(ctx, w, s, http.StatusCreated)
}

// Update updates a config in the system.
func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var upd config.UpdateConfig
	if err := web.Decode(r, &upd); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	configKey := web.Param(r, "key")

	if err := h.Config.Update(ctx, configKey, upd, v.Now); err != nil {
		switch {
		case errors.Is(err, config.ErrInvalidConfig):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, config.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("config key[%s] Config[%+v]: %w", configKey, &upd, err)
		}
	}

	h.K.UpdateConfig(configKey, *upd.Value)

	return mid.Respond(ctx, w, nil, http.StatusOK)
}

// Delete removes one or many configs from the system.
func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	configKey := web.Param(r, "key")

	if err := h.Config.Delete(ctx, configKey); err != nil {
		switch {
		default:
			return fmt.Errorf("config key[%s]: %w", configKey, err)
		}
	}

	return mid.Respond(ctx, w, nil, http.StatusOK)
}

// QueryByKey returns a config by its Key.
func (h Handlers) QueryByKey(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	configKey := web.Param(r, "key")
	s, err := h.Config.QueryByKey(ctx, configKey)
	if err != nil {
		switch {
		case errors.Is(err, config.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("config key[%s]: %w", configKey, err)
		}
	}

	return mid.Respond(ctx, w, s, http.StatusOK)
}
