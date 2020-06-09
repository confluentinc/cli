package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type command struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
	logger    *log.Logger
	clientID  string
}

// New returns the default command object for interacting with Kafka.
func New(prerunner pcmd.PreRunner, logger *log.Logger, clientID string) *cobra.Command {
	cliCmd := pcmd.NewCLICommand(
		&cobra.Command{
			Use:   "kafka",
			Short: "Manage Apache Kafka.",
		}, prerunner)
	cmd := &command{
		CLICommand: cliCmd,
		prerunner:  prerunner,
		logger:     logger,
		clientID:   clientID,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewTopicCommand(c.prerunner, c.logger, c.clientID))
	c.AddCommand(NewClusterCommand(c.prerunner))
	c.AddCommand(NewACLCommand(c.prerunner))
	c.AddCommand(NewRegionCommand(c.prerunner))
}
