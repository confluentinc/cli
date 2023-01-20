package iam

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"
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
		RunE:  c.describe,
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
	opts := &mdsv2alpha1.RoleDetailOpts{Namespace: dataplaneNamespace}

	// Currently we don't allow multiple namespace in opts so as a workaround we first check with dataplane
	// namespace and if we get an error try without any namespace.
	details, r, err := c.MDSv2Client.RBACRoleDefinitionsApi.RoleDetail(c.createContext(), role, opts)
	if err != nil || r.StatusCode == http.StatusNoContent {
		details, r, err = c.MDSv2Client.RBACRoleDefinitionsApi.RoleDetail(c.createContext(), role, nil)
		if err != nil {
			if r.StatusCode == http.StatusNotFound {
				publicRolenames, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Rolenames(c.createContext(), nil)
				if err != nil {
					return err
				}

				opts := &mdsv2alpha1.RolenamesOpts{Namespace: dataplaneNamespace}
				dataplaneRolenames, _, _ := c.MDSv2Client.RBACRoleDefinitionsApi.Rolenames(c.createContext(), opts)
				rolenames := append(publicRolenames, dataplaneRolenames...)

				suggestionsMsg := fmt.Sprintf(errors.UnknownRoleSuggestions, strings.Join(rolenames, ", "))
				return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.UnknownRoleErrorMsg, role), suggestionsMsg)
			}

			return err
		}
	}

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, details)
	}

	roleDisplay, err := createPrettyRoleV2(details)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(roleDisplay)
	return table.PrintWithAutoWrap(false)
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

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, details)
	}

	roleDisplay, err := createPrettyRole(details)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	list.Add(roleDisplay)
	return list.PrintWithAutoWrap(false)
}
