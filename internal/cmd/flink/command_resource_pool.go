package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type resourcePoolCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

type out struct {
	Id string `human:"ID" serialized:"id"`
}

func newResourcePoolCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource-pool",
		Short: "Manage Flink resource pools.",
		Args:  cobra.ExactArgs(1),
	}

	c := &resourcePoolCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}
