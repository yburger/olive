package engine

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/go-dora/filenamify"
	"github.com/go-olive/olive/src/config"
	"github.com/go-olive/olive/src/dispatcher"
	"github.com/go-olive/olive/src/enum"
	l "github.com/go-olive/olive/src/log"
	"github.com/go-olive/olive/src/parser"
	"github.com/go-olive/olive/src/util"

	"github.com/go-olive/tv"
)

var (
	defaultOutTmpl = template.Must(template.New("filename").Funcs(util.NameFuncMap).
		Parse(`[{{ .StreamerName }}][{{ .RoomName }}][{{ now | date "2006-01-02 15-04-05"}}].flv`))
)

type ID string

type Show interface {
	GetID() ID
	GetPlatform() string
	GetRoomID() string
	GetStreamerName() string
	GetOut() string
	GetOutTmpl() string
	GetSaveDir() string
	GetPostCmds() []*exec.Cmd
	GetSplitRule() *config.SplitRule
	SatisfySplitRule(time.Time, string) bool

	AddMonitor() error
	RemoveMonitor() error
	AddRecorder() error
	RemoveRecorder() error
	RestartRecorder()

	NewParser() (parser.Parser, error)

	tv.ITv
}

type show struct {
	ID        ID
	Platform  string
	RoomID    string
	Streamer  string
	OutTmpl   string
	Parser    string
	SaveDir   string
	PostCmds  []*exec.Cmd
	SplitRule *config.SplitRule
	enum.ShowTaskStatusID
	stop chan struct{}

	*tv.Tv
}

type ShowOption func(*show)

func WithStreamerName(name string) ShowOption {
	return func(s *show) {
		s.Streamer = name
	}
}

func WithOutTmpl(tmpl string) ShowOption {
	return func(s *show) {
		s.OutTmpl = tmpl
	}
}

func WithParser(parser string) ShowOption {
	return func(s *show) {
		s.Parser = parser
	}
}

func WithSaveDir(saveDir string) ShowOption {
	return func(s *show) {
		s.SaveDir = saveDir
	}
}

func WithPostCmds(postCmds []*exec.Cmd) ShowOption {
	return func(s *show) {
		s.PostCmds = postCmds
	}
}

func WithSplitRule(rule *config.SplitRule) ShowOption {
	return func(s *show) {
		s.SplitRule = rule
	}
}

func NewShow(platformType, roomID string, opts ...ShowOption) (Show, error) {
	var cookie string
	switch platformType {
	case "douyin":
		cookie = config.APP.PlatformConfig.DouyinCookie
	case "kuaishou":
		cookie = config.APP.PlatformConfig.KuaishouCookie
	}

	t, err := tv.New(platformType, roomID, tv.SetCookie(cookie))
	if err != nil {
		return nil, fmt.Errorf("Show init failed! err msg: %s", err.Error())
	}

	s := &show{
		Platform: platformType,
		RoomID:   roomID,

		stop: make(chan struct{}),

		Tv: t,
	}
	for _, opt := range opts {
		opt(s)
	}

	s.ID = s.genID()
	return s, nil
}

func (s *show) GetID() ID {
	return s.ID
}

func (s *show) GetRoomID() string {
	return s.RoomID
}

func (s *show) GetStreamerName() string {
	if s.Streamer == "" {
		s.Streamer, _ = s.StreamerName()
	}
	return s.Streamer
}

func (s *show) GetPlatform() string {
	return s.Platform
}

func (s *show) GetOutTmpl() string {
	return s.OutTmpl
}

// GetOut generate output filename
func (s *show) GetOut() (out string) {
	roomName, _ := s.RoomName()

	// generate template info
	info := &struct {
		StreamerName string
		RoomName     string
		SiteName     string
	}{
		StreamerName: s.GetStreamerName(),
		RoomName:     roomName,
		SiteName:     s.SiteName(),
	}

	// generate file name
	tmpl, err := template.New("user_defined_filename").Funcs(util.NameFuncMap).Parse(s.GetOutTmpl())
	if err != nil {
		l.Logger.Error(err)
		tmpl = defaultOutTmpl
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, info); err != nil {
		l.Logger.Error(err)
		const format = "2006-01-02 15-04-05"
		out = fmt.Sprintf("[%s][%s][%s].flv", info.StreamerName, roomName, time.Now().Format(format))
	} else {
		out = buf.String()
	}

	out = filenamify.FilenamifyMustCompile(out)
	return
}

func (s *show) GetParser() string {
	return s.Parser
}

// GetSaveDir generate save dir
func (s *show) GetSaveDir() string {
	defaultSaveDir, _ := os.Getwd()
	defaultSaveDir = strings.TrimSpace(s.SaveDir)

	roomName, _ := s.RoomName()
	// generate template info
	info := &struct {
		StreamerName string
		RoomName     string
		SiteName     string
	}{
		StreamerName: s.GetStreamerName(),
		RoomName:     roomName,
		SiteName:     s.SiteName(),
	}

	tmpl, err := template.New("user_defined_savedir_tmpl").Funcs(util.NameFuncMap).Parse(s.SaveDir)
	if err != nil {
		l.Logger.Error(err)
		return defaultSaveDir
	}
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, info); err != nil {
		l.Logger.Error(err)
		return defaultSaveDir
	}

	return buf.String()
}

func (s *show) GetPostCmds() []*exec.Cmd {
	return s.PostCmds
}

func (s *show) GetSplitRule() *config.SplitRule {
	return s.SplitRule
}

func (s *show) SatisfySplitRule(startTime time.Time, out string) bool {
	return s.SplitRule.Satisfy(startTime, out)
}

func (s *show) genID() ID {
	h := md5.New()
	b := []byte(fmt.Sprintf("%s%s%d", s.Platform, s.RoomID, time.Now().UnixNano()))
	h.Write(b)
	return ID(hex.EncodeToString(h.Sum(nil)))
}

func (s *show) NewParser() (parser.Parser, error) {
	v, ok := parser.SharedManager.Parser(s.GetParser())
	if !ok {
		return nil, errors.New("parser not exist")
	}
	return v.New(), nil
}

func (s *show) AddMonitor() error {
	e := dispatcher.NewEvent(enum.EventType.AddMonitor, s)
	return dispatcher.SharedManager.Dispatch(e)
}

func (s *show) RemoveMonitor() error {
	e := dispatcher.NewEvent(enum.EventType.RemoveMonitor, s)
	return dispatcher.SharedManager.Dispatch(e)
}

func (s *show) AddRecorder() error {
	e := dispatcher.NewEvent(enum.EventType.AddRecorder, s)
	return dispatcher.SharedManager.Dispatch(e)
}

func (s *show) RemoveRecorder() error {
	e := dispatcher.NewEvent(enum.EventType.RemoveRecorder, s)
	return dispatcher.SharedManager.Dispatch(e)
}

func (s *show) RestartRecorder() {
	s.RemoveRecorder()
	s.AddRecorder()
}

func (s *show) Stop() {
	dispatcher.SharedManager.Dispatch(dispatcher.NewEvent(enum.EventType.RemoveMonitor, s))
	dispatcher.SharedManager.Dispatch(dispatcher.NewEvent(enum.EventType.RemoveRecorder, s))
}
