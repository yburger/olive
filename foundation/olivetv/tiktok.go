package olivetv

import (
	"log"
	"strings"
	"time"

	"github.com/Davincible/gotiktoklive"
)

func init() {
	registerSite("tiktok", &tiktok{})
}

type tiktok struct {
	base
}

func (this *tiktok) Name() string {
	return "tiktok"
}

func (this *tiktok) Snap(tv *TV) error {
	tv.Info = &Info{
		Timestamp: time.Now().Unix(),
	}
	return this.set(tv)
}

func (this *tiktok) set(tv *TV) error {
	defer func() {
		if err := recover(); err != nil {
			log.Println("tiktok panic: ", err)
		}
	}()

	tiktok := gotiktoklive.NewTikTok()
	info, err := tiktok.GetRoomInfo(tv.RoomID)
	if err != nil {
		return err
	}

	candi := []string{
		info.StreamURL.FlvPullURL.FullHd1,
		info.StreamURL.FlvPullURL.Hd1,
		info.StreamURL.FlvPullURL.Sd1,
		info.StreamURL.FlvPullURL.Sd2,
	}
	var streamURL string
	for _, v := range candi {
		if v != "" {
			streamURL = v
			break
		}
	}

	if streamURL != "" {
		tv.roomName = info.Owner.Nickname + " is LIVE now"
		tv.streamerName = info.Owner.Nickname
		tv.roomOn = true
		tv.streamURL = streamURL
	}

	return nil
}

// Permit parse the stream url to get streamer info.
// eg. https://www.tiktok.com/@maki_1414
func (this *tiktok) Permit(roomURL RoomURL) (*TV, error) {
	tv, error := this.base.Permit(roomURL)
	if error != nil {
		return nil, error
	}
	tv.RoomID = strings.TrimPrefix(tv.RoomID, "@")
	return tv, nil
}
