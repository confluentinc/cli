package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
}

// New returns the default command object for interacting with Kafka.
func New(prerunner pcmd.PreRunner, config *config.Config) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "kafka",
			Short: "Manage Apache Kafka.",
		},
		config, prerunner)
	cmd := &command{CLICommand: cliCmd, prerunner: prerunner}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewTopicCommand(c.prerunner, c.Config))
	context := c.Config.Context()
	if context != nil && context.Credential.CredentialType == config.APIKey { // TODO: Does not work with --context flag.
		return
	}
	c.AddCommand(NewClusterCommand(c.prerunner, c.Config))
	c.AddCommand(NewACLCommand(c.Config, c.prerunner))
}
