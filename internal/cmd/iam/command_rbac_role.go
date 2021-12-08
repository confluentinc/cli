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

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	roleFields = []string{"Name", "AccessPolicy"}
	roleLabels = []string{"Name", "AccessPolicy"}
)

type roleCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	cfg                        *v1.Config
	ccloudRbacDataplaneEnabled bool
}

type prettyRole struct {
	Name         string
	AccessPolicy string
}

// NewRoleCommand returns the sub-command object for interacting with RBAC roles.
func NewRoleCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cobraRoleCmd := &cobra.Command{
		Use:   "role",
		Short: "Manage RBAC and IAM roles.",
		Long:  "Manage Role-Based Access Control (RBAC) and Identity and Access Management (IAM) roles.",
	}
	var cliCmd *pcmd.AuthenticatedStateFlagCommand
	if cfg.IsOnPremLogin() {
		cliCmd = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cobraRoleCmd, prerunner, RoleSubcommandFlags)
	} else {
		cliCmd = pcmd.NewAuthenticatedStateFlagCommand(cobraRoleCmd, prerunner, nil)
	}
	ccloudRbacDataplaneEnabled := false
	if os.Getenv("XX_CCLOUD_RBAC_DATAPLANE") != "" {
		ccloudRbacDataplaneEnabled = true
	}
	roleCmd := &roleCommand{
		AuthenticatedStateFlagCommand: cliCmd,
		cfg:                           cfg,
		ccloudRbacDataplaneEnabled:    ccloudRbacDataplaneEnabled,
	}
	roleCmd.init()
	return roleCmd.Command
}

func (c *roleCommand) createContext() context.Context {
	if c.cfg.IsCloudLogin() {
		return context.WithValue(context.Background(), mdsv2alpha1.ContextAccessToken, c.State.AuthToken)
	} else {
		return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
	}
}

func (c *roleCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the available RBAC roles.",
		Long:  "List the available RBAC roles and associated information, such as the resource types and operations that the role has permission to perform.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	pcmd.AddOutputFlag(listCmd)
	c.AddCommand(listCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the resources and operations allowed for a role.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
	}
	pcmd.AddOutputFlag(describeCmd)
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
	roles := mdsv2alpha1.RolesOpts{}
	if c.ccloudRbacDataplaneEnabled {
		roles.Namespace = dataplaneNamespace
	}
	// Currently we don't allow multiple namespace in roles so as a workaround we first check with dataplane
	// namespace and if we get an error try without any namespace.
	rolesV2, r, err := c.MDSv2Client.RBACRoleDefinitionsApi.Roles(c.createContext(), &roles)
	if err != nil || r.StatusCode == http.StatusNoContent {
		rolesV2, _, err = c.MDSv2Client.RBACRoleDefinitionsApi.Roles(c.createContext(), nil)
		if err != nil {
			return err
		}
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
	if c.cfg.IsCloudLogin() {
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
	roleDetail := mdsv2alpha1.RoleDetailOpts{}
	if c.ccloudRbacDataplaneEnabled {
		roleDetail.Namespace = dataplaneNamespace
	}
	// Currently we don't allow multiple namespace in roleDetail so as a workaround we first check with dataplane
	// namespace and if we get an error try without any namespace.
	details, r, err := c.MDSv2Client.RBACRoleDefinitionsApi.RoleDetail(c.createContext(), role, &roleDetail)
	if err != nil || r.StatusCode == http.StatusNoContent {
		details, r, err = c.MDSv2Client.RBACRoleDefinitionsApi.RoleDetail(c.createContext(), role, nil)
		if err != nil {
			if r.StatusCode == http.StatusNotFound {
				availableRoleNames, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Rolenames(c.createContext(), nil)
				if err != nil {
					return err
				}

				suggestionsMsg := fmt.Sprintf(errors.UnknownRoleSuggestions, strings.Join(availableRoleNames, ","))
				return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.UnknownRoleErrorMsg, role), suggestionsMsg)
			}

			return err
		}
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

	if c.cfg.IsCloudLogin() {
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
