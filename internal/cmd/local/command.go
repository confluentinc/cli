package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

func NewCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	localCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "local-v2 [command]",
			Short: "Manage a local Confluent Platform development environment.",
		},
		cfg, prerunner,
	)

	localCommand.AddCommand(NewConnectorsCommand(prerunner, cfg))
	localCommand.AddCommand(NewCurrentCommand(prerunner, cfg))
	// TODO: confluent local demo
	// TODO: confluent local destroy
	localCommand.AddCommand(NewPluginsCommand(prerunner, cfg))
	localCommand.AddCommand(NewServicesCommand(prerunner, cfg))
	// TODO: confluent local topics
	localCommand.AddCommand(NewVersionCommand(prerunner, cfg))

	return localCommand.Command
}
