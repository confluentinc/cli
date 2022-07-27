package iam

import (
	"github.com/antihax/optional"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	"github.com/spf13/cobra"
	"os"

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
	roles := []mdsv2alpha1.Role{}

	// add public roles
	publicRoles, err := c.publicRoles()
	if err != nil {
		return err
	}
	roles = append(roles, publicRoles...)

	// add dataplane roles
	dataplaneRoles, err := c.dataplaneRoles()
	if err != nil {
		return err
	}
	roles = append(roles, dataplaneRoles...)

	// add ksql and datagovernance roles
	if os.Getenv("XX_DATAPLANE_3_ENABLE") == "1" {
		ksqlRoles, err := c.ksqlRoles()
		if err != nil {
			return err
		}
		roles = append(roles, ksqlRoles...)

		datagovernanceRoles, err := c.datagovernanceRoles()
		if err != nil {
			return err
		}
		roles = append(roles, datagovernanceRoles...)
	}

	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if format == output.Human.String() {
		var data [][]string
		for _, role := range roles {
			roleDisplay, err := createPrettyRoleV2(role)
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

func (c *roleCommand) namespaceRoles(namespace optional.String) ([]mdsv2alpha1.Role, error) {
	opts := &mdsv2alpha1.RolesOpts{Namespace: namespace}
	roles, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Roles(c.createContext(), opts)
	return roles, err
}

func (c *roleCommand) publicRoles() ([]mdsv2alpha1.Role, error) {
	roles, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Roles(c.createContext(), nil)
	return roles, err
}

func (c *roleCommand) dataplaneRoles() ([]mdsv2alpha1.Role, error) {
	return c.namespaceRoles(dataplaneNamespace)
}

func (c *roleCommand) ksqlRoles() ([]mdsv2alpha1.Role, error) {
	return c.namespaceRoles(ksqlNamespace)
}

func (c *roleCommand) datagovernanceRoles() ([]mdsv2alpha1.Role, error) {
	return c.namespaceRoles(datagovernanceNamespace)
}
