package olivetv

import (
	"time"
)

func init() {
	registerSite("tmpl", &tmpl{})
}

type tmpl struct {
	base
}

func (this *tmpl) Name() string {
	return "tmpl"
}

func (this *tmpl) Snap(tv *TV) error {
	tv.Info = &Info{
		Timestamp: time.Now().Unix(),
	}
	return this.set(tv)
}

func (this *tmpl) set(tv *TV) error {
	tv.roomName = "tmpl room name"
	tv.streamerName = "tmpl streamer name"
	tv.roomOn = true
	tv.streamURL = "tmpl stream url"
	return nil
}
