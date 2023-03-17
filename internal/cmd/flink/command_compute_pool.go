package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type computePoolOut struct {
	Id   string `human:"ID" serialized:"id"`
	Name string `human:"Name" serialized:"name"`
}

func (c *command) newComputePoolCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute-pool",
		Short: "Manage Flink compute pools.",
	}

	cmd.AddCommand(c.newComputePoolCreateCommand())
	cmd.AddCommand(c.newComputePoolDeleteCommand())
	cmd.AddCommand(c.newComputePoolDescribeCommand())
	cmd.AddCommand(c.newComputePoolListCommand())
	cmd.AddCommand(c.newComputePoolUpdateCommand())

	return cmd
}
