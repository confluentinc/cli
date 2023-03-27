package iam

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

var (
	publicNamespace         = optional.NewString("public")
	dataGovernanceNamespace = optional.NewString("datagovernance")
	dataplaneNamespace      = optional.NewString("dataplane")
	ksqlNamespace           = optional.NewString("ksql")
	streamCatalogNamespace  = optional.NewString("streamcatalog")
	identityNamespace       = optional.NewString("identity")
)

func newRBACCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rbac",
		Short: "Manage RBAC permissions.",
		Long:  "Manage Role-Based Access Control (RBAC) permissions.",
	}

	cmd.AddCommand(newRoleCommand(cfg, prerunner))
	cmd.AddCommand(newRoleBindingCommand(cfg, prerunner))

	return cmd
}

func (c *roleCommand) namespaceRoles(namespace optional.String) ([]mdsv2alpha1.Role, error) {
	opts := &mdsv2alpha1.RolesOpts{Namespace: namespace}
	roles, _, err := c.MDSv2Client.RBACRoleDefinitionsApi.Roles(c.createContext(), opts)
	return roles, err
}
