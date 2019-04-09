package rbac

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go"

	"context"
	"fmt"
	"strings"
)

var (
	roleListFields     = []string{"Name", "SuperUser", "AllowedOperations"}
	roleListLabels     = []string{"Name", "SuperUser", "AllowedOperations"}
	roleDescribeFields = []string{"Name", "SuperUser", "AllowedOperations"}
	roleDescribeLabels = []string{"Name", "SuperUser", "AllowedOperations"}
)

type rolesCommand struct {
	*cobra.Command
	config *config.Config
	client *mds.APIClient
}

// NewRolesCommand returns the sub-command object for interacting with RBAC roles.
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
		Short: "List available roles, ... or roles for a given principal", // TODO https://confluentinc.atlassian.net/wiki/spaces/PM/pages/804029551/UX+CLI+for+defining+roles ?
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})

	c.AddCommand(&cobra.Command{
		Use:   "describe ROLE",
		Short: "Describes resourceTypes and operations allowed for a given role",
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
		data = append(data, printer.ToRow(&role, roleListFields))
	}
	printer.RenderCollectionTable(data, roleListLabels)

	return nil
}

func (c *rolesCommand) describe(cmd *cobra.Command, args []string) error {
	role := args[0]

	details, r, err := c.client.RoleDefinitionsApi.RoleDetail(context.Background(), role)
	if err != nil {
		if r.StatusCode == 204 {
			availableRoleNames, _, err := c.client.RoleDefinitionsApi.Rolenames(context.Background())
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}

			cmd.SilenceUsage = true
			return fmt.Errorf("Unknown role specified.  Role should be one of " + strings.Join(availableRoleNames, ", "))
		}

		return errors.HandleCommon(err, cmd)
	}

	var data [][]string
	data = append(data, printer.ToRow(&details, roleDescribeFields))
	printer.RenderCollectionTable(data, roleDescribeLabels)

	return nil
}
