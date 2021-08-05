package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
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
func New(cfg *v3.Config, isAPIKeyLogin bool, prerunner pcmd.PreRunner, logger *log.Logger, clientID string,
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
	cmd.init(cfg, isAPIKeyLogin)
	return cmd.Command
}

func (c *command) init(cfg *v3.Config, isAPIKeyLogin bool) {
	groupCmd := NewGroupCommand(c.prerunner, c.serverCompleter)

	c.AddCommand(groupCmd.Command)
	c.AddCommand(NewRegionCommand(c.prerunner))
	c.AddCommand(NewLinkCommand(c.prerunner))
	c.AddCommand(NewMirrorCommand(c.prerunner))

	if cfg.IsCloudLogin() {
		c.serverCompleter.AddCommand(groupCmd)
		c.serverCompleter.AddCommand(groupCmd.lagCmd)
	}

	// TODO: Combine
	if cfg.IsCloudLogin() {
		aclCmd := NewACLCommand(c.prerunner)
		clusterCmd := NewClusterCommand(c.prerunner, c.analyticsClient)
		topicCmd := NewTopicCommand(isAPIKeyLogin, c.prerunner, c.logger, c.clientID)

		c.AddCommand(aclCmd.Command)
		c.AddCommand(clusterCmd.Command)
		c.AddCommand(topicCmd.hasAPIKeyTopicCommand.Command)

		c.serverCompleter.AddCommand(aclCmd)
		c.serverCompleter.AddCommand(clusterCmd)
		c.serverCompleter.AddCommand(topicCmd)

		return
	}

	// These on-prem commands can also be run without logging in.
	c.AddCommand(NewAclCommandOnPrem(c.prerunner))
	c.AddCommand(NewTopicCommandOnPrem(c.prerunner))

	if cfg.IsOnPremLogin() {
		c.AddCommand(NewClusterCommandOnPrem(c.prerunner))
	}
}
