// Package db contains show related CRUD functionality.
package db

import (
	"context"
	"fmt"

	"github.com/go-olive/olive/business/sys/database"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Store manages the set of APIs for show access.
type Store struct {
	log          *zap.SugaredLogger
	tr           database.Transactor
	db           sqlx.ExtContext
	isWithinTran bool
}

// NewStore constructs a data for api access.
func NewStore(log *zap.SugaredLogger, db *sqlx.DB) Store {
	return Store{
		log: log,
		tr:  db,
		db:  db,
	}
}

// WithinTran runs passed function and do commit/rollback at the end.
func (s Store) WithinTran(ctx context.Context, fn func(sqlx.ExtContext) error) error {
	if s.isWithinTran {
		return fn(s.db)
	}
	return database.WithinTran(ctx, s.log, s.tr, fn)
}

// Tran return new Store with transaction in it.
func (s Store) Tran(tx sqlx.ExtContext) Store {
	return Store{
		log:          s.log,
		tr:           s.tr,
		db:           tx,
		isWithinTran: true,
	}
}

// Create inserts a new show into the database.
func (s Store) Create(ctx context.Context, show Show) error {
	const q = `
	INSERT INTO shows
		(show_id, enable, platform, room_id, streamer_name, out_tmpl, parser, save_dir, post_cmds, split_rule, date_created, date_updated)
	VALUES
		(:show_id, :enable, :platform, :room_id, :streamer_name, :out_tmpl, :parser, :save_dir, :post_cmds, :split_rule, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, show); err != nil {
		return fmt.Errorf("inserting show: %w", err)
	}

	return nil
}

// Update replaces a show document in the database.
func (s Store) Update(ctx context.Context, show Show) error {
	const q = `
	UPDATE
		shows
	SET 
		"enable" = :enable,
		"platform" = :platform,
		"room_id" = :room_id,
		"streamer_name" = :streamer_name,
		"out_tmpl" = :out_tmpl,
		"parser" = :parser,
		"save_dir" = :save_dir,
		"post_cmds" = :post_cmds,
		"split_rule" = :split_rule,
		"date_updated" = :date_updated
	WHERE
		show_id = :show_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, show); err != nil {
		return fmt.Errorf("updating showID[%s]: %w", show.ID, err)
	}

	return nil
}

// Delete removes a show from the database.
func (s Store) Delete(ctx context.Context, showID string) error {
	data := struct {
		ShowID string `db:"show_id"`
	}{
		ShowID: showID,
	}

	const q = `
	DELETE FROM
		shows
	WHERE
		show_id = :show_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("deleting showID[%s]: %w", showID, err)
	}

	return nil
}

// Query retrieves a list of existing shows from the database.
func (s Store) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Show, error) {
	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		*
	FROM
		shows
	ORDER BY
		show_id
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var shows []Show
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, data, &shows); err != nil {
		return nil, fmt.Errorf("selecting shows: %w", err)
	}

	return shows, nil
}

// QueryByID gets the specified show from the database.
func (s Store) QueryByID(ctx context.Context, showID string) (Show, error) {
	data := struct {
		ShowID string `db:"show_id"`
	}{
		ShowID: showID,
	}

	const q = `
	SELECT
		*
	FROM
		shows
	WHERE 
		show_id = :show_id`

	var show Show
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &show); err != nil {
		return Show{}, fmt.Errorf("selecting showID[%q]: %w", showID, err)
	}

	return show, nil
}
