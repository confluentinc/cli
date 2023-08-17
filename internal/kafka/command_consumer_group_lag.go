package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/config"
)

type lagOut struct {
	ClusterId       string `human:"Cluster" serialized:"cluster"`
	ConsumerGroupId string `human:"Consumer Group" serialized:"consumer_group"`
	Lag             int64  `human:"Lag" serialized:"lag"`
	LogEndOffset    int64  `human:"Log End Offset" serialized:"log_end_offset"`
	CurrentOffset   int64  `human:"Current Offset" serialized:"current_offset"`
	ConsumerId      string `human:"Consumer" serialized:"consumer"`
	InstanceId      string `human:"Instance" serialized:"instance"`
	ClientId        string `human:"Client" serialized:"client"`
	TopicName       string `human:"Topic" serialized:"topic"`
	PartitionId     int32  `human:"Partition" serialized:"partition"`
}

func (c *consumerCommand) newLagCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "lag",
		Short:  "View consumer lag.",
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newLagGetCommand())
		cmd.AddCommand(c.newLagListCommand())
		cmd.AddCommand(c.newLagSummarizeCommand())
	} else {
		cmd.AddCommand(c.newLagGetCommandOnPrem())
		cmd.AddCommand(c.newLagListCommandOnPrem())
		cmd.AddCommand(c.newLagSummarizeCommandOnPrem())
	}

	return cmd
}
