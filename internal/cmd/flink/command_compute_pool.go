package flink

import (
	"github.com/spf13/cobra"
)

type computePoolOut struct {
	Current bool   `human:"Is Current" serialized:"is_current"`
	Id      string `human:"ID" serialized:"id"`
	Name    string `human:"Name" serialized:"name"`
}

func (c *command) newComputePoolCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute-pool",
		Short: "Manage Flink compute pools.",
	}

	cmd.AddCommand(c.newComputePoolDescribeCommand())
	cmd.AddCommand(c.newComputePoolListCommand())
	cmd.AddCommand(c.newComputePoolUseCommand())

	return cmd
}
