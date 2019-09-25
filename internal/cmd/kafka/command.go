package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type command struct {
	*cobra.Command
	config    *config.Config
	client    ccloud.Kafka
	ch        *pcmd.ConfigHelper
	prerunner pcmd.PreRunner
	logger    *log.Logger
	clientID  string
}

// New returns the default command object for interacting with Kafka.
func New(prerunner pcmd.PreRunner, config *config.Config, logger *log.Logger, clientID string, client ccloud.Kafka, ch *pcmd.ConfigHelper) (*cobra.Command, error) {
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
		logger:    logger,
		clientID:  clientID,
	}
	err := cmd.init()
	if err != nil {
		return nil, err
	}
	return cmd.Command, nil
}

func (c *command) init() error {
	topicCmd, err := NewTopicCommand(c.prerunner, c.config, c.logger, c.clientID, c.client, c.ch)
	if err != nil {
		return err
	}
	c.AddCommand(topicCmd)
	credType, err := c.config.CredentialType()
	if err != nil && err != errors.ErrNoContext {
		return err
	}
	if err == nil && credType == config.APIKey {
		return nil
	}
	c.AddCommand(NewClusterCommand(c.config, c.client, c.ch))
	c.AddCommand(NewACLCommand(c.config, c.client, c.ch))
	return nil
}
