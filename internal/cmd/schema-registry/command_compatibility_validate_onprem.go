package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func (c *compatibilityCommand) newValidateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "validate",
		Short:       "Validate a schema with a subject version.",
		Long:        "Validate that a schema is compatible against a given subject version.",
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
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version or "latest".`)
	cmd.Flags().String("schema", "", "The path to the schema file.")
	pcmd.AddSchemaTypeFlag(cmd)
	cmd.Flags().String("refs", "", "The path to the references file.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *compatibilityCommand) onPremValidate(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return validateSchemaCompatibility(cmd, srClient, ctx)
}
