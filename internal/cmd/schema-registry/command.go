package schemaregistry

import (
	"fmt"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func New(cfg *v1.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "schema-registry",
		Aliases:     []string{"sr"},
		Short:       "Manage Schema Registry.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := pcmd.NewAuthenticatedCLICommand(cmd, prerunner)

	c.AddCommand(newClusterCommand(cfg, prerunner, srClient))
	c.AddCommand(newCompatibilityCommand(cfg, prerunner, srClient))
	c.AddCommand(newConfigCommand(cfg, prerunner, srClient))
	c.AddCommand(newExporterCommand(cfg, prerunner, srClient))
	c.AddCommand(newSchemaCommand(cfg, prerunner, srClient))
	c.AddCommand(newSubjectCommand(cfg, prerunner, srClient))
	return c.Command
}

func addCompatibilityFlag(cmd *cobra.Command) {
	compatabilities := []string{"backward", "backward_transitive", "forward", "forward_transitive", "full", "full_transitive", "none"}
	cmd.Flags().String("compatibility", "", fmt.Sprintf("Can be %s.", utils.ArrayToCommaDelimitedString(compatabilities)))
	pcmd.RegisterFlagCompletionFunc(cmd, "compatibility", func(_ *cobra.Command, _ []string) []string {
		return compatabilities
	})
}

func addModeFlag(cmd *cobra.Command) {
	modes := []string{"readwrite", "readonly", "import"}
	cmd.Flags().String("mode", "", fmt.Sprintf("Can be %s.", utils.ArrayToCommaDelimitedString(modes)))
	pcmd.RegisterFlagCompletionFunc(cmd, "mode", func(_ *cobra.Command, _ []string) []string { return modes })
}
