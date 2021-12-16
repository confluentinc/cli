package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
)

type command struct {
	*pcmd.CLICommand
	serverCompleter completer.ServerSideCompleter
	analyticsClient analytics.Client
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner, logger *log.Logger, clientID string, serverCompleter completer.ServerSideCompleter, analyticsClient analytics.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage Apache Kafka.",
	}

	c := &command{
		CLICommand:      pcmd.NewCLICommand(cmd, prerunner),
		serverCompleter: serverCompleter,
		analyticsClient: analyticsClient,
	}

	aclCmd := newAclCommand(cfg, prerunner)
	clusterCmd := newClusterCommand(cfg, prerunner, c.analyticsClient)
	groupCmd := newConsumerGroupCommand(cfg, prerunner, c.serverCompleter)
	topicCmd := newTopicCommand(cfg, prerunner, logger, clientID)

	c.AddCommand(newBrokerCommand(prerunner))
	c.AddCommand(newLinkCommand(prerunner))
	c.AddCommand(newMirrorCommand(prerunner))
	c.AddCommand(newPartitionCommand(prerunner))
	c.AddCommand(newRegionCommand(prerunner))
	c.AddCommand(aclCmd.Command)
	c.AddCommand(clusterCmd.Command)
	c.AddCommand(groupCmd.Command)

	if topicCmd.hasAPIKeyTopicCommand != nil {
		c.AddCommand(topicCmd.hasAPIKeyTopicCommand.Command)
	} else if topicCmd.authenticatedTopicCommand != nil {
		c.AddCommand(topicCmd.authenticatedTopicCommand.Command)
	}

	if cfg.IsCloudLogin() {
		c.serverCompleter.AddCommand(aclCmd)
		c.serverCompleter.AddCommand(clusterCmd)
		c.serverCompleter.AddCommand(groupCmd)
		c.serverCompleter.AddCommand(topicCmd)
	}

	return c.Command
}
