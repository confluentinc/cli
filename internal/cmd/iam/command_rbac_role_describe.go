package iam

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *roleCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the resources and operations allowed for a role.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
	}

	pcmd.AddOutputFlag(cmd)
	if c.cfg.IsOnPremLogin() {
		pcmd.AddContextFlag(cmd, c.CLICommand)
	}

	return cmd
}

func (c *roleCommand) describe(cmd *cobra.Command, args []string) error {
	role := args[0]

	if c.cfg.IsCloudLogin() {
		return c.ccloudDescribe(cmd, role)
	} else {
		return c.confluentDescribe(cmd, role)
	}
}

func (c *roleCommand) ccloudDescribe(cmd *cobra.Command, role string) error {
	roleDetail := mdsv2alpha1.RoleDetailOpts{Namespace: dataplaneNamespace}

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
