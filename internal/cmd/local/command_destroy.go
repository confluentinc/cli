package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/local"
)

func NewDestroyCommand(prerunner cmd.PreRunner) *cobra.Command {
	destroyCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "destroy",
			Short: "Delete the data and logs for the current Confluent run.",
			Args:  cobra.NoArgs,
			RunE:  runDestroyCommand,
		}, prerunner)

	return destroyCommand.Command
}

func runDestroyCommand(command *cobra.Command, _ []string) error {
	if err := runServicesStopCommand(command, []string{}); err != nil {
		return err
	}

	cc := local.NewConfluentCurrentManager()

	dir, err := cc.GetCurrentDir()
	if err != nil {
		return err
	}

	command.Printf("Deleting: %s\n", dir)
	if err := cc.RemoveCurrentDir(); err != nil {
		return err
	}

	return cc.RemoveTrackingFile()
}
