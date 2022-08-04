package iam

import (
	"os"
	"strings"

	"github.com/antihax/optional"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
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
	var roles []mdsv2alpha1.Role

	// add public and dataplane roles
	publicAndDataplaneNamespace := []string{publicNamespace.Value(), dataplaneNamespace.Value()}
	publicAndDataplaneNamespaceOpt := optional.NewString(strings.Join(publicAndDataplaneNamespace, ","))
	publicAndDataplaneRoles, err := c.namespaceRoles(publicAndDataplaneNamespaceOpt)
	if err != nil {
		return err
	}
	roles = append(roles, publicAndDataplaneRoles...)

	// add ksql and datagovernance roles
	if os.Getenv("XX_DATAPLANE_3_ENABLE") != "" {
		ksqlAndDataGovernanceNamespace := []string{ksqlNamespace.Value(), dataGovernanceNamespace.Value()}
		ksqlAndDataGovernanceNamespaceOpt := optional.NewString(strings.Join(ksqlAndDataGovernanceNamespace, ","))
		ksqlAndDataGovernanceRoles, err := c.namespaceRoles(ksqlAndDataGovernanceNamespaceOpt)
		if err != nil {
			return err
		}
		roles = append(roles, ksqlAndDataGovernanceRoles...)
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
