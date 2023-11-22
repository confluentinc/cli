package iam

import (
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/featureflags"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *roleCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the available RBAC roles.",
		Long:  "List the available RBAC roles and associated information, such as the resource types and operations that the role has permission to perform.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	if c.cfg.IsOnPremLogin() {
		pcmd.AddContextFlag(cmd, c.CLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *roleCommand) list(cmd *cobra.Command, _ []string) error {
	if c.cfg.IsCloudLogin() {
		return c.ccloudList(cmd)
	} else {
		return c.confluentList(cmd)
	}
}

func (c *roleCommand) ccloudList(cmd *cobra.Command) error {
	// add roles from all publicly released namespaces
	namespaces := []string{
		dataplaneNamespace.Value(),
		dataGovernanceNamespace.Value(),
		identityNamespace.Value(),
		ksqlNamespace.Value(),
		publicNamespace.Value(),
		streamCatalogNamespace.Value(),
	}
	opt := optional.NewString(strings.Join(namespaces, ","))
	roles, err := c.namespaceRoles(opt)
	if err != nil {
		return err
	}

	ldClient := featureflags.GetCcloudLaunchDarklyClient(c.Context.PlatformName)
	if featureflags.Manager.BoolVariation("flink.rbac.namespace.cli.enable", c.Context, ldClient, true, false) {
		flinkRoles, err := c.namespaceRoles(flinkNamespace)
		if err != nil {
			return err
		}
		roles = append(roles, flinkRoles...)
		workloadRoles, err := c.namespaceRoles(workloadNamespace)
		if err != nil {
			return err
		}
		roles = append(roles, workloadRoles...)
	}

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, roles)
	}

	list := output.NewList(cmd)
	for _, role := range roles {
		roleDisplay, err := createPrettyRoleV2(role)
		if err != nil {
			return err
		}
		list.Add(roleDisplay)
	}
	return list.PrintWithAutoWrap(false)
}

func (c *roleCommand) confluentList(cmd *cobra.Command) error {
	roles, _, err := c.MDSClient.RBACRoleDefinitionsApi.Roles(c.createContext())
	if err != nil {
		return err
	}

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, roles)
	}

	list := output.NewList(cmd)
	for _, role := range roles {
		roleDisplay, err := createPrettyRole(role)
		if err != nil {
			return err
		}
		list.Add(roleDisplay)
	}
	return list.PrintWithAutoWrap(false)
}
