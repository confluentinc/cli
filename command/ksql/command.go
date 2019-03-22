package ksql

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/cli/shared"
)

type command struct {
	*cobra.Command
	config *shared.Config
	client ccloud.KSQL
}

// New returns the default command object for interacting with KSQL.
func New(config *shared.Config, client ccloud.KSQL) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "ksql",
			Short: "Manage KSQL",
		},
		config: config,
		client: client,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewClusterCommand(c.config, c.client))
}
