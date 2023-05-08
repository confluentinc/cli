package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
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
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
	}

	cmd.AddCommand(c.newClusterCommand(cfg))
	cmd.AddCommand(c.newCompatibilityCommand(cfg))
	cmd.AddCommand(c.newConfigCommand(cfg))
	cmd.AddCommand(c.newExporterCommand(cfg))
	cmd.AddCommand(c.newRegionCommand())
	cmd.AddCommand(c.newSchemaCommand(cfg))
	cmd.AddCommand(c.newSubjectCommand(cfg))

	return cmd
}

func addCompatibilityFlag(cmd *cobra.Command) {
	compatibilities := []string{"backward", "backward_transitive", "forward", "forward_transitive", "full", "full_transitive", "none"}
	cmd.Flags().String("compatibility", "", fmt.Sprintf("Can be %s.", utils.ArrayToCommaDelimitedString(compatibilities, "or")))
	cmd.Flags().String("compatibility-group", "", "The name of the compatibility group.")
	cmd.Flags().String("metadata-defaults", "", "The path to the schema metadata defaults file.")
	cmd.Flags().String("metadata-overrides", "", "The path to the schema metadata overrides file.")
	cmd.Flags().String("ruleset-defaults", "", "The path to the schema ruleset defaults file.")
	cmd.Flags().String("ruleset-overrides", "", "The path to the schema ruleset overrides file.")

	pcmd.RegisterFlagCompletionFunc(cmd, "compatibility", func(_ *cobra.Command, _ []string) []string {
		return compatibilities
	})

	cobra.CheckErr(cmd.MarkFlagFilename("metadata-defaults", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("metadata-overrides", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("ruleset-defaults", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("ruleset-overrides", "json"))
}

func addModeFlag(cmd *cobra.Command) {
	modes := []string{"readwrite", "readonly", "import"}
	cmd.Flags().String("mode", "", fmt.Sprintf("Can be %s.", utils.ArrayToCommaDelimitedString(modes, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "mode", func(_ *cobra.Command, _ []string) []string { return modes })
}
