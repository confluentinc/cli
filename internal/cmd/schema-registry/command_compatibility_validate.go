package schemaregistry

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
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
				Text: `Validate the compatibility of schema "payments" against the latest version of subject "records".`,
				Code: "confluent schema-registry compatibility validate --schema payments.avsc --type avro --subject records --version latest",
			},
		),
	}

	cmd.Flags().String("schema", "", "The path to the schema file.")
	pcmd.AddSchemaTypeFlag(cmd)
	cmd.Flags().String("subject", "", SubjectUsage)
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version or "latest".`)
	cmd.Flags().String("references", "", "The path to the references file.")
	if c.Config.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	// Deprecated
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	cobra.CheckErr(cmd.Flags().MarkHidden("api-key"))

	// Deprecated
	pcmd.AddApiSecretFlag(cmd)
	cobra.CheckErr(cmd.Flags().MarkHidden("api-secret"))

	cobra.CheckErr(cmd.MarkFlagFilename("schema", "avsc", "json", "proto"))
	cobra.CheckErr(cmd.MarkFlagFilename("references", "json"))

	return cmd
}

func (c *command) compatibilityValidate(cmd *cobra.Command, args []string) error {
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

	references, err := ReadSchemaReferences(cmd)
	if err != nil {
		return err
	}

	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	req := srsdk.RegisterSchemaRequest{
		Schema:     string(schema),
		SchemaType: schemaType,
		References: references,
	}

	res, err := client.TestCompatibilityBySubjectName(subject, version, req)
	if err != nil {
		return catchSchemaNotFoundError(err, subject, version)
	}

	table := output.NewTable(cmd)
	table.Add(&validateOut{IsCompatible: res.IsCompatible})
	return table.Print()
}
