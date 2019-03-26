package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config *config.Config
	client ccloud.Kafka
}

// New returns the default command object for interacting with Kafka.
func New(config *config.Config, client ccloud.Kafka) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "kafka",
			Short: "Manage Kafka",
		},
		config: config,
		client: client,
	}
	// Should uncomment this when/if ACL/topic commands need this flag (currently just in cluster cmd)
	//cmd.PersistentFlags().String("environment", "", "ID of the environment in which to run the command")
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewClusterCommand(c.config, c.client))
	c.AddCommand(NewTopicCommand(c.config, c.client))
	c.AddCommand(NewACLCommand(c.config, c.client))
}
