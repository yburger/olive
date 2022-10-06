// Package show provides an example of a core business API. Right now these
// calls are just wrapping the data/data layer. But at some point you will
// want auditing or something that isn't specific to the data/store layer.
package show

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-olive/olive/business/core/show/db"
	"github.com/go-olive/olive/business/sys/database"
	"github.com/go-olive/olive/business/sys/validate"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound         = errors.New("show not found")
	ErrInvalidID        = errors.New("ID is not in its proper form")
	ErrInvalidPostCmds  = errors.New("PostCmds is not valid")
	ErrInvalidSplitRule = errors.New("SplitRule is not valid")
)

// Core manages the set of APIs for show access.
type Core struct {
	store db.Store
}

// NewCore constructs a core for show api access.
func NewCore(log *zap.SugaredLogger, sqlxDB *sqlx.DB) Core {
	return Core{
		store: db.NewStore(log, sqlxDB),
	}
}

// Create inserts a new show into the database.
func (c Core) Create(ctx context.Context, newShow NewShow, now time.Time) (Show, error) {
	if err := validate.Check(newShow); err != nil {
		return Show{}, fmt.Errorf("validating data: %w", err)
	}

	if err := validate.CheckPostCmds(newShow.PostCmds); err != nil {
		return Show{}, ErrInvalidPostCmds
	}
	if err := validate.CheckSplitRule(newShow.SplitRule); err != nil {
		return Show{}, ErrInvalidSplitRule
	}

	dbShow := db.Show{
		ID:           validate.GenerateID(),
		Enable:       newShow.Enable,
		Platform:     newShow.Platform,
		RoomID:       newShow.RoomID,
		StreamerName: newShow.StreamerName,
		OutTmpl:      newShow.OutTmpl,
		Parser:       newShow.Parser,
		SaveDir:      newShow.SaveDir,
		PostCmds:     newShow.PostCmds,
		SplitRule:    newShow.SplitRule,
		DateCreated:  now,
		DateUpdated:  now,
	}

	tran := func(tx sqlx.ExtContext) error {
		if err := c.store.Tran(tx).Create(ctx, dbShow); err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return nil
	}

	if err := c.store.WithinTran(ctx, tran); err != nil {
		return Show{}, fmt.Errorf("tran: %w", err)
	}

	return toShow(dbShow), nil
}

// Update replaces a show document in the database.
func (c Core) Update(ctx context.Context, showID string, updateShow UpdateShow, now time.Time) error {
	if err := validate.CheckID(showID); err != nil {
		return ErrInvalidID
	}

	if err := validate.Check(updateShow); err != nil {
		return fmt.Errorf("validating data: %w", err)
	}

	dbShow, err := c.store.QueryByID(ctx, showID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("updating show showID[%s]: %w", showID, err)
	}

	if updateShow.Enable != nil {
		dbShow.Enable = *updateShow.Enable
	}
	if updateShow.Platform != nil {
		dbShow.Platform = *updateShow.Platform
	}
	if updateShow.RoomID != nil {
		dbShow.RoomID = *updateShow.RoomID
	}
	if updateShow.StreamerName != nil {
		dbShow.StreamerName = *updateShow.StreamerName
	}
	if updateShow.OutTmpl != nil {
		dbShow.OutTmpl = *updateShow.OutTmpl
	}
	if updateShow.Parser != nil {
		dbShow.Parser = *updateShow.Parser
	}
	if updateShow.SaveDir != nil {
		dbShow.SaveDir = *updateShow.SaveDir
	}
	if updateShow.PostCmds != nil {
		dbShow.PostCmds = *updateShow.PostCmds
	}
	if updateShow.SplitRule != nil {
		dbShow.SplitRule = *updateShow.SplitRule
	}
	dbShow.DateUpdated = now

	if err := validate.CheckPostCmds(dbShow.PostCmds); err != nil {
		return ErrInvalidPostCmds
	}
	if err := validate.CheckSplitRule(dbShow.SplitRule); err != nil {
		return ErrInvalidSplitRule
	}

	if err := c.store.Update(ctx, dbShow); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// Delete removes a show from the database.
func (c Core) Delete(ctx context.Context, showID string) error {
	showIDList := strings.Split(showID, ",")
	for _, id := range showIDList {
		if err := validate.CheckID(id); err != nil {
			return fmt.Errorf("delete: %w showID:%s", ErrInvalidID, id)
		}
	}

	if err := c.store.Delete(ctx, showIDList); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Show, error) {
	dbShows, err := c.store.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return toShowSlice(dbShows), nil
}

// QueryByID gets the specified show from the database.
func (c Core) QueryByID(ctx context.Context, showID string) (Show, error) {
	if err := validate.CheckID(showID); err != nil {
		return Show{}, ErrInvalidID
	}

	dbShow, err := c.store.QueryByID(ctx, showID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Show{}, ErrNotFound
		}
		return Show{}, fmt.Errorf("query: %w", err)
	}

	return toShow(dbShow), nil
}

// TotalNum gets the total number of shows from the database.
func (c Core) TotalNum(ctx context.Context) (int64, error) {
	num, err := c.store.TotalNum(ctx)
	if err != nil {
		return 0, fmt.Errorf("query: %w", err)
	}

	return num, nil
}
