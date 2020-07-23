package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type command struct {
	// this is anonymous typed
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
	logger    *log.Logger
	clientID  string
}

// Returns the command describing the kafka subcommands, which is added to the base command
// New returns the default command object for interacting with Kafka.
func New(isAPIKeyLogin bool, cliName string, prerunner pcmd.PreRunner, logger *log.Logger, clientID string) *cobra.Command {
	cliCmd := pcmd.NewCLICommand(
		&cobra.Command{
			Use:   "kafka",
			Short: "Manage Apache Kafka.",
		}, prerunner)

	// "command" struct contains info needed for kafkaCommand
	cmd := &command{
		// and CLICommand contains info needed for CLICommand (prerunner information)
		CLICommand: cliCmd,
		prerunner:  prerunner,
		logger:     logger,
		clientID:   clientID,
	}
	// init registers all the commands
	cmd.init(isAPIKeyLogin, cliName)
	// since CLICommand is anonymous, CLICommand's internals can be directly accessed
	// equivalent to cmd.CLICommand.Command. Now why build the command wrapper then?
	// command_topic_onprem.go shows that when we register RunE, we can register
	// (for example) topicCommand.List(), which I then have information stored in
	// topicCommand for the List function.
	return cmd.Command
}

func (c *command) init(isAPIKeyLogin bool, cliName string) {
	if cliName == "ccloud" {
		c.AddCommand(NewTopicCommand(isAPIKeyLogin, c.prerunner, c.logger, c.clientID))
		if isAPIKeyLogin {
			return
		}
		c.AddCommand(NewClusterCommand(c.prerunner))
		c.AddCommand(NewACLCommand(c.prerunner))
		c.AddCommand(NewRegionCommand(c.prerunner))
	} else {
		c.AddCommand(NewClusterCommandOnPrem(c.prerunner))
		// each of these commands, since CLICommand is anonymous in kafkaCmd, and cobra.Command is anonymous in CLICommand
		// this is equivalent to c.CLICommand.Command.AddCommand
	}
}
