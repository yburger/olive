package config

import (
	"time"
	"unsafe"

	"github.com/go-olive/olive/business/core/config/db"
)

// Config represents an individual config.
type Config struct {
	Key         string    `db:"key"`
	Value       string    `db:"value"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

// NewConfig contains information needed to create a new Config.
type NewConfig struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

// UpdateConfig defines what information may be provided to modify an existing
// Config. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type UpdateConfig struct {
	Value *string `json:"value"`
}

// =============================================================================

func toConfig(dbConfig db.Config) Config {
	c := (*Config)(unsafe.Pointer(&dbConfig))
	return *c
}
