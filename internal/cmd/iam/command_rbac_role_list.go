package iam

import (
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *roleCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the available RBAC roles.",
		Long:  "List the available RBAC roles and associated information, such as the resource types and operations that the role has permission to perform.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)
	if c.cfg.IsOnPremLogin() {
		pcmd.AddContextFlag(cmd, c.CLICommand)
	}

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
	// add public, dataplane, datagovernance, ksql, and streamcatalog roles
	namespaces := []string{
		dataplaneNamespace.Value(),
		dataGovernanceNamespace.Value(),
		ksqlNamespace.Value(),
		publicNamespace.Value(),
		streamCatalogNamespace.Value(),
	}
	opt := optional.NewString(strings.Join(namespaces, ","))
	roles, err := c.namespaceRoles(opt)
	if err != nil {
		return err
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
