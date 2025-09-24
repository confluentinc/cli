package iam

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
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
		pcmd.AddMDSOnPremMTLSFlags(cmd)
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
	// Get all production public roles by passing in default empty namespace
	roles, err := c.namespaceRoles(optional.EmptyString())
	if err != nil {
		return err
	}

	usmRoles, err := c.namespaceRoles(optional.NewString(usmNamespace.Value()))
	if err != nil {
		return err
	}
	roles = append(roles, usmRoles...)

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
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	roles, _, err := client.RBACRoleDefinitionsApi.Roles(c.createContext())
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
