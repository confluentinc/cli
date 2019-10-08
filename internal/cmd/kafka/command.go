package kafka

import (
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config    *config.Config
	client    ccloud.Kafka
	ch        *pcmd.ConfigHelper
	prerunner pcmd.PreRunner
}

func (c *command) SetContext(context *config.Context) {

}

// New returns the default command object for interacting with Kafka.
func New(prerunner pcmd.PreRunner, config *config.Config, client ccloud.Kafka, ch *pcmd.ConfigHelper) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "kafka",
			Short: "Manage Apache Kafka.",
		},
		config:    config,
		client:    client,
		ch:        ch,
		prerunner: prerunner,
	}
	cmd.PersistentPreRunE = prerunner.Authenticated(cmd)
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewTopicCommand(c.prerunner, c.config, c.client, c.ch))
	context := c.config.Context()
	if context != nil && context.Credential.CredentialType == config.APIKey {
		return
	}
	c.AddCommand(NewClusterCommand(c.prerunner, c.config, c.client, c.ch))
	c.AddCommand(NewACLCommand(c.config, c.client, c.ch))
}
