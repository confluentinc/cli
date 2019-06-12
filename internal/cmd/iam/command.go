package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/version"
	mds "github.com/confluentinc/mds-sdk-go"
)

type command struct {
	*cobra.Command
	config *config.Config
	ch     *pcmd.ConfigHelper
	client *mds.APIClient
}

// New returns the default command object for interacting with RBAC.
func New(config *config.Config, ch *pcmd.ConfigHelper, version *version.Version, client *mds.APIClient) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "iam",
			Short: "Manage RBAC and IAM permissions.",
			Long:  "Manage Role Based Access (RBAC) and Identity and Access Management (IAM) permissions.",
		},
		config: config,
		ch:     ch,
		client: client,
	}

	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewRoleCommand(c.config, c.client))
	c.AddCommand(NewRolebindingCommand(c.config, c.ch, c.client))
}
