package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func NewRBACCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rbac",
		Short: "Manage RBAC permissions.",
		Long:  "Manage Role-Based Access Control (RBAC) permissions.",
	}

	cmd.AddCommand(NewRoleCommand(cfg, prerunner))
	cmd.AddCommand(NewRolebindingCommand(cfg, prerunner))

	return cmd
}
