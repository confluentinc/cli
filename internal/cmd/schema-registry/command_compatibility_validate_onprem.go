package schemaregistry

import (
	"fmt"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *compatibilityCommand) newValidateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "validate",
		Short:       "Validate input schema against a particular version of a subject for compatibility.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.onPremValidate),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Validate the compatibility of schema `payments` against the latest version of subject `records`.",
				Code: fmt.Sprintf("%s schema-registry compatibility validate --schema payments.avro --type AVRO --subject records --version latest %s", pversion.CLIName, OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().String("subject", "", SubjectUsage)
	cmd.Flags().String("version", "", "Version of the schema. Can be a specific version or 'latest'.")
	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().String("type", "", `Specify the schema type as "avro", "protobuf", or "jsonschema".`)
	cmd.Flags().String("refs", "", "The path to the references file.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *compatibilityCommand) onPremValidate(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetAPIClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return validateSchemaCompatibility(cmd, srClient, ctx)
}
