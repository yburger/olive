package db

import (
	"time"
)

// Show represent the structure we need for moving data
// between the app and the database.
type Show struct {
	ID           string    `db:"show_id"`
	Enable       bool      `db:"enable"`
	Platform     string    `db:"platform"`
	RoomID       string    `db:"room_id"`
	StreamerName string    `db:"streamer_name"`
	OutTmpl      string    `db:"out_tmpl"`
	Parser       string    `db:"parser"`
	SaveDir      string    `db:"save_dir"`
	PostCmds     string    `db:"post_cmds"`
	SplitRule    string    `db:"split_rule"`
	DateCreated  time.Time `db:"date_created"`
	DateUpdated  time.Time `db:"date_updated"`
}
