package local

import (
	"github.com/confluentinc/cli/internal/pkg/local"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
)

func NewCurrentCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	currentCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "current",
			Short: "Get the path of the data and logs for the current Confluent run.",
			Args:  cobra.NoArgs,
			RunE:  runCurrentCommand,
		},
		cfg, prerunner)

	return currentCommand.Command
}

func runCurrentCommand(command *cobra.Command, _ []string) error {
	cc := local.NewConfluentCurrentManager()

	dir, err := cc.GetCurrentDir()
	if err != nil {
		return err
	}

	command.Println(dir)
	return nil
}

