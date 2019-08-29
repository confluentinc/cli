package kafka

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/confluentinc/ccloud-sdk-go"

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

// New returns the default command object for interacting with Kafka.
func New(prerunner pcmd.PreRunner, config *config.Config, client ccloud.Kafka, ch *pcmd.ConfigHelper) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "kafka",
			Short:             "Manage Apache Kafka.",
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config:    config,
		client:    client,
		ch:        ch,
		prerunner: prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	credType, err := c.config.CredentialType()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c.AddCommand(NewTopicCommand(c.prerunner, c.config, c.client, c.ch))
	if credType == config.Username {
		c.AddCommand(NewClusterCommand(c.config, c.client, c.ch))
		c.AddCommand(NewACLCommand(c.config, c.client, c.ch))
	}
}
