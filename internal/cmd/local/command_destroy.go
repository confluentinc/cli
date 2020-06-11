package local

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
)

func NewDestroyCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	destroyCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "destroy",
			Short: "Delete the data and logs for the current Confluent run.",
			Args:  cobra.NoArgs,
			RunE:  runDestroyCommand,
		},
		cfg, prerunner)

	return destroyCommand.Command
}

func runDestroyCommand(command *cobra.Command, _ []string) error {
	if err := runServicesStopCommand(command, []string{}); err != nil {
		return err
	}

	confluentCurrent, err := getConfluentCurrent()
	if err != nil {
		return err
	}

	command.Printf("Deleting: %s\n", confluentCurrent)
	if err := os.RemoveAll(confluentCurrent); err != nil {
		return err
	}

	root := os.Getenv("CONFLUENT_CURRENT")
	if root == "" {
		root = os.TempDir()
	}

	trackingFile := filepath.Join(root, "confluent.current")
	return os.Remove(trackingFile)
}
