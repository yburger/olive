package show

import (
	"time"
	"unsafe"

	"github.com/go-olive/olive/business/core/show/db"
	"github.com/go-olive/olive/engine/kernel"
)

// Show represents an individual show.
type Show = kernel.Show

type SplitRule struct {
	FileSize int64
	Duration time.Duration
}

// NewShow contains information needed to create a new Show.
type NewShow struct {
	Enable       bool   `json:"enable"`
	Platform     string `json:"platform" validate:"required"`
	RoomID       string `json:"room_id" validate:"required"`
	StreamerName string `json:"streamer_name"`
	OutTmpl      string `json:"out_tmpl"`
	Parser       string `json:"parser"`
	SaveDir      string `json:"save_dir"`
	PostCmds     string `json:"post_cmds"`
	SplitRule    string `json:"split_rule"`
}

// UpdateShow defines what information may be provided to modify an existing
// Show. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type UpdateShow struct {
	Enable       *bool   `json:"enable"`
	Platform     *string `json:"platform"`
	RoomID       *string `json:"room_id"`
	StreamerName *string `json:"streamer_name"`
	OutTmpl      *string `json:"out_tmpl"`
	Parser       *string `json:"parser"`
	SaveDir      *string `json:"save_dir"`
	PostCmds     *string `json:"post_cmds"`
	SplitRule    *string `json:"split_rule"`
}

// =============================================================================

func toShow(dbShow db.Show) Show {
	s := (*Show)(unsafe.Pointer(&dbShow))
	return *s
}

func toShowSlice(dbShows []db.Show) []Show {
	Shows := make([]Show, len(dbShows))
	for i, dbShow := range dbShows {
		Shows[i] = toShow(dbShow)
	}
	return Shows
}
