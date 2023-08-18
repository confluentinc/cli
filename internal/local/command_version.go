package local

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
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

func (c *Command) runVersionCommand(_ *cobra.Command, _ []string) error {
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

	output.Printf("%s: %s\n", flavor, version)
	return nil
}
