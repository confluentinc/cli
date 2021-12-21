package schemaregistry

import (
	"fmt"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *compatibilityCommand) newValidateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <subject>",
		Short: "Validate input schema against a particular version of a subject for compatibility.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.onPremValidate),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Validate the compatibility of <schema> against a subject of latest version.",
				Code: fmt.Sprintf("%s schema-registry compatibility validate --subject <subject-name> --version latest --schema <schema-path>  %s", pversion.CLIName, errors.OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().StringP("version", "V", "", "Version of the schema. Can be a specific version or 'latest'.")
	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().String("type", "", `Specify the schema type as "avro", "protobuf", or "jsonschema".`)
	cmd.Flags().String("refs", "", "The path to the references file.")
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
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
