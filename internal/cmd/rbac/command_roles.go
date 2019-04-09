package rbac

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go"

	"context"
	"fmt"
)

var (
	listFields = []string{"Name", "SuperUser", "AllowedOperations"}
	listLabels = []string{"Name", "SuperUser", "AllowedOperations"}
)

type rolesCommand struct {
	*cobra.Command
	config *config.Config
	client *mds.APIClient
}

// New returns the default command object for interacting with RBAC.
func NewRolesCommand(config *config.Config, client *mds.APIClient) *cobra.Command {
	cmd := &rolesCommand{
		Command: &cobra.Command{
			Use:   "role",
			Short: "Manage RBAC/IAM roles",
		},
		config: config,
		client: client,
	}

	cmd.init()
	return cmd.Command
}

func (c *rolesCommand) init() {
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List environments",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})
}

func (c *rolesCommand) list(cmd *cobra.Command, args []string) error {
	roles, _, err := c.client.RoleDefinitionsApi.Roles(context.Background())
	if err != nil {
		fmt.Println("err! ", err)
	}

	var data [][]string
	for _, role := range roles {
		data = append(data, printer.ToRow(&role, listFields))
	}
	printer.RenderCollectionTable(data, listLabels)

	return nil
}
