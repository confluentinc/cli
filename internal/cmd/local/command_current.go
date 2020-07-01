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
			Long:  "Print the filesystem path of the data and logs of the services managed by the current \"confluent local\" command. If such a path does not exist, it will be created.",
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
