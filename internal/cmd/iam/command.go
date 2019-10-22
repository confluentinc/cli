package iam

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
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
			Use:               "iam",
			Short:             "Manage RBAC, ACL and IAM permissions.",
			Long:              "Manage Role Based Access (RBAC), Access Control Lists (ACL), and Identity and Access Management (IAM) permissions.",
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config: config,
		client: client,
	}

	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewRoleCommand(c.config, c.client))
	c.AddCommand(NewRolebindingCommand(c.config, c.client))
	c.AddCommand(NewACLCommand(c.config, c.client))
}
