package rbac

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
	mds "github.com/confluentinc/mds-sdk-go"
)

type command struct {
	*cobra.Command
	config *config.Config
	client *mds.APIClient
}

// New returns the default command object for interacting with RBAC.
func New(config *config.Config) *cobra.Command {
	cfg := mds.NewConfiguration()
	cfg.BasePath = "http://localhost:8090"       // TODO
	cfg.UserAgent = "OpenAPI-Generator/1.0.0/go" // TODO

	client := mds.NewAPIClient(cfg)

	cmd := &command{
		Command: &cobra.Command{
			Use:   "iam",
			Short: "Manage RBAC/IAM permissions",
		},
		config: config,
		client: client,
	}

	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewRolesCommand(c.config, c.client))
	c.AddCommand(NewRolebindingsCommand(c.config, c.client))
}
