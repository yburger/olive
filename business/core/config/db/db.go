// Package db contains config related CRUD functionality.
package db

import (
	"context"
	"fmt"

	"github.com/go-olive/olive/business/sys/database"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Store manages the set of APIs for config access.
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

// Create inserts a new config into the database.
func (s Store) Create(ctx context.Context, config Config) error {
	const q = `
	INSERT INTO configs
		(key, value, date_created, date_updated)
	VALUES
		(:key, :value, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, config); err != nil {
		return fmt.Errorf("inserting config: %w", err)
	}

	return nil
}

// Update replaces a config document in the database.
func (s Store) Update(ctx context.Context, config Config) error {
	const q = `
	UPDATE
		configs
	SET 
		"value" = :value,
		"date_updated" = :date_updated
	WHERE
		key = :key`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, config); err != nil {
		return fmt.Errorf("updating config key[%s]: %w", config.Key, err)
	}

	return nil
}

// Delete removes a config from the database.
func (s Store) Delete(ctx context.Context, configKey string) error {
	data := struct {
		ConfigKey string `db:"key"`
	}{
		ConfigKey: configKey,
	}

	const q = `
	DELETE FROM
		configs
	WHERE
		key = :key`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("deleting config key[%s]: %w", configKey, err)
	}

	return nil
}

// QueryByKey gets the specified config from the database.
func (s Store) QueryByKey(ctx context.Context, configKey string) (Config, error) {
	data := struct {
		ConfigKey string `db:"key"`
	}{
		ConfigKey: configKey,
	}

	const q = `
	SELECT
		*
	FROM
		configs
	WHERE 
		key = :key`

	var config Config
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &config); err != nil {
		return Config{}, fmt.Errorf("selecting config key[%q]: %w", configKey, err)
	}

	return config, nil
}
