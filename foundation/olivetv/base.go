package olivetv

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

type base struct{}

func (b *base) Name() string {
	return "undefined"
}

func (b *base) Snap(tv *TV) error {
	return fmt.Errorf("site(ID = %s) Snap Method not implemented", tv.SiteID)
}

func (b *base) Permit(roomURL RoomURL) (*TV, error) {
	u, err := url.Parse(string(roomURL))
	if err != nil {
		return nil, err
	}
	eTLDPO, err := publicsuffix.EffectiveTLDPlusOne(u.Hostname())
	if err != nil {
		return nil, err
	}
	siteID := strings.Split(eTLDPO, ".")[0]
	base := strings.TrimPrefix(u.Path, "/")
	roomIDTmp := strings.Split(base, "/")
	roomID := roomIDTmp[len(roomIDTmp)-1]
	return &TV{
		SiteID: siteID,
		RoomID: roomID,
	}, nil
}
