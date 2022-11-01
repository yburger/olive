package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var build = "v0.5.1-src"

var _ cmder = (*versionCmd)(nil)

type versionCmd struct {
	*baseCmd
}

func newVersionCmd() *versionCmd {
	return &versionCmd{
		newBaseCmd(&cobra.Command{
			Use:   "version",
			Short: "Print the version number of olive",
			Long:  `All software has versions. This is olive's.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				printOliveVersion()
				return nil
			},
		}),
	}
}

func printOliveVersion() {
	fmt.Println(build)
}
