package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

func NewCommand(prerunner cmd.PreRunner) *cobra.Command {
	localCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "local-v2 [command]",
			Short: "Manage a local Confluent Platform development environment.",
		}, prerunner,
	)

	localCommand.AddCommand(NewCurrentCommand(prerunner))
	localCommand.AddCommand(NewDemoCommand(prerunner))
	localCommand.AddCommand(NewDestroyCommand(prerunner))
	localCommand.AddCommand(NewServicesCommand(prerunner))
	localCommand.AddCommand(NewVersionCommand(prerunner))

	return localCommand.Command
}
