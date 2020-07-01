package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

func NewCurrentCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "current",
			Short: "Get the path of the current Confluent run.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runCurrentCommand
	return c.Command
}

func (c *Command) runCurrentCommand(command *cobra.Command, _ []string) error {
	dir, err := c.cc.GetCurrentDir()
	if err != nil {
		return err
	}

	command.Println(dir)
	return nil
}
