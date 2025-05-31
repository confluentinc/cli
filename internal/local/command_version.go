package local

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func NewVersionCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the Confluent Platform version.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runVersionCommand
	return c.Command
}

func (c *command) runVersionCommand(_ *cobra.Command, _ []string) error {
	isCP, err := c.ch.IsConfluentPlatform()
	if err != nil {
		return err
	}

	flavor := "Confluent Community Software"
	if isCP {
		flavor = "Confluent Platform"
	}

	// The bool value doesn't matter when we pass in "Confluent Platform" or "Confluent Community Software", so we just choose "false"
	version, err := c.ch.GetVersion(flavor, false)
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "%s: %s\n", flavor, version)
	return nil
}
