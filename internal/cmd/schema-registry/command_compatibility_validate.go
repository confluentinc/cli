package schemaregistry

import (
	"context"
	"fmt"
	"os"
	"strings"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

type validateOut struct {
	IsCompatible bool `human:"Compatible" serialized:"is_compatible"`
}

func (c *command) newCompatibilityValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a schema with a subject version.",
		Long:  "Validate that a schema is compatible against a given subject version.",
		Args:  cobra.NoArgs,
		RunE:  c.compatibilityValidate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Validate the compatibility of schema `payments` against the latest version of subject `records`.",
				Code: fmt.Sprintf("%s schema-registry compatibility validate --schema payments.avro --type avro --subject records --version latest", pversion.CLIName),
			},
		),
	}

	cmd.Flags().String("subject", "", SubjectUsage)
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version or "latest".`)
	cmd.Flags().String("schema", "", "The path to the schema file.")
	pcmd.AddSchemaTypeFlag(cmd)
	cmd.Flags().String("references", "", "The path to the references file.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) compatibilityValidate(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return validateSchemaCompatibility(cmd, srClient, ctx)
}

func validateSchemaCompatibility(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	schemaPath, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}

	schemaType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	schemaType = strings.ToUpper(schemaType)

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	refs, err := ReadSchemaRefs(cmd)
	if err != nil {
		return err
	}

	req := srsdk.RegisterSchemaRequest{Schema: string(schema), SchemaType: schemaType, References: refs}

	compatibilityCheck, httpResp, err := srClient.DefaultApi.TestCompatibilityBySubjectName(ctx, subject, version, req, nil)
	if err != nil {
		return errors.CatchSchemaNotFoundError(err, httpResp)
	}

	table := output.NewTable(cmd)
	table.Add(&validateOut{IsCompatible: compatibilityCheck.IsCompatible})
	return table.Print()
}
