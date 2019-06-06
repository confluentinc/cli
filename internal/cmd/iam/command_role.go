package iam

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var (
	roleListFields     = []string{"Name", "AccessPolicy"}
	roleListLabels     = []string{"Name", "AccessPolicy"}
	roleDescribeFields = []string{"Name", "AccessPolicy"}
	roleDescribeLabels = []string{"Name", "AccessPolicy"}
)

type roleCommand struct {
	*cobra.Command
	config *config.Config
	client *mds.APIClient
	ctx    context.Context
}

// NewRoleCommand returns the sub-command object for interacting with RBAC roles.
func NewRoleCommand(config *config.Config, client *mds.APIClient) *cobra.Command {
	cmd := &roleCommand{
		Command: &cobra.Command{
			Use:   "role",
			Short: "Manage RBAC and IAM roles.",
			Long:  "Manage Role Based Access (RBAC) and Identity and Access Management (IAM) roles.",
		},
		config: config,
		client: client,
		ctx:    context.WithValue(context.Background(), mds.ContextAccessToken, config.AuthToken),
	}

	cmd.init()
	return cmd.Command
}

func (c *roleCommand) init() {
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List the available roles.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})

	c.AddCommand(&cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the resources and operations allowed for a role.",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	})
}

func (c *roleCommand) list(cmd *cobra.Command, args []string) error {
	roles, _, err := c.client.RoleDefinitionsApi.Roles(c.ctx)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	tablePrinter := tablewriter.NewWriter(os.Stdout)
	var data [][]string
	for _, role := range roles {
		marshalled, err := json.Marshal(role.AccessPolicy)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		prettyRole := struct {
			Name         string
			AccessPolicy string
		}{
			role.Name,
			string(pretty.Pretty(marshalled)),
		}
		data = append(data, printer.ToRow(&prettyRole, roleListFields))
	}
	tablePrinter.SetAutoWrapText(false)
	tablePrinter.SetAutoFormatHeaders(false)
	tablePrinter.SetHeader(roleListLabels)
	tablePrinter.AppendBulk(data)
	tablePrinter.SetBorder(false)
	tablePrinter.Render()

	return nil
}

func (c *roleCommand) describe(cmd *cobra.Command, args []string) error {
	role := args[0]

	details, r, err := c.client.RoleDefinitionsApi.RoleDetail(c.ctx, role)
	if err != nil {
		if r.StatusCode == http.StatusNoContent {
			availableRoleNames, _, err := c.client.RoleDefinitionsApi.Rolenames(c.ctx)
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
