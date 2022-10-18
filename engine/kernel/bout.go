// Package kernel is the core of olive.
package kernel

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/go-dora/filenamify"
	"github.com/go-olive/olive/engine/config"
	"github.com/go-olive/olive/engine/dispatcher"
	"github.com/go-olive/olive/engine/enum"
	l "github.com/go-olive/olive/engine/log"
	"github.com/go-olive/olive/engine/util"
	"github.com/go-olive/olive/foundation/olivetv"
	"github.com/go-olive/olive/foundation/syncmap"
	jsoniter "github.com/json-iterator/go"
)

var (
	_ config.Bout = (*bout)(nil)

	defaultOutTmpl = template.Must(template.New("filename").Funcs(util.NameFuncMap).
			Parse(`[{{ .StreamerName }}][{{ .RoomName }}][{{ now | date "2006-01-02 15-04-05"}}].flv`))
)

// bout represents one live show.
type bout struct {
	showID  string
	show    Show
	showMap *syncmap.RWMap[string, Show]
	cfg     *config.Config

	*olivetv.TV
}

func NewBout(showID string, showMap *syncmap.RWMap[string, Show], cfg *config.Config) (*bout, error) {
	showCfg, ok := showMap.Get(showID)
	if !ok {
		return nil, fmt.Errorf("show[ID = %s] config does not exist", showID)
	}
	tv, err := olivetv.New(showCfg.Platform, showCfg.RoomID)
	if err != nil {
		return nil, err
	}

	return &bout{
		TV:      tv,
		showID:  showID,
		show:    showCfg,
		showMap: showMap,
		cfg:     cfg,
	}, nil
}

func (b *bout) Refresh() {
	s, ok := b.showMap.Get(b.showID)
	if !ok {
		return
	}

	if s.Platform != b.SiteID || s.RoomID != b.RoomID {
		newTV, err := olivetv.New(s.Platform, s.RoomID)
		if err != nil {
			return
		}
		b.TV = newTV
	}

	s.CheckAndFix(b.cfg)
	b.show = s
}

func (b *bout) Snap() error {
	b.Refresh()

	switch b.TV.SiteID {
	case "douyin":
		return b.TV.SnapWithCookie(b.cfg.DouyinCookie)
	case "kuaishou":
		return b.TV.SnapWithCookie(b.cfg.KuaishouCookie)
	default:
		return b.TV.Snap()
	}
}

func (b *bout) GetID() config.ID {
	return config.ID(b.showID)
}

func (b *bout) GetRoomID() string {
	b.Refresh()

	return b.RoomID
}

func (b *bout) GetStreamerName() string {
	b.Snap()

	streamerName := b.show.StreamerName
	if streamerName == "" {
		streamerName, _ = b.StreamerName()
	}
	return streamerName
}

func (b *bout) GetPlatform() string {
	b.Refresh()

	return b.SiteID
}

func (b *bout) GetOutTmpl() string {
	b.Refresh()

	return b.show.OutTmpl
}

// GetOutFilename generate output filename
func (b *bout) GetOutFilename() (out string) {
	b.Refresh()

	roomName, _ := b.RoomName()

	// generate template info
	info := &struct {
		StreamerName string
		RoomName     string
		SiteName     string
	}{
		StreamerName: b.GetStreamerName(),
		RoomName:     roomName,
		SiteName:     b.SiteName(),
	}

	// generate file name
	tmpl, err := template.New("user_defined_filename").Funcs(util.NameFuncMap).Parse(b.show.OutTmpl)
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

func (b *bout) GetParser() string {
	b.Refresh()

	return b.show.Parser
}

// GetSaveDir generate save dir
func (b *bout) GetSaveDir() string {
	b.Refresh()

	defaultSaveDir := strings.TrimSpace(b.show.SaveDir)

	roomName, _ := b.RoomName()
	// generate template info
	info := &struct {
		StreamerName string
		RoomName     string
		SiteName     string
	}{
		StreamerName: b.GetStreamerName(),
		RoomName:     roomName,
		SiteName:     b.SiteName(),
	}

	tmpl, err := template.New("user_defined_savedir_tmpl").Funcs(util.NameFuncMap).Parse(b.show.SaveDir)
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

func (b *bout) GetPostCmds() []*exec.Cmd {
	b.Refresh()

	s, ok := b.showMap.Get(b.showID)
	if !ok {
		return nil
	}
	var cmds []*exec.Cmd
	if err := jsoniter.UnmarshalFromString(s.PostCmds, &cmds); err != nil {
		return nil
	}
	return cmds
}

func (b *bout) SatisfySplitRule(startTime time.Time, out string) bool {
	b.Refresh()

	sr, err := NewSplitRule(b.show.SplitRule)
	if err != nil {
		return false
	}
	return sr.Satisfy(startTime, out)
}

func (b *bout) AddMonitor() error {
	b.Refresh()

	e := dispatcher.NewEvent(enum.EventType.AddMonitor, b)
	return dispatcher.SharedManager.Dispatch(e)
}

func (b *bout) RemoveMonitor() error {
	e := dispatcher.NewEvent(enum.EventType.RemoveMonitor, b)
	return dispatcher.SharedManager.Dispatch(e)
}

func (b *bout) AddRecorder() error {
	b.Refresh()

	e := dispatcher.NewEvent(enum.EventType.AddRecorder, b)
	return dispatcher.SharedManager.Dispatch(e)
}

func (b *bout) RemoveRecorder() error {
	e := dispatcher.NewEvent(enum.EventType.RemoveRecorder, b)
	return dispatcher.SharedManager.Dispatch(e)
}

func (b *bout) RestartRecorder() {
	b.RemoveRecorder()
	b.AddRecorder()
}
