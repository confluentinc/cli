package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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
func New(isAPIKeyLogin bool, cliName string, prerunner pcmd.PreRunner, logger *log.Logger, clientID string,
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
	cmd.init(isAPIKeyLogin, cliName)
	return cmd.Command
}

func (c *command) init(isAPIKeyLogin bool, cliName string) {
	if cliName == "ccloud" {
		topicCmd := NewTopicCommand(isAPIKeyLogin, c.prerunner, c.logger, c.clientID)
		// Order matters here. If we add to the server-side completer first then the command doesn't have a parent
		// and that doesn't trigger completion.
		c.AddCommand(topicCmd.hasAPIKeyTopicCommand.Command)
		c.serverCompleter.AddCommand(topicCmd)
		if isAPIKeyLogin {
			return
		}
		groupCmd := NewGroupCommand(c.prerunner, c.serverCompleter)
		c.AddCommand(groupCmd.Command)
		c.serverCompleter.AddCommand(groupCmd)
		c.serverCompleter.AddCommand(groupCmd.lagCmd)
		clusterCmd := NewClusterCommand(c.prerunner, c.analyticsClient, c.logger)
		c.AddCommand(clusterCmd.Command)
		c.serverCompleter.AddCommand(clusterCmd)
		aclCmd := NewACLCommand(c.prerunner)
		c.AddCommand(aclCmd.Command)
		c.serverCompleter.AddCommand(aclCmd)
		c.AddCommand(NewRegionCommand(c.prerunner))
		c.AddCommand(NewLinkCommand(c.prerunner))
		c.AddCommand(NewMirrorCommand(c.prerunner))
	} else {
		c.AddCommand(NewClusterCommandOnPrem(c.prerunner))
		c.AddCommand(NewTopicCommandOnPrem(c.prerunner))
		c.AddCommand(NewAclCommandOnPrem(c.prerunner))
		c.AddCommand(NewPartitionCommandOnPrem(c.prerunner))
	}
}
