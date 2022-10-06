// Package showgrp maintains the group of handlers for show access.
package showgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-olive/olive/business/core/show"
	v1Web "github.com/go-olive/olive/business/web/v1"
	"github.com/go-olive/olive/business/web/v1/mid"
	"github.com/go-olive/olive/foundation/web"
)

// Handlers manages the set of show endpoints.
type Handlers struct {
	Show show.Core
}

// Create adds a new show to the system.
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var newShow show.NewShow
	if err := web.Decode(r, &newShow); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	s, err := h.Show.Create(ctx, newShow, v.Now)
	if err != nil {
		return fmt.Errorf("show[%+v]: %w", &s, err)
	}

	return mid.Respond(ctx, w, s, http.StatusCreated)
}

// Update updates a show in the system.
func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var upd show.UpdateShow
	if err := web.Decode(r, &upd); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	showID := web.Param(r, "id")

	if err := h.Show.Update(ctx, showID, upd, v.Now); err != nil {
		switch {
		case errors.Is(err, show.ErrInvalidPostCmds),
			errors.Is(err, show.ErrInvalidSplitRule):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, show.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s] Show[%+v]: %w", showID, &upd, err)
		}
	}

	return mid.Respond(ctx, w, nil, http.StatusOK)
}

// Delete removes one or many shows from the system.
func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	showID := web.Param(r, "id")

	if err := h.Show.Delete(ctx, showID); err != nil {
		switch {
		case errors.Is(err, show.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("ID[%s]: %w", showID, err)
		}
	}

	return mid.Respond(ctx, w, nil, http.StatusOK)
}

// Query returns a list of shows with paging.
func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := web.Param(r, "pageIndex")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid page format [%s]", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "pageSize")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid rows format [%s]", rows), http.StatusBadRequest)
	}

	shows, err := h.Show.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return fmt.Errorf("unable to query for shows: %w", err)
	}

	num, err := h.Show.TotalNum(ctx)
	if err != nil {
		return fmt.Errorf("unable to query for total number: %w", err)
	}

	data := struct {
		Total int64       `json:"total"`
		List  []show.Show `json:"list"`
	}{
		Total: num,
		List:  shows,
	}

	return mid.Respond(ctx, w, data, http.StatusOK)
}

// QueryByID returns a show by its ID.
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	showID := web.Param(r, "id")

	s, err := h.Show.QueryByID(ctx, showID)
	if err != nil {
		switch {
		case errors.Is(err, show.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, show.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", showID, err)
		}
	}

	return web.Respond(ctx, w, s, http.StatusOK)
}
