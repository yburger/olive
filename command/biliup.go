package command

import (
	"github.com/go-olive/olive/foundation/biliup"
	"github.com/spf13/cobra"
)

var _ cmder = (*biliupCmd)(nil)

type biliupCmd struct {
	*baseBuilderCmd
}

func (b *commandsBuilder) newBiliupCmd() *biliupCmd {
	cc := &biliupCmd{}
	cmd := &cobra.Command{
		Use:   "biliup",
		Short: "Biliup is a cli utility which generates bilibli cookies.",
		Long:  `Biliup is a cli utility which generates bilibli cookies.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run()
		},
	}
	cc.baseBuilderCmd = b.newBuilderCmd(cmd)
	return cc
}

func (c *biliupCmd) run() error {
	return biliup.Login()
}
