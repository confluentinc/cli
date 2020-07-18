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
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	roleFields = []string{"Name", "AccessPolicy"}
	roleLabels = []string{"Name", "AccessPolicy"}
)

type roleCommand struct {
	*cmd.AuthenticatedCLICommand
	cliName string
}

type prettyRole struct {
	Name         string
	AccessPolicy string
}

// NewRoleCommand returns the sub-command object for interacting with RBAC roles.
func NewRoleCommand(cliName string, prerunner cmd.PreRunner) *cobra.Command {
	cliCmd := cmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "role",
			Short: "Manage RBAC and IAM roles.",
			Long:  "Manage Role-Based Access Control (RBAC) and Identity and Access Management (IAM) roles.",
		}, prerunner)
	roleCmd := &roleCommand{
		AuthenticatedCLICommand: cliCmd,
		cliName:                 cliName,
	}
	roleCmd.init()
	return roleCmd.Command
}

func (c *roleCommand) createContext() context.Context {
	if c.cliName == "ccloud" {
		return context.WithValue(context.Background(), mdsv2alpha1.ContextAccessToken, c.State.AuthToken)
	} else {
		return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
	}
}

func (c *roleCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the available RBAC roles.",
		Args:  cobra.NoArgs,
		RunE:  cmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the resources and operations allowed for a role.",
		Args:  cobra.ExactArgs(1),
		RunE:  cmd.NewCLIRunE(c.describe),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)
}

func (c *roleCommand) confluentList(cmd *cobra.Command) error {
	roles, _, err := c.MDSClient.RBACRoleDefinitionsApi.Roles(c.createContext())
	if err != nil {
		return err
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	if format == output.Human.String() {
		var data [][]string
		for _, role := range roles {
			roleDisplay, err := createPrettyRole(role)
			if err != nil {
				return err
			}
			data = append(data, printer.ToRow(roleDisplay, roleFields))
		}
		outputTable(data)
	} else {
		return output.StructuredOutput(format, roles)
	}
	return nil
}

func (c *roleCommand) ccloudList(cmd *cobra.Command) error {
	rolesV2, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Roles(c.createContext())
	if err != nil {
		return err
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	if format == output.Human.String() {
		var data [][]string
		for _, role := range rolesV2 {
			roleDisplay, err := createPrettyRoleV2(role)
			if err != nil {
				return err
			}
			data = append(data, printer.ToRow(roleDisplay, roleFields))
		}
		outputTable(data)
	} else {
		return output.StructuredOutput(format, rolesV2)
	}
	return nil
}

func (c *roleCommand) list(cmd *cobra.Command, _ []string) error {
	if c.cliName == "ccloud" {
		return c.ccloudList(cmd)
	} else {
		return c.confluentList(cmd)
	}
}

func (c *roleCommand) confluentDescribe(cmd *cobra.Command, role string) error {
	details, r, err := c.MDSClient.RBACRoleDefinitionsApi.RoleDetail(c.createContext(), role)
	if err != nil {
		if r.StatusCode == http.StatusNoContent {
			availableRoleNames, _, err := c.MDSClient.RBACRoleDefinitionsApi.Rolenames(c.createContext())
			if err != nil {
				return err
			}
			suggestionsMsg := fmt.Sprintf(errors.UnknownRoleSuggestions, strings.Join(availableRoleNames, ","))
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.UnknownRoleErrorMsg, role), suggestionsMsg)
		}

		return err
	}

	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if format == output.Human.String() {
		var data [][]string
		roleDisplay, err := createPrettyRole(details)
		if err != nil {
			return err
		}
		data = append(data, printer.ToRow(roleDisplay, roleFields))
		outputTable(data)
	} else {
		return output.StructuredOutput(format, details)
	}

	return nil
}

func (c *roleCommand) ccloudDescribe(cmd *cobra.Command, role string) error {
	details, r, err := c.MDSv2Client.RBACRoleDefinitionsApi.RoleDetail(c.createContext(), role)
	if err != nil {
		if r.StatusCode == http.StatusNotFound {
			availableRoleNames, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Rolenames(c.createContext())
			if err != nil {
				return err
			}

			suggestionsMsg := fmt.Sprintf(errors.UnknownRoleSuggestions, strings.Join(availableRoleNames, ","))
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.UnknownRoleErrorMsg, role), suggestionsMsg)
		}

		return err
	}

	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if format == output.Human.String() {
		var data [][]string
		roleDisplay, err := createPrettyRoleV2(details)
		if err != nil {
			return err
		}
		data = append(data, printer.ToRow(roleDisplay, roleFields))
		outputTable(data)
	} else {
		return output.StructuredOutput(format, details)
	}

	return nil
}

func (c *roleCommand) describe(cmd *cobra.Command, args []string) error {
	role := args[0]

	if c.cliName == "ccloud" {
		return c.ccloudDescribe(cmd, role)
	} else {
		return c.confluentDescribe(cmd, role)
	}
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

func createPrettyRoleV2(role mdsv2alpha1.Role) (*prettyRole, error) {
	marshalled, err := json.Marshal(role.Policies)
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
