package iam

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
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
	// check for role in all namespaces
	namespacesList := []string{
		dataplaneNamespace.Value(),
		dataGovernanceNamespace.Value(),
		ksqlNamespace.Value(),
		publicNamespace.Value(),
		streamCatalogNamespace.Value(),
	}

	// check if IdentityAdmin is enabled
	ldClient := v1.GetCcloudLaunchDarklyClient(c.Context.PlatformName)
	if featureflags.Manager.BoolVariation("auth.rbac.identity_admin.enable", c.Context, ldClient, true, false) {
		namespacesList = append(namespacesList, identityNamespace.Value())
	}

	if featureflags.Manager.BoolVariation("flink.rbac.namespace.cli.enable", c.Context, ldClient, true, false) {
		namespacesList = append(namespacesList, flinkNamespace.Value())
	}

	namespaces := optional.NewString(strings.Join(namespacesList, ","))

	opts := &mdsv2alpha1.RoleDetailOpts{Namespace: namespaces}

	details, r, err := c.MDSv2Client.RBACRoleDefinitionsApi.RoleDetail(c.createContext(), role, opts)

	if err != nil || r.StatusCode == http.StatusNotFound {
		opts := &mdsv2alpha1.RolenamesOpts{Namespace: namespaces}
		roleNames, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Rolenames(c.createContext(), opts)
		if err != nil {
			return err
		}

		suggestionsMsg := fmt.Sprintf(errors.UnknownRoleSuggestions, strings.Join(roleNames, ", "))
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.UnknownRoleErrorMsg, role), suggestionsMsg)
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
