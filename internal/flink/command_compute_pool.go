package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type computePoolOut struct {
	IsCurrent   bool   `human:"Current" serialized:"is_current"`
	Id          string `human:"ID" serialized:"id"`
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	CurrentCfu  int32  `human:"Current CFU" serialized:"currrent_cfu"`
	MaxCfu      int32  `human:"Max CFU" serialized:"max_cfu"`
	Cloud       string `human:"Cloud" serialized:"cloud"`
	Region      string `human:"Region" serialized:"region"`
	Status      string `human:"Status" serialized:"status"`
}

func (c *command) newComputePoolCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "compute-pool",
		Short:       "Manage Flink compute pools.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newComputePoolCreateCommand())
	cmd.AddCommand(c.newComputePoolDeleteCommand())
	cmd.AddCommand(c.newComputePoolDescribeCommand())
	cmd.AddCommand(c.newComputePoolListCommand())
	cmd.AddCommand(c.newComputePoolUnsetCommand())
	cmd.AddCommand(c.newComputePoolUpdateCommand())
	cmd.AddCommand(c.newComputePoolUseCommand())

	return cmd
}

func (c *command) validComputePoolArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.autocompleteComputePools(cmd, args)
}
