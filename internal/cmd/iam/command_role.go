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

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	roleFields = []string{"Name", "AccessPolicy"}
	roleLabels = []string{"Name", "AccessPolicy"}
)

type roleCommand struct {
	*cmd.AuthenticatedCLICommand
}

type prettyRole struct {
	Name         string
	AccessPolicy string
}

// NewRoleCommand returns the sub-command object for interacting with RBAC roles.
func NewRoleCommand(cfg *v2.Config, prerunner cmd.PreRunner) *cobra.Command {
	cliCmd := cmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "role",
			Short: "Manage RBAC and IAM roles.",
			Long:  "Manage Role Based Access (RBAC) and Identity and Access Management (IAM) roles.",
		},
		cfg, prerunner)
	roleCmd := &roleCommand{
		AuthenticatedCLICommand: cliCmd,
	}
	roleCmd.init()
	return roleCmd.Command
}

func (c *roleCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
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
	roles, _, err := c.MDSClient.RoleDefinitionsApi.Roles(c.createContext())
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	var data [][]string
	for _, role := range roles {
		roleDisplay, err := createPrettyRole(role)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		data = append(data, printer.ToRow(roleDisplay, roleFields))
	}
	outputTable(data)
	return nil
}

func (c *roleCommand) describe(cmd *cobra.Command, args []string) error {
	role := args[0]

	details, r, err := c.MDSClient.RoleDefinitionsApi.RoleDetail(c.createContext(), role)
	if err != nil {
		if r.StatusCode == http.StatusNoContent {
			availableRoleNames, _, err := c.MDSClient.RoleDefinitionsApi.Rolenames(c.createContext())
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}

			cmd.SilenceUsage = true
			return fmt.Errorf("Unknown role specified.  Role should be one of " + strings.Join(availableRoleNames, ", "))
		}

		return errors.HandleCommon(err, cmd)
	}

	var data [][]string
	roleDisplay, err := createPrettyRole(details)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	data = append(data, printer.ToRow(roleDisplay, roleFields))
	outputTable(data)
	return nil
}

func createPrettyRole(role mds.Role) (*prettyRole, error) {
	marshalled, err := json.Marshal(role.AccessPolicy)
	if err != nil {
		return nil, err
	}
	return &prettyRole{
		role.Name,
		string(pretty.Pretty(marshalled)),
	}, nil
}

func outputTable(data [][]string) {
	tablePrinter := tablewriter.NewWriter(os.Stdout)
	tablePrinter.SetAutoWrapText(false)
	tablePrinter.SetAutoFormatHeaders(false)
	tablePrinter.SetHeader(roleLabels)
	tablePrinter.AppendBulk(data)
	tablePrinter.SetBorder(false)
	tablePrinter.Render()
}
