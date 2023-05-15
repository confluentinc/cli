package flink

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/cobra"
)

type computePoolOut struct {
	IsCurrent bool   `human:"Current" serialized:"is_current"`
	Id        string `human:"ID" serialized:"id"`
	Name      string `human:"Name" serialized:"name"`
}

func (c *command) newComputePoolCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute-pool",
		Short: "Manage Flink compute pools.",
	}

	cmd.AddCommand(c.newComputePoolCreateCommand())
	cmd.AddCommand(c.newComputePoolDeleteCommand())
	cmd.AddCommand(c.newComputePoolDescribeCommand())
	cmd.AddCommand(c.newComputePoolListCommand())
	cmd.AddCommand(c.newComputePoolUpdateCommand())
	cmd.AddCommand(c.newComputePoolUseCommand())

	return cmd
}

func (c *command) addRegionFlag(cmd *cobra.Command) {
	cmd.Flags().String("region", "", `Cloud region for compute pool (use "confluent flink region list" to see all).`)
	pcmd.RegisterFlagCompletionFunc(cmd, "region", c.autocompleteRegions)
}

func (c *command) autocompleteRegions(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return nil
	}

	regions, err := c.V2Client.ListFlinkRegions(cloud)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(regions))
	for i, region := range regions {
		suggestions[i] = region.GetId()
	}
	return suggestions
}
