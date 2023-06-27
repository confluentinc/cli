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
	pcmd.RegisterFlagCompletionFunc(cmd, "compatibility", func(_ *cobra.Command, _ []string) []string {
		return compatibilities
	})
}

func addCompatibilityGroupFlag(cmd *cobra.Command) {
	cmd.Flags().String("compatibility-group", "", "The name of the compatibility group.")
}

func addMetadataDefaultsFlag(cmd *cobra.Command) {
	cmd.Flags().String("metadata-defaults", "", "The path to the schema metadata defaults file.")
	cobra.CheckErr(cmd.MarkFlagFilename("metadata-defaults", "json"))
}

func addMetadataOverridesFlag(cmd *cobra.Command) {
	cmd.Flags().String("metadata-overrides", "", "The path to the schema metadata overrides file.")
	cobra.CheckErr(cmd.MarkFlagFilename("metadata-overrides", "json"))
}

func addRulesetDefaultsFlag(cmd *cobra.Command) {
	cmd.Flags().String("ruleset-defaults", "", "The path to the schema ruleset defaults file.")
	cobra.CheckErr(cmd.MarkFlagFilename("ruleset-defaults", "json"))
}

func addRulesetOverridesFlag(cmd *cobra.Command) {
	cmd.Flags().String("ruleset-overrides", "", "The path to the schema ruleset overrides file.")
	cobra.CheckErr(cmd.MarkFlagFilename("ruleset-overrides", "json"))
}

func addModeFlag(cmd *cobra.Command) {
	modes := []string{"readwrite", "readonly", "import"}
	cmd.Flags().String("mode", "", fmt.Sprintf("Can be %s.", utils.ArrayToCommaDelimitedString(modes, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "mode", func(_ *cobra.Command, _ []string) []string { return modes })
}
