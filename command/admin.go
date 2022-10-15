package command

import (
	"github.com/go-olive/olive/app/tooling/olive-admin/commands"
	"github.com/go-olive/olive/business/sys/database"
	"github.com/spf13/cobra"
)

var _ cmder = (*adminCmd)(nil)

type adminCmd struct {
	DB

	*baseBuilderCmd
}

func (b *commandsBuilder) newAdminCmd() *adminCmd {
	cc := &adminCmd{}
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Admin is a cli utility for database.",
		Long:  `Admin is a cli utility for database.`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "migrate",
			Short: "Migrating database.",
			Long:  `Migrating database.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				return cc.migrate()
			},
		},
		&cobra.Command{
			Use:   "seed",
			Short: "Seeding database.",
			Long:  `Seeding database.`,
			RunE: func(cmd *cobra.Command, args []string) error {
				return cc.seed()
			},
		},
	)

	cmd.PersistentFlags().StringVar(&cc.User, "db-user", "postgres", "")
	cmd.PersistentFlags().StringVar(&cc.Password, "db-password", "postgres", "")
	cmd.PersistentFlags().StringVar(&cc.Host, "db-host", "localhost", "")
	cmd.PersistentFlags().StringVar(&cc.Name, "db-name", "postgres", "")
	cmd.PersistentFlags().IntVar(&cc.MaxIdleConns, "db-max-idle-conns", 0, "")
	cmd.PersistentFlags().IntVar(&cc.MaxOpenConns, "db-max-open-conns", 0, "")
	cmd.PersistentFlags().BoolVar(&cc.DisableTLS, "db-disable-tls", true, "")

	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	return cc
}

func (c *adminCmd) migrate() error {
	return commands.Migrate(database.Config(c.DB))
}

func (c *adminCmd) seed() error {
	return commands.Seed(database.Config(c.DB))
}
