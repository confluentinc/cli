package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

const (
	configFileFlagName = "config-file"
	dryrunFlagName     = "dry-run"
)

type linkCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newLinkCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "link",
		Short:       "Manage inter-cluster links.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &linkCommand{}

	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)

		c.AddCommand(c.newConfigurationCommand(cfg))
		c.AddCommand(c.newCreateCommand())
		c.AddCommand(c.newDeleteCommand())
		c.AddCommand(c.newListCommand())
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
		c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

		c.AddCommand(c.newConfigurationCommand(cfg))
		c.AddCommand(c.newCreateCommandOnPrem())
		c.AddCommand(c.newDeleteCommandOnPrem())
		c.AddCommand(c.newListCommandOnPrem())
	}

	return c.Command
}
