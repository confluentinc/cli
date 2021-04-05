package kafka

import (
	"fmt"
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
		fmt.Print("\ncalling NewGroupCommand\n")
		groupCmd := NewGroupCommand(c.prerunner, c.serverCompleter)
		// // tried moving this block down to after lagCmd is initialized, still had segfault
		//c.AddCommand(groupCmd.Command)
		//c.serverCompleter.AddCommand(groupCmd)
		//
		fmt.Print("\ncalling NewLagCommand\n")
		lagCmd := NewLagCommand(c.prerunner, groupCmd)
		fmt.Printf("\n about to call groupCmd.AddCommand(lagCmd.Command), let's make sure group's PersistentPreRunE exists: %p\n", groupCmd.AuthenticatedCLICommand.PersistentPreRunE)
		groupCmd.AddCommand(lagCmd.Command)
		fmt.Print("\n CALLING g.serverCompleter.AddSubCommand(lagCmd)\n")
		fmt.Printf("   lagCmd.Command is %p\n", lagCmd.Command)
		groupCmd.serverCompleter.AddSubCommand(lagCmd)
		fmt.Print("\n RETURNED\n")
		groupCmd.completableChildren = append(groupCmd.completableChildren, lagCmd.completableChildren...)
		groupCmd.completableFlagChildren["cluster"] = append(groupCmd.completableFlagChildren["cluster"], lagCmd.completableChildren...)

		// block moved down from before lagCmd was initialized
		c.AddCommand(groupCmd.Command)
		fmt.Printf("    groupCmd.Command is %p\n", groupCmd.Command)
		fmt.Print("\nabout to call c.serverCompleter.AddCommand(groupCmd)\n")
		c.serverCompleter.AddCommand(groupCmd)

		fmt.Printf("\ncalling NewClusterCommand\n")
		clusterCmd := NewClusterCommand(c.prerunner, c.analyticsClient)
		c.AddCommand(clusterCmd.Command)
		c.serverCompleter.AddCommand(clusterCmd)
		aclCmd := NewACLCommand(c.prerunner)
		c.AddCommand(aclCmd.Command)
		c.serverCompleter.AddCommand(aclCmd)
		c.AddCommand(NewRegionCommand(c.prerunner))
		c.AddCommand(NewLinkCommand(c.prerunner))
	} else {
		c.AddCommand(NewClusterCommandOnPrem(c.prerunner))
		c.AddCommand(NewTopicCommandOnPrem(c.prerunner))
	}
}
