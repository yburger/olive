package olivetv

import (
	"os/exec"
	"time"
)

func init() {
	registerSite("streamlink", &streamlink{})
}

type streamlink struct {
	base
}

func (this *streamlink) Name() string {
	return "streamlink"
}

func (this *streamlink) Snap(tv *Tv) error {
	tv.Info = &Info{
		Timestamp: time.Now().Unix(),
	}
	return this.set(tv)
}

func (this *streamlink) set(tv *Tv) error {
	cmd := exec.Command(
		"streamlink",
		tv.RoomID,
	)
	if err := cmd.Run(); err != nil {
		return nil
	}

	tv.roomOn = true
	tv.streamUrl = tv.RoomID

	return nil
}
