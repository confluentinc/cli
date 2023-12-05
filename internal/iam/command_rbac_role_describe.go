package iam

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/featureflags"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

const (
	unknownRoleErrorMsg    = `unknown role "%s"`
	unknownRoleSuggestions = "The available roles are %s."
)

func (c *roleCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the resources and operations allowed for a role.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	if c.cfg.IsOnPremLogin() {
		pcmd.AddContextFlag(cmd, c.CLICommand)
	}
	pcmd.AddOutputFlag(cmd)

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
		identityNamespace.Value(),
		ksqlNamespace.Value(),
		publicNamespace.Value(),
		streamCatalogNamespace.Value(),
	}

	ldClient := featureflags.GetCcloudLaunchDarklyClient(c.Context.PlatformName)
	if featureflags.Manager.BoolVariation("flink.rbac.namespace.cli.enable", c.Context, ldClient, true, false) {
		namespacesList = append(namespacesList, flinkNamespace.Value(), workloadNamespace.Value())
	}

	namespaces := optional.NewString(strings.Join(namespacesList, ","))

	opts := &mdsv2alpha1.RoleDetailOpts{Namespace: namespaces}

	details, httpResp, err := c.MDSv2Client.RBACRoleDefinitionsApi.RoleDetail(c.createContext(), role, opts)
	if err != nil || httpResp.StatusCode == http.StatusNotFound {
		opts := &mdsv2alpha1.RolenamesOpts{Namespace: namespaces}
		roleNames, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Rolenames(c.createContext(), opts)
		if err != nil {
			return err
		}

		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(unknownRoleErrorMsg, role),
			fmt.Sprintf(unknownRoleSuggestions, utils.ArrayToCommaDelimitedString(roleNames, "and")),
		)
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
	details, httpResp, err := c.MDSClient.RBACRoleDefinitionsApi.RoleDetail(c.createContext(), role)
	if err != nil {
		if httpResp.StatusCode == http.StatusNoContent {
			availableRoleNames, _, err := c.MDSClient.RBACRoleDefinitionsApi.Rolenames(c.createContext())
			if err != nil {
				return err
			}
			return errors.NewErrorWithSuggestions(
				fmt.Sprintf(unknownRoleErrorMsg, role),
				fmt.Sprintf(unknownRoleSuggestions, utils.ArrayToCommaDelimitedString(availableRoleNames, "and")),
			)
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
