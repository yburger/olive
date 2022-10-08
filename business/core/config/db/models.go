package db

import (
	"time"
)

// Config represent the structure we need for moving data
// between the app and the database.
type Config struct {
	Key         string    `db:"key"`
	Value       string    `db:"value"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}
