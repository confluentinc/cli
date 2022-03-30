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

	aclCmd := newAclCommand(cfg, prerunner)
	clusterCmd := newClusterCommand(cfg, prerunner)
	groupCmd := newConsumerGroupCommand(prerunner)
	topicCmd := newTopicCommand(cfg, prerunner, clientID)

	c.AddCommand(newBrokerCommand(prerunner))
	c.AddCommand(newClientConfigCommand(prerunner, clientID))
	c.AddCommand(newLinkCommand(cfg, prerunner))
	c.AddCommand(newMirrorCommand(prerunner))
	c.AddCommand(newPartitionCommand(prerunner))
	c.AddCommand(newReplicaCommand(prerunner))
	c.AddCommand(newRegionCommand(prerunner))
	c.AddCommand(aclCmd.Command)
	c.AddCommand(clusterCmd.Command)
	c.AddCommand(groupCmd.Command)

	if topicCmd.hasAPIKeyTopicCommand != nil {
		c.AddCommand(topicCmd.hasAPIKeyTopicCommand.Command)
	} else if topicCmd.authenticatedTopicCommand != nil {
		c.AddCommand(topicCmd.authenticatedTopicCommand.Command)
	}

	return c.Command
}
