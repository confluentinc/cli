package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type computePoolCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

type out struct {
	Id   string `human:"ID" serialized:"id"`
	Name string `human:"Name" serialized:"name"`
}

func newComputePoolCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute-pool",
		Short: "Manage Flink compute pools.",
	}

	c := &computePoolCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}
