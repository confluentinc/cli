package local

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

func NewDestroyCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "destroy",
			Short: "Delete the data and logs for the current Confluent run.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runDestroyCommand
	return c.Command
}

func (c *LocalCommand) runDestroyCommand(command *cobra.Command, _ []string) error {
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
