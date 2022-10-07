package olivetv

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-olive/olive/foundation/olivetv/util"
)

func init() {
	registerSite("lang", &lang{})
}

type lang struct {
	base
}

func (this *lang) Name() string {
	return "浪LIVE"
}

func (this *lang) Snap(tv *TV) error {
	tv.Info = &Info{
		Timestamp: time.Now().Unix(),
	}
	return this.set(tv)
}

func (this *lang) set(tv *TV) (err error) {
	roomURL := fmt.Sprintf("https://www.lang.live/room/%s", tv.RoomID)
	roomContent, err := util.GetURLContent(roomURL)
	if err != nil {
		return err
	}
	roomContent = strings.ReplaceAll(roomContent, "\\", "")
	tv.streamURL, err = util.Match(`"liveurl":"([^"]+)"`, roomContent)
	if err != nil {
		return err
	}
	title, _ := util.Match(`<head><title>([^<]+)</title>`, roomContent)
	tv.roomName = strings.Split(title, " - "+tv.RoomID)[0]
	if tv.streamerName == "" {
		tv.streamerName = tv.roomName
	}
	tv.roomOn = true
	return nil
}
