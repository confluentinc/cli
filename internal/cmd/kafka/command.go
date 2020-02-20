package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type command struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
	logger    *log.Logger
	clientID  string
}

// New returns the default command object for interacting with Kafka.
func New(prerunner pcmd.PreRunner, config *v2.Config, logger *log.Logger, clientID string) *cobra.Command {
	cliCmd := pcmd.NewCLICommand(
		&cobra.Command{
			Use:   "kafka",
			Short: "Manage Apache Kafka.",
		},
		config, prerunner)
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
	c.AddCommand(NewTopicCommand(c.prerunner, c.Config.Config, c.logger, c.clientID))
	context := c.Config.Config.Context()
	if context != nil && context.Credential.CredentialType == v2.APIKey { // TODO: Change to DynamicConfig to handle flags.
		return
	}
	c.AddCommand(NewClusterCommand(c.prerunner, c.Config.Config))
	c.AddCommand(NewACLCommand(c.prerunner, c.Config.Config))
	c.AddCommand(NewRegionCommand(c.prerunner, c.Config.Config))
}
