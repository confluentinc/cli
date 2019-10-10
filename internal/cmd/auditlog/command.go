package auditlog

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/mds-sdk-go"
)

type command struct {
	*cobra.Command
	config *config.Config
	client *mds.APIClient
}

// New returns the default command object for interacting with RBAC.
func New(prerunner pcmd.PreRunner, config *config.Config, client *mds.APIClient) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "auditlog",
			Short:             "Manage audit log configuration.",
			Long:              "Manage which auditable events are logged, and where the events are sent.",
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config: config,
		client: client,
	}

	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewConfigCommand(c.config, c.client))
	c.AddCommand(NewRouteCommand(c.config, c.client))
}
