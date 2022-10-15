// Package command defines and implements command-line commands and flags
// used by olive.
package command

import "github.com/spf13/cobra"

type baseBuilderCmd struct {
	*baseCmd
	*commandsBuilder
}

func (b *commandsBuilder) newBuilderCmd(cmd *cobra.Command) *baseBuilderCmd {
	bcmd := &baseBuilderCmd{commandsBuilder: b, baseCmd: &baseCmd{cmd: cmd}}
	return bcmd
}

type commandsBuilder struct {
	commands []cmder
}

func newCommandsBuilder() *commandsBuilder {
	return &commandsBuilder{}
}

func (b *commandsBuilder) addCommands(commands ...cmder) *commandsBuilder {
	b.commands = append(b.commands, commands...)
	return b
}

func (b *commandsBuilder) addAll() *commandsBuilder {
	b.addCommands(
		newVersionCmd(),
		b.newTVCmd(),
		b.newRunCmd(),
	)

	return b
}

func (b *commandsBuilder) build() *oliveCmd {
	h := b.newOliveCmd()
	addCommands(h.getCommand(), b.commands...)
	return h
}

func addCommands(root *cobra.Command, commands ...cmder) {
	for _, command := range commands {
		cmd := command.getCommand()
		if cmd == nil {
			continue
		}
		root.AddCommand(cmd)
	}
}

type baseCmd struct {
	cmd *cobra.Command
}

func newBaseCmd(cmd *cobra.Command) *baseCmd {
	return &baseCmd{cmd: cmd}
}

func (c *baseCmd) getCommand() *cobra.Command {
	return c.cmd
}

type cmder interface {
	getCommand() *cobra.Command
}

// The Response value from Execute.
type Response struct {
	// Err is set when the command failed to execute.
	Err error

	// The command that was executed.
	Cmd *cobra.Command
}

func Execute(args []string) Response {
	oliveCmd := newCommandsBuilder().addAll().build()
	cmd := oliveCmd.getCommand()

	c, err := cmd.ExecuteC()
	resp := Response{
		Err: err,
		Cmd: c,
	}

	return resp
}
