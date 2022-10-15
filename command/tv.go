package command

import (
	"errors"
	"fmt"

	"github.com/go-olive/olive/foundation/olivetv"
	"github.com/spf13/cobra"
)

var _ cmder = (*tvCmd)(nil)

type tvCmd struct {
	cookie string
	url    string
	roomID string
	siteID string

	*baseBuilderCmd
}

func (b *commandsBuilder) newTVCmd() *tvCmd {
	cc := &tvCmd{}
	cmd := &cobra.Command{
		Use:   "tv",
		Short: "TV is a cli utility which gets stream url.",
		Long:  `TV is a cli utility which gets stream url.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run()
		},
	}
	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	cmd.Flags().StringVarP(&cc.cookie, "cookie", "c", "", "site cookie")
	cmd.Flags().StringVarP(&cc.url, "url", "u", "", "room url")
	cmd.Flags().StringVarP(&cc.roomID, "rid", "r", "", "room ID")
	cmd.Flags().StringVarP(&cc.siteID, "sid", "s", "", "site ID")

	return cc
}

func (c *tvCmd) run() error {
	switch {
	case c.url != "":
		t, err := olivetv.NewWithURL(c.url, olivetv.SetCookie(c.cookie))
		if err != nil {
			return err
		}
		t.Snap()
		fmt.Println(t)

	case c.roomID != "" && c.siteID != "":
		t, err := olivetv.New(c.siteID, c.roomID, olivetv.SetCookie(c.cookie))
		if err != nil {
			return err
		}
		t.Snap()
		fmt.Println(t)

	default:
		return errors.New("need to specify [roomd id and site id] or [room url]")
	}

	return nil
}
