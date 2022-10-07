// Package config stores config type.
package config

import (
	"os/exec"
	"time"
)

type Config struct {
	// portal
	PortalUsername string `conf:"default:olive"`
	PortalPassword string `conf:"default:olive"`

	// core
	SaveDir                  string `conf:"default:/"`
	OutTmpl                  string `conf:"default:[{{ .StreamerName }}][{{ .RoomName }}][{{ now | date \"2006-01-02 15-04-05\"}}].flv"`
	LogLevel                 uint32 `conf:"default:5"`
	SnapRestSeconds          uint   `conf:"default:15"`
	SplitRestSeconds         uint   `conf:"default:60"`
	CommanderPoolSize        uint   `conf:"default:1"`
	ParserMonitorRestSeconds uint   `conf:"default:300"`

	// tv
	DouyinCookie   string `conf:"default:__ac_nonce=06245c89100e7ab2dd536; __ac_signature=_02B4Z6wo00f01LjBMSAAAIDBwA.aJ.c4z1C44TWAAEx696;"`
	KuaishouCookie string `conf:"default:did=web_d86297aa2f579589b8abc2594b0ea985"`
}

type ID string

type Bout interface {
	// show settings
	GetID() ID
	GetPlatform() string
	GetRoomID() string
	GetStreamerName() string
	GetOutFilename() string
	GetOutTmpl() string
	GetSaveDir() string
	GetParser() string
	GetPostCmds() []*exec.Cmd
	SatisfySplitRule(time.Time, string) bool

	// show events
	AddMonitor() error
	RemoveMonitor() error
	AddRecorder() error
	RemoveRecorder() error
	RestartRecorder()

	// tv
	Snap() error
	StreamURL() (string, bool)
	RoomName() (string, bool)
	StreamerName() (string, bool)
	SiteName() string
}
