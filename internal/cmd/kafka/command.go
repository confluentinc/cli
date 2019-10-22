package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config    *config.Config
	prerunner pcmd.PreRunner
}

// New returns the default command object for interacting with Kafka.
func New(prerunner pcmd.PreRunner, config *config.Config) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "kafka",
			Short:             "Manage Apache Kafka.",
			PersistentPreRunE: prerunner.Authenticated(config),
		},
		config:    config,
		prerunner: prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewTopicCommand(c.prerunner, c.config))
	context := c.config.Context()
	if context != nil && context.Credential.CredentialType == config.APIKey {
		return
	}
	c.AddCommand(NewClusterCommand(c.prerunner, c.config))
	c.AddCommand(NewACLCommand(c.config))
}
