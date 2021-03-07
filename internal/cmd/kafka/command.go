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
		c.AddCommand(topicCmd.hasAPIKeyTopicCommand.Command)
		c.serverCompleter.AddCommand(topicCmd)
		if isAPIKeyLogin {
			return
		}
		clusterCmd := NewClusterCommand(c.prerunner, c.analyticsClient)
		// Order matters here. If we add to the server-side completer first then the command doesn't have a parent
		// and that doesn't trigger completion.
		c.AddCommand(clusterCmd.Command)
		c.serverCompleter.AddCommand(clusterCmd)
		c.AddCommand(NewACLCommand(c.prerunner))
		c.AddCommand(NewRegionCommand(c.prerunner))
		c.AddCommand(NewLinkCommand(c.prerunner))
	} else {
		c.AddCommand(NewClusterCommandOnPrem(c.prerunner))
	}
}
