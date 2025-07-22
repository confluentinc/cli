package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type consumerGroupOut struct {
	Cluster           string `human:"Cluster" serialized:"cluster"`
	ConsumerGroup     string `human:"Consumer Group" serialized:"consumer_group"`
	Coordinator       string `human:"Coordinator" serialized:"coordinator"`
	IsSimple          bool   `human:"Simple" serialized:"is_simple"`
	PartitionAssignor string `human:"Partition Assignor" serialized:"partition_assignor"`
	State             string `human:"State" serialized:"state"`
}

func (c *consumerCommand) newGroupCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage Kafka consumer groups.",
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

	return pcmd.AutocompleteConsumerGroups(cmd, c.AuthenticatedCLICommand)
}
