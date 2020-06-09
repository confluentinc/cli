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

	// TODO: confluent local acl
	// TODO: confluent local config
	// TODO: confluent local consume
	// TODO: confluent local current
	// TODO: confluent local demo
	// TODO: confluent local destroy
	// TODO: confluent local list
	// TODO: confluent local load
	// TODO: confluent local log
	// TODO: confluent local produce
	// TODO: confluent local start
	// TODO: confluent local status
	// TODO: confluent local stop
	// TODO: confluent local top
	// TODO: confluent local unload
	localCommand.AddCommand(NewVersionCommand(prerunner))

	return localCommand.Command
}
