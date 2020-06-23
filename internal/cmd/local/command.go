package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

func NewCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "local-v2 [command]",
			Short: "Manage a local Confluent Platform development environment.",
		}, prerunner,
	)

	c.AddCommand(NewCurrentCommand(prerunner))
	c.AddCommand(NewDemoCommand(prerunner))
	c.AddCommand(NewDestroyCommand(prerunner))
	c.AddCommand(NewServicesCommand(prerunner))
	c.AddCommand(NewVersionCommand(prerunner))

	return c.Command
}
