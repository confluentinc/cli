package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

func NewVersionCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the Confluent Platform version.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runVersionCommand
	return c.Command
}

func (c *Command) runVersionCommand(command *cobra.Command, _ []string) error {
	isCP, err := c.ch.IsConfluentPlatform()
	if err != nil {
		return err
	}

	flavor := "Confluent Community Software"
	if isCP {
		flavor = "Confluent Platform"
	}

	version, err := c.ch.GetVersion(flavor)
	if err != nil {
		return err
	}

	cmd.Printf(command, "%s: %s\n", flavor, version)
	return nil
}
