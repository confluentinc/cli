package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

type computePoolOut struct {
	IsCurrent  bool   `human:"Current" serialized:"is_current"`
	Id         string `human:"ID" serialized:"id"`
	Name       string `human:"Name" serialized:"name"`
	CurrentCfu int32  `human:"Current CFU" serialized:"currrent_cfu"`
	MaxCfu     int32  `human:"Max CFU" serialized:"max_cfu"`
	Region     string `human:"Region" serialized:"region"`
	Status     string `human:"Status" serialized:"status"`
}

func (c *command) newComputePoolCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute-pool",
		Short: "Manage Flink compute pools.",
	}

	cmd.AddCommand(c.newComputePoolDescribeCommand())
	cmd.AddCommand(c.newComputePoolListCommand())
	cmd.AddCommand(c.newComputePoolUpdateCommand())
	cmd.AddCommand(c.newComputePoolUseCommand())
	cmd.AddCommand(c.newComputePoolCreateCommand())
	cmd.AddCommand(c.newComputePoolDeleteCommand())

	return cmd
}

func (c *command) validComputePoolArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validComputePoolArgsMultiple(cmd, args)
}

func (c *command) validComputePoolArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteComputePools()
}

func (c *command) autocompleteComputePools() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	computePools, err := c.V2Client.ListFlinkComputePools(environmentId, "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(computePools))
	for i, computePool := range computePools {
		suggestions[i] = fmt.Sprintf("%s\t%s", computePool.GetId(), computePool.Spec.GetDisplayName())
	}
	return suggestions
}

func (c *command) addRegionFlag(cmd *cobra.Command) {
	cmd.Flags().String("region", "", `Cloud region for compute pool (use "confluent flink region list" to see all).`)
	pcmd.RegisterFlagCompletionFunc(cmd, "region", c.autocompleteRegions)
}
