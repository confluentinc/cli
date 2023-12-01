package schemaregistry

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/schemaregistry"
)

type validateOut struct {
	IsCompatible bool `human:"Compatible" serialized:"is_compatible"`
}

func (c *command) newCompatibilityValidateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a schema with a subject version.",
		Long:  "Validate that a schema is compatible against a given subject version.",
		Args:  cobra.NoArgs,
		RunE:  c.compatibilityValidate,
	}

	example := examples.Example{
		Text: `Validate the compatibility of schema "payments" against the latest version of subject "records".`,
		Code: "confluent schema-registry compatibility validate --schema payments.avsc --type avro --subject records --version latest",
	}
	if cfg.IsOnPremLogin() {
		example.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example)

	cmd.Flags().String("schema", "", "The path to the schema file.")
	pcmd.AddSchemaTypeFlag(cmd)
	cmd.Flags().String("subject", "", subjectUsage)
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version or "latest".`)
	cmd.Flags().String("references", "", "The path to the references file.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	if cfg.IsCloudLogin() {
		// Deprecated
		pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-key"))

		// Deprecated
		pcmd.AddApiSecretFlag(cmd)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-secret"))
	}

	cobra.CheckErr(cmd.MarkFlagFilename("schema", "avsc", "json", "proto"))
	cobra.CheckErr(cmd.MarkFlagFilename("references", "json"))

	return cmd
}

func (c *command) compatibilityValidate(cmd *cobra.Command, _ []string) error {
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

	references, err := cmd.Flags().GetString("references")
	if err != nil {
		return err
	}
	refs, err := schemaregistry.ReadSchemaReferences(references)
	if err != nil {
		return err
	}

	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	req := srsdk.RegisterSchemaRequest{
		Schema:     srsdk.PtrString(string(schema)),
		SchemaType: srsdk.PtrString(schemaType),
		References: &refs,
	}

	res, err := client.TestCompatibilityBySubjectName(subject, version, req)
	if err != nil {
		return catchSchemaNotFoundError(err, subject, version)
	}

	table := output.NewTable(cmd)
	table.Add(&validateOut{IsCompatible: res.GetIsCompatible()})
	return table.Print()
}
