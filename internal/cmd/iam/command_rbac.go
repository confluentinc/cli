package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

func NewRBACCommand(cfg *v3.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rbac",
		Short: "Manage RBAC permissions.",
		Long:  "Manage Role-Based Access Control (RBAC) permissions.",
	}

	cmd.AddCommand(NewRoleCommand(cfg, prerunner))
	cmd.AddCommand(NewRolebindingCommand(cfg, prerunner))

	return cmd
}
