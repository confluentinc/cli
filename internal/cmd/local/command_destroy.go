package local

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func NewDestroyCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "destroy",
			Args:  cobra.NoArgs,
			Short: "Delete the data and logs for the current Confluent run.",
			Example: examples.BuildExampleString(
				examples.Example{
					Desc: "If you run the ``confluent local destroy`` command, your output will confirm that every service is stopped and the deleted filesystem path is printed:",
					Code: "confluent local destroy",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runDestroyCommand
	return c.Command
}

func (c *Command) runDestroyCommand(command *cobra.Command, _ []string) error {
	if !c.cc.HasTrackingFile() {
		return fmt.Errorf("nothing to destroy")
	}

	if err := c.runServicesStopCommand(command, []string{}); err != nil {
		return err
	}

	dir, err := c.cc.GetCurrentDir()
	if err != nil {
		return err
	}

	command.Printf("Deleting: %s\n", dir)
	if err := c.cc.RemoveCurrentDir(); err != nil {
		return err
	}

	return c.cc.RemoveTrackingFile()
}
