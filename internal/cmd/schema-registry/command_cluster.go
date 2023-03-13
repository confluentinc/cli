package schemaregistry

import (
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	srClient *srsdk.APIClient
}

func newClusterCommand(cfg *v1.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	if cfg.IsCloudLogin() {
		cmd.Short = "Manage Schema Registry cluster."
		cmd.Long = "Manage the Schema Registry cluster for the current environment."
	} else {
		cmd.Short = "Manage Schema Registry clusters."
	}

	c := &clusterCommand{srClient: srClient}

	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
	}

	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand(cfg))
	cmd.AddCommand(c.newEnableCommand(cfg))
	cmd.AddCommand(c.newListCommandOnPrem())
	cmd.AddCommand(c.newUpdateCommand(cfg))
	cmd.AddCommand(c.newUpgradeCommand())

	return cmd
}
