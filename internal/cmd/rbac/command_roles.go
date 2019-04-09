package rbac

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go"

	"context"
	//"fmt"
)

var (
	listFields     = []string{"Name", "SuperUser", "AllowedOperations"}
	listLabels     = []string{"Name", "SuperUser", "AllowedOperations"}
	describeFields = []string{"Name", "SuperUser", "AllowedOperations"}
	describeLabels = []string{"Name", "SuperUser", "AllowedOperations"}
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

	c.AddCommand(&cobra.Command{
		Use:   "describe ROLE",
		Short: "Describes resoruceTypes and operations allowed for a given role",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	})
}

func (c *rolesCommand) list(cmd *cobra.Command, args []string) error {
	roles, _, err := c.client.RoleDefinitionsApi.Roles(context.Background())
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	var data [][]string
	for _, role := range roles {
		data = append(data, printer.ToRow(&role, listFields))
	}
	printer.RenderCollectionTable(data, listLabels)

	return nil
}

func (c *rolesCommand) describe(cmd *cobra.Command, args []string) error {
	role := args[0]

	details, _, err := c.client.RoleDefinitionsApi.RoleDetail(context.Background(), role)
	if err != nil {
		return errors.HandleCommon(err.(error), cmd)
	}

	var data [][]string
	data = append(data, printer.ToRow(&details, describeFields))
	printer.RenderCollectionTable(data, describeLabels)

	return nil
}
