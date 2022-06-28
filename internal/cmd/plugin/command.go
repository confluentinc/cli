package plugin

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage Confluent plugins.",
	}

	cmd.AddCommand(newListCommand())
	c := &command{CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner)}
	return c.Command
}
