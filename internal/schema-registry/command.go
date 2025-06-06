package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schema-registry",
		Aliases: []string{"sr"},
		Short:   "Manage Schema Registry.",
	}

	c := &command{}
	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
	}

	cmd.AddCommand(c.newClusterCommand(cfg))
	cmd.AddCommand(c.newConfigurationCommand(cfg))
	cmd.AddCommand(c.newDekCommand(cfg))
	cmd.AddCommand(c.newEndpointsCommand())
	cmd.AddCommand(c.newExporterCommand(cfg))
	cmd.AddCommand(c.newKekCommand(cfg))
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

func addCaLocationAndClientPathFlags(cmd *cobra.Command) {
	cmd.Flags().String("certificate-authority-path", "", "File or directory path to Certificate Authority certificates to authenticate the Schema Registry client.")
	cmd.Flags().String("client-cert-path", "", "File or directory path to client certificate to authenticate the Schema Registry client.")
	cmd.Flags().String("client-key-path", "", "File or directory path to client key to authenticate the Schema Registry client.")
	cmd.MarkFlagsRequiredTogether("client-cert-path", "client-key-path")
}

func addSchemaRegistryEndpointFlag(cmd *cobra.Command) {
	cmd.Flags().String("schema-registry-endpoint", "", "The URL of the Schema Registry cluster.")
}
