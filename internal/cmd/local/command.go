package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

func NewCommand(prerunner cmd.PreRunner) *cobra.Command {
	localCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:              "local-v2 [command]",
			Short:            "Manage a local Confluent Platform development environment.",
		}, prerunner,
	)
	localCommand.AddCommand(NewConnectorsCommand(prerunner))
	localCommand.AddCommand(NewCurrentCommand(prerunner))
	// TODO: confluent local demo
	// TODO: confluent local destroy
	localCommand.AddCommand(NewPluginsCommand(prerunner))
	localCommand.AddCommand(NewServicesCommand(prerunner))
	// TODO: confluent local topics
	localCommand.AddCommand(NewVersionCommand(prerunner))

	return localCommand.Command
}
