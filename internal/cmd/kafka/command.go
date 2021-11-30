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
	prerunner       pcmd.PreRunner
	logger          *log.Logger
	clientID        string
	serverCompleter completer.ServerSideCompleter
	analyticsClient analytics.Client
}

// New returns the default command object for interacting with Kafka.
func New(cfg *v1.Config, prerunner pcmd.PreRunner, logger *log.Logger, clientID string,
	serverCompleter completer.ServerSideCompleter, analyticsClient analytics.Client) *cobra.Command {
	cliCmd := pcmd.NewCLICommand(
		&cobra.Command{
			Use:   "kafka",
			Short: "Manage Apache Kafka.",
		}, prerunner)
	cmd := &command{
		CLICommand:      cliCmd,
		prerunner:       prerunner,
		logger:          logger,
		clientID:        clientID,
		serverCompleter: serverCompleter,
		analyticsClient: analyticsClient,
	}
	cmd.init(cfg)
	return cmd.Command
}

func (c *command) init(cfg *v1.Config) {
	aclCmd := NewACLCommand(cfg, c.prerunner)
	clusterCmd := NewClusterCommand(cfg, c.prerunner, c.analyticsClient)
	groupCmd := NewGroupCommand(c.prerunner, c.serverCompleter)
	topicCmd := NewTopicCommand(cfg, c.prerunner, c.logger, c.clientID)

	c.AddCommand(NewBrokerCommand(c.prerunner))
	c.AddCommand(NewReplicaCommand(c.prerunner))
	c.AddCommand(NewLinkCommand(c.prerunner))
	c.AddCommand(NewMirrorCommand(c.prerunner))
	c.AddCommand(NewPartitionCommand(c.prerunner))
	c.AddCommand(NewRegionCommand(c.prerunner))
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
		c.serverCompleter.AddCommand(groupCmd.lagCmd)
		c.serverCompleter.AddCommand(topicCmd)
	}
}
