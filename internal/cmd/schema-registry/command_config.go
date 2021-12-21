package schemaregistry

import (
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type configCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	srClient *srsdk.APIClient
}

func newConfigCommand(cfg *v1.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Schema Registry config.",
	}

	c := &configCommand{
		srClient: srClient,
	}
	if cfg.IsCloudLogin() {
		cmd.Annotations = map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin}
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner, SchemaSubcommandFlags)
	} else {
		cmd.Annotations = map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin}
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner, nil)
	}

	if cfg.IsCloudLogin() {
		c.AddCommand(c.newDescribeCommand())
	} else {
		c.AddCommand(c.newDescribeCommandOnPrem())
	}

	return c.Command
}
