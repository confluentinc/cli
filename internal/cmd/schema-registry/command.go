package schemaregistry

import (
	"fmt"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	srClient *srsdk.APIClient
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "schema-registry",
		Aliases:     []string{"sr"},
		Short:       "Manage Schema Registry.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &command{srClient: srClient}
	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
	}

	cmd.AddCommand(c.newClusterCommand(cfg))
	cmd.AddCommand(c.newCompatibilityCommand(cfg))
	cmd.AddCommand(c.newConfigCommand(cfg))
	cmd.AddCommand(c.newExporterCommand(cfg))
	cmd.AddCommand(c.newSchemaCommand(cfg))
	cmd.AddCommand(c.newSubjectCommand(cfg))

	return cmd
}

func addCompatibilityFlag(cmd *cobra.Command) {
	compatibilities := []string{"backward", "backward_transitive", "forward", "forward_transitive", "full", "full_transitive", "none"}
	cmd.Flags().String("compatibility", "", fmt.Sprintf("Can be %s.", utils.ArrayToCommaDelimitedString(compatibilities, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "compatibility", func(_ *cobra.Command, _ []string) []string {
		return compatibilities
	})
}

func addModeFlag(cmd *cobra.Command) {
	modes := []string{"readwrite", "readonly", "import"}
	cmd.Flags().String("mode", "", fmt.Sprintf("Can be %s.", utils.ArrayToCommaDelimitedString(modes, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "mode", func(_ *cobra.Command, _ []string) []string { return modes })
}
