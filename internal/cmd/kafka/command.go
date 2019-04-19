package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config *config.Config
	client ccloud.Kafka
	ch     *pcmd.ConfigHelper
}

// New returns the default command object for interacting with Kafka.
func New(prerunner pcmd.PreRunner, config *config.Config, client ccloud.Kafka, ch *pcmd.ConfigHelper) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "kafka",
			Short:             "Manage Kafka",
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config: config,
		client: client,
		ch:     ch,
	}
	// Should uncomment this when/if ACL/topic commands need this flag (currently just in cluster cmd)
	//cmd.PersistentFlags().String("environment", "", "ID of the environment in which to run the command")
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewClusterCommand(c.config, c.client))
	c.AddCommand(NewTopicCommand(c.config, c.client, c.ch))
	c.AddCommand(NewACLCommand(c.config, c.client, c.ch))
}
