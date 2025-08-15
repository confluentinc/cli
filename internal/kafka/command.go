package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage Apache Kafka.",
	}

	cmd.AddCommand(newAclCommand(cfg, prerunner))
	cmd.AddCommand(newBrokerCommand(prerunner))
	cmd.AddCommand(newClientConfigCommand(cfg, prerunner))
	cmd.AddCommand(newClusterCommand(cfg, prerunner))
	cmd.AddCommand(newConsumerCommand(cfg, prerunner))
	cmd.AddCommand(newLinkCommand(cfg, prerunner))
	cmd.AddCommand(newMirrorCommand(prerunner))
	cmd.AddCommand(newPartitionCommand(cfg, prerunner))
	cmd.AddCommand(newQuotaCommand(prerunner))
	cmd.AddCommand(newRegionCommand(prerunner))
	cmd.AddCommand(newReplicaCommand(prerunner))
	cmd.AddCommand(newTopicCommand(cfg, prerunner))

	return cmd
}
