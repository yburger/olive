// Package olivetv provides support for retrieving stream urls and other streamers' details.
package olivetv

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

const (
	EmptyRoomName     = ""
	EmptyStreamerName = ""
)

var (
	_ ITV = (*TV)(nil)

	ErrNotSupported = errors.New("streamer not supported")
	ErrSiteInvalid  = errors.New("site invalid")
)

type ITV interface {
	Snap() error
	StreamURL() (string, bool)
	RoomName() (string, bool)
	StreamerName() (string, bool)
	SiteName() string
}

type TV struct {
	SiteID string
	RoomID string

	cookie string

	*Info
}

func New(siteID, roomID string, opts ...Option) (*TV, error) {
	_, valid := Sniff(siteID)
	if !valid {
		return nil, ErrNotSupported
	}

	t := &TV{
		SiteID: siteID,
		RoomID: roomID,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t, nil
}

func NewWithURL(roomURL string, opts ...Option) (*TV, error) {
	u := RoomURL(roomURL)
	t, err := u.Stream()
	if err != nil {
		err = fmt.Errorf("%+v (err msg = %s)", ErrNotSupported, err.Error())
		return nil, err
	}

	for _, opt := range opts {
		opt(t)
	}
	return t, nil
}

type Option func(*TV) error

func SetCookie(cookie string) Option {
	return func(t *TV) error {
		t.cookie = cookie
		return nil
	}
}

type Info struct {
	Timestamp int64

	streamURL    string
	roomOn       bool
	roomName     string
	streamerName string
}

// Snap takes the latest snapshot of the streamer info that could be retrieved individually.
func (tv *TV) Snap() error {
	if tv == nil {
		return errors.New("tv is nil")
	}
	site, ok := Sniff(tv.SiteID)
	if !ok {
		return fmt.Errorf("site(ID = %s) not supported", tv.SiteID)
	}
	return site.Snap(tv)
}

// SnapWithCookie takes the latest snapshot of the streamer info that could be retrieved individually with the cookie passed in.
func (tv *TV) SnapWithCookie(cookie string) error {
	if tv == nil {
		return errors.New("tv is nil")
	}
	tv.cookie = cookie
	site, ok := Sniff(tv.SiteID)
	if !ok {
		return fmt.Errorf("site(ID = %s) not supported", tv.SiteID)
	}
	return site.Snap(tv)
}

func (tv *TV) SiteName() string {
	if tv == nil {
		return ""
	}
	site, ok := Sniff(tv.SiteID)
	if !ok {
		return ""
	}
	return site.Name()
}

func (tv *TV) StreamURL() (string, bool) {
	if tv == nil || tv.Info == nil {
		return "", false
	}
	return tv.streamURL, tv.roomOn
}

func (tv *TV) RoomName() (string, bool) {
	if tv == nil || tv.Info == nil {
		return "", false
	}
	return tv.roomName, tv.roomName != EmptyRoomName
}

func (tv *TV) StreamerName() (string, bool) {
	if tv == nil || tv.Info == nil {
		return "", false
	}
	return tv.streamerName, tv.streamerName != EmptyStreamerName
}

func (tv *TV) String() string {
	sb := &strings.Builder{}
	sb.WriteString("Powered by go-olive/tv\n")
	sb.WriteString(format("SiteID", tv.SiteID))
	sb.WriteString(format("SiteName", tv.SiteName()))
	sb.WriteString(format("RoomID", tv.RoomID))
	if roomName, ok := tv.RoomName(); ok {
		sb.WriteString(format("RoomName", roomName))
	}
	if streamerName, ok := tv.StreamerName(); ok {
		sb.WriteString(format("Streamer", streamerName))
	}
	if streamURL, ok := tv.StreamURL(); ok {
		sb.WriteString(format("StreamUrl", streamURL))
	}
	return sb.String()
}

func format(k, v string) string {
	return fmt.Sprintf("  %-12s%-s\n", k, v)
}

type RoomURL string

func (this RoomURL) SiteID() string {
	u, err := url.Parse(string(this))
	if err != nil {
		return ""
	}
	eTLDPO, err := publicsuffix.EffectiveTLDPlusOne(u.Hostname())
	if err != nil {
		return ""
	}
	siteID := strings.Split(eTLDPO, ".")[0]
	return siteID
}

func (this RoomURL) Stream() (*TV, error) {
	site, ok := Sniff(this.SiteID())
	if !ok {
		return nil, ErrSiteInvalid
	}
	return site.Permit(this)
}
