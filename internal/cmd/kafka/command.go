package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type command struct {
	*pcmd.CLICommand
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner, clientID string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage Apache Kafka.",
	}

	c := &command{pcmd.NewCLICommand(cmd, prerunner)}

	clusterCmd := newClusterCommand(cfg, prerunner)

	cmd.AddCommand(newAclCommand(cfg, prerunner))
	cmd.AddCommand(newBrokerCommand(prerunner))
	cmd.AddCommand(newClientConfigCommand(prerunner, clientID))
	cmd.AddCommand(clusterCmd.Command)
	cmd.AddCommand(newConsumerGroupCommand(prerunner))
	cmd.AddCommand(newLinkCommand(cfg, prerunner))
	cmd.AddCommand(newMirrorCommand(prerunner))
	cmd.AddCommand(newPartitionCommand(prerunner))
	cmd.AddCommand(newRegionCommand(prerunner))
	cmd.AddCommand(newReplicaCommand(prerunner))

	topicCmd := newTopicCommand(cfg, prerunner, clientID)
	if topicCmd.hasAPIKeyTopicCommand != nil {
		c.AddCommand(topicCmd.hasAPIKeyTopicCommand.Command)
	} else if topicCmd.authenticatedTopicCommand != nil {
		c.AddCommand(topicCmd.authenticatedTopicCommand.Command)
	}

	return cmd
}
