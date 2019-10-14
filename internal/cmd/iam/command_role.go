package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/mds-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
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
func NewRoleCommand(cfg *config.Config, client *mds.APIClient) *cobra.Command {
	roleCmd := &roleCommand{
		Command: &cobra.Command{
			Use:   "role",
			Short: "Manage RBAC and IAM roles.",
			Long:  "Manage Role Based Access (RBAC) and Identity and Access Management (IAM) roles.",
		},
		config: cfg,
		client: client,
	}
	state, err := cfg.AuthenticatedState()
	if err != nil {
		state = new(config.ContextState)
	}
	roleCmd.ctx = context.WithValue(context.Background(), mds.ContextAccessToken, state.AuthToken)
	roleCmd.init()
	return roleCmd.Command
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
