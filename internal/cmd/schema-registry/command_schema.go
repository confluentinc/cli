package schemaregistry

import (
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type schemaCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	srClient *srsdk.APIClient
}

func newSchemaCommand(cfg *v1.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "schema",
		Short:       "Manage Schema Registry schemas.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	c := &schemaCommand{
		srClient: srClient,
	}
	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
		c.AddCommand(c.newCreateCommand())
		c.AddCommand(c.newDescribeCommand())
		c.AddCommand(c.newDeleteCommand())
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
		c.AddCommand(c.newCreateCommandOnPrem())
		c.AddCommand(c.newDescribeCommandOnPrem())
		c.AddCommand(c.newDeleteCommandOnPrem())
	}
	return c.Command
}
