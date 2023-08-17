package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

type consumerGroupOut struct {
	ClusterId         string `human:"Cluster" serialized:"cluster_id"`
	ConsumerGroupId   string `human:"Consumer Group" serialized:"consumer_group_id"`
	Coordinator       string `human:"Coordinator" serialized:"coordinator"`
	IsSimple          bool   `human:"Simple" serialized:"is_simple"`
	PartitionAssignor string `human:"Partition Assignor" serialized:"partition_assignor"`
	State             string `human:"State" serialized:"state"`
}

func (c *consumerCommand) newGroupCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "group",
		Short:       "Manage Kafka consumer groups.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newGroupDescribeCommand())
		cmd.AddCommand(c.newGroupListCommand())
	} else {
		cmd.AddCommand(c.newGroupDescribeCommandOnPrem())
		cmd.AddCommand(c.newGroupListCommandOnPrem())
	}
	cmd.AddCommand(c.newLagCommand(cfg))

	return cmd
}

func (c *consumerCommand) validGroupArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteConsumerGroups(c.AuthenticatedCLICommand)
}
