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
	if cfg.IsCloudLogin() {
		topicCmd := NewTopicCommand(isAPIKeyLogin, c.prerunner, c.logger, c.clientID)
		// Order matters here. If we add to the server-side completer first then the command doesn't have a parent
		// and that doesn't trigger completion.
		c.AddCommand(topicCmd.hasAPIKeyTopicCommand.Command)
		c.serverCompleter.AddCommand(topicCmd)

		if isAPIKeyLogin {
			return
		}

		aclCmd := NewACLCommand(c.prerunner)
		clusterCmd := NewClusterCommand(c.prerunner, c.analyticsClient)
		groupCmd := NewGroupCommand(c.prerunner, c.serverCompleter)

		c.AddCommand(aclCmd.Command)
		c.AddCommand(clusterCmd.Command)
		c.AddCommand(groupCmd.Command)
		c.AddCommand(NewLinkCommand(c.prerunner))
		c.AddCommand(NewMirrorCommand(c.prerunner))
		c.AddCommand(NewRegionCommand(c.prerunner))

		c.serverCompleter.AddCommand(aclCmd)
		c.serverCompleter.AddCommand(clusterCmd)
		c.serverCompleter.AddCommand(groupCmd)
		c.serverCompleter.AddCommand(groupCmd.lagCmd)

		return
	}

	// These on-prem commands can also be run without logging in.
	c.AddCommand(NewAclCommandOnPrem(c.prerunner))
	c.AddCommand(NewTopicCommandOnPrem(c.prerunner))

	if cfg.IsOnPremLogin() {
		c.AddCommand(NewClusterCommandOnPrem(c.prerunner))
	}
}
