package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func NewCurrentCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "current",
			Short: "Get the path of the current Confluent run.",
			Long:  "Print the filesystem path of the data and logs of the services managed by the current `confluent local` command. If such a path does not exist, it will be created.",
			Args:  cobra.NoArgs,
			Example: examples.BuildExampleString(
				examples.Example{
					Text: "In Linux, running `confluent local current` should resemble the following:",
					Code: "/tmp/confluent.SpBP4fQi",
				},
				examples.Example{
					Text: "In macOS, running `confluent local current` should resemble the following:",
					Code: "/var/folders/cs/1rndf6593qb3kb6r89h50vgr0000gp/T/confluent.000000",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runCurrentCommand
	return c.Command
}

func (c *command) runCurrentCommand(_ *cobra.Command, _ []string) error {
	dir, err := c.cc.GetCurrentDir()
	if err != nil {
		return err
	}

	output.Println(c.Config.EnableColor, dir)
	return nil
}
