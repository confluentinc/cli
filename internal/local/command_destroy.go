package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func NewDestroyCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "destroy",
			Short: "Delete the data and logs for the current Confluent run.",
			Long:  "Delete an existing Confluent Platform run. All running services are stopped and the data and the log files of all services are deleted.",
			Args:  cobra.NoArgs,
			Example: examples.BuildExampleString(
				examples.Example{
					Text: "If you run the `confluent local destroy` command, your output will confirm that every service is stopped and the deleted filesystem path is printed:",
					Code: "confluent local destroy",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runDestroyCommand
	return c.Command
}

func (c *command) runDestroyCommand(cmd *cobra.Command, _ []string) error {
	if !c.cc.HasTrackingFile() {
		return errors.New("nothing to destroy")
	}

	if err := c.runServicesStopCommand(cmd, []string{}); err != nil {
		return err
	}

	dir, err := c.cc.GetCurrentDir()
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Deleting: %s\n", dir)
	if err := c.cc.RemoveCurrentDir(); err != nil {
		return err
	}

	return c.cc.RemoveTrackingFile()
}
