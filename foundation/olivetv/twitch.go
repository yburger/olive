package olivetv

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-olive/olive/foundation/olivetv/util"
)

func init() {
	registerSite("twitch", &twitch{})
}

type twitch struct {
	base
}

func (this *twitch) Name() string {
	return "推趣"
}

func (this *twitch) Snap(tv *TV) error {
	tv.Info = &Info{
		Timestamp: time.Now().Unix(),
	}
	return this.set(tv)
}

func (this *twitch) set(tv *TV) error {
	roomURL := fmt.Sprintf("https://www.twitch.tv/%s", tv.RoomID)
	content, err := util.GetURLContent(roomURL)
	if err != nil {
		return err
	}

	tv.roomOn = strings.Contains(content, `"isLiveBroadcast":true`)
	if !tv.roomOn {
		return nil
	}

	tv.streamURL = roomURL
	title, err := util.Match(`"description":"([^"]+)"`, content)
	if err != nil {
		return nil
	}

	tv.roomName = title

	return nil
}
