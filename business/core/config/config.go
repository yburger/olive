// Package config provides business API for config.
package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-olive/olive/business/core/config/db"
	"github.com/go-olive/olive/business/sys/database"
	"github.com/go-olive/olive/business/sys/validate"
	"github.com/go-olive/olive/engine/config"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound      = errors.New("config not found")
	ErrInvalidConfig = errors.New("config is not in its proper form")
)

// Core manages the set of APIs for config access.
type Core struct {
	store db.Store
}

// NewCore constructs a core for config api access.
func NewCore(log *zap.SugaredLogger, sqlxDB *sqlx.DB) Core {
	return Core{
		store: db.NewStore(log, sqlxDB),
	}
}

// Create inserts a new config into the database.
func (c Core) Create(ctx context.Context, newConfig NewConfig, now time.Time) (Config, error) {
	if err := validate.Check(newConfig); err != nil {
		return Config{}, fmt.Errorf("validating data: %w", err)
	}

	if err := validate.CheckConfig(newConfig.Key, newConfig.Value); err != nil {
		return Config{}, ErrInvalidConfig
	}

	dbConfig := db.Config{
		Key:         newConfig.Key,
		Value:       newConfig.Value,
		DateCreated: now,
		DateUpdated: now,
	}

	tran := func(tx sqlx.ExtContext) error {
		if err := c.store.Tran(tx).Create(ctx, dbConfig); err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return nil
	}

	if err := c.store.WithinTran(ctx, tran); err != nil {
		return Config{}, fmt.Errorf("tran: %w", err)
	}

	return toConfig(dbConfig), nil
}

// Update replaces a config document in the database.
func (c Core) Update(ctx context.Context, configKey string, updateConfig UpdateConfig, now time.Time) error {
	if err := validate.Check(updateConfig); err != nil {
		return fmt.Errorf("validating data: %w", err)
	}

	dbConfig, err := c.store.QueryByKey(ctx, configKey)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("updating config configKey[%s]: %w", configKey, err)
	}

	if updateConfig.Value != nil {
		dbConfig.Value = *updateConfig.Value
	}
	dbConfig.DateUpdated = now

	if err := validate.CheckConfig(configKey, dbConfig.Value); err != nil {
		return ErrInvalidConfig
	}

	if err := c.store.Update(ctx, dbConfig); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// Delete removes a config from the database.
func (c Core) Delete(ctx context.Context, configKey string) error {
	if err := c.store.Delete(ctx, configKey); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// QueryByKey gets the specified config from the database.
func (c Core) QueryByKey(ctx context.Context, configKey string) (Config, error) {
	dbConfig, err := c.store.QueryByKey(ctx, configKey)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Config{}, ErrNotFound
		}
		return Config{}, fmt.Errorf("query: %w", err)
	}

	return toConfig(dbConfig), nil
}

// QueryEngineConfig gets the parsed engine config from the database.
func (c Core) QueryEngineConfig(ctx context.Context) (*config.Config, error) {
	Config, err := c.QueryByKey(ctx, config.CoreConfigKey)
	if err != nil {
		return nil, err
	}

	var engineConfig config.Config
	if err := jsoniter.UnmarshalFromString(Config.Value, &engineConfig); err != nil {
		return nil, err
	}

	return &engineConfig, nil
}
