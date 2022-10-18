package kernel

import (
	"os"
	"time"

	"github.com/go-olive/olive/engine/config"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

// Show represents an individual show.
type Show struct {
	ID           string    `json:"show_id"`
	Enable       bool      `json:"enable"`
	Platform     string    `json:"platform"`
	RoomID       string    `json:"room_id"`
	StreamerName string    `json:"streamer_name"`
	OutTmpl      string    `json:"out_tmpl"`
	Parser       string    `json:"parser"`
	SaveDir      string    `json:"save_dir"`
	PostCmds     string    `json:"post_cmds"`
	SplitRule    string    `json:"split_rule"`
	DateCreated  time.Time `json:"date_created"`
	DateUpdated  time.Time `json:"date_updated"`
}

func (s *Show) CheckAndFix(cfg *config.Config) {
	// generate an ID if not given
	if s.ID == "" {
		s.ID = uuid.NewString()
	}

	// fix parser
	if s.Parser == "" {
		switch s.Platform {
		case "youtube",
			"twitch",
			"streamlink":
			s.Parser = "streamlink"
		default:
			s.Parser = "flv"
		}
	}

	// fix SaveDir
	if s.SaveDir == "" {
		s.SaveDir = cfg.SaveDir
	}

	// fix OutTmpl
	if s.OutTmpl == "" {
		s.OutTmpl = cfg.OutTmpl
	}
}

type SplitRule struct {
	FileSize int64
	Duration string

	parsedDuration time.Duration
}

func NewSplitRule(str string) (*SplitRule, error) {
	var sr SplitRule
	if err := jsoniter.UnmarshalFromString(str, &sr); err != nil {
		return nil, err
	}
	sr.parsedDuration, _ = time.ParseDuration(sr.Duration)
	return &sr, nil
}

func (sr *SplitRule) IsValid() bool {
	if sr == nil {
		return false
	}
	if sr.parsedDuration <= 0 && sr.FileSize <= 0 {
		return false
	}
	return true
}

func (sr *SplitRule) Satisfy(startTime time.Time, out string) bool {
	if !sr.IsValid() {
		return false
	}

	if sr.parsedDuration > 0 {
		if time.Since(startTime) >= sr.parsedDuration {
			return true
		}
	}

	if sr.FileSize > 0 {
		if fi, err := os.Stat(out); err == nil {
			if fi.Size() >= sr.FileSize {
				return true
			}
		}
	}

	return false
}
