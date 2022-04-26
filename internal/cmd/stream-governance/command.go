package streamgovernance

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type streamGovernanceCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "stream-governance",
		Aliases:     []string{"sg"},
		Short:       "Manage Stream Governance",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := pcmd.NewAuthenticatedCLICommand(cmd, prerunner)
	sgCommand := &streamGovernanceCommand{}

	if cfg.IsCloudLogin() {
		sgCommand.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
	} else {
		sgCommand.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
	}

	c.AddCommand(sgCommand.newEnableCommand(cfg))
	c.AddCommand(sgCommand.newDescribeCommand(cfg))
	c.AddCommand(sgCommand.newUpgradeCommand(cfg))

	return c.Command
}
