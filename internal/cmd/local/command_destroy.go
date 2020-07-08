package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
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
		return errors.HandleCommon(errors.New(errors.NothingToDestroyErrorMsg), c.Command)
	}

	if err := c.runServicesStopCommand(command, []string{}); err != nil {
		return errors.HandleCommon(err, c.Command)
	}

	dir, err := c.cc.GetCurrentDir()
	if err != nil {
		return errors.HandleCommon(err, c.Command)
	}

	command.Printf(errors.DestroyDeletingMsg, dir)
	if err := c.cc.RemoveCurrentDir(); err != nil {
		return errors.HandleCommon(err, c.Command)
	}

	return c.cc.RemoveTrackingFile()
}
