package schemaregistry

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *command) newSchemaCreateCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a schema.",
		Args:  cobra.NoArgs,
		RunE:  c.schemaCreate,
	}

	example := examples.Example{
		Text: "Register a new Avro schema.",
		Code: "confluent schema-registry schema create --subject employee --schema employee.avsc --type avro",
	}
	if cfg.IsOnPremLogin() {
		example.Code += " " + OnPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(
		example,
		examples.Example{
			Text: `Where "employee.avsc" may include these contents:`,
			Code: `{
	"type" : "record",
	"namespace" : "Example",
	"name" : "Employee",
	"fields" : [
		{ "name" : "Name" , "type" : "string" },
		{ "name" : "Age" , "type" : "int" }
	]
}`,
		},
		examples.Example{
			Text: "For more information on schema types and references, see https://docs.confluent.io/platform/current/schema-registry/fundamentals/serdes-develop/index.html",
		},
	)

	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().String("subject", "", SubjectUsage)
	pcmd.AddSchemaTypeFlag(cmd)
	cmd.Flags().String("references", "", "The path to the references file.")
	cmd.Flags().String("metadata", "", "The path to metadata file.")
	cmd.Flags().String("ruleset", "", "The path to schema ruleset file.")
	cmd.Flags().Bool("normalize", false, "Alphabetize the list of schema fields.")
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddContextFlag(cmd, c.CLICommand)
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
	cobra.CheckErr(cmd.MarkFlagFilename("metadata", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("ruleset", "json"))

	cobra.CheckErr(cmd.MarkFlagRequired("schema"))
	cobra.CheckErr(cmd.MarkFlagRequired("subject"))

	return cmd
}

func (c *command) schemaCreate(cmd *cobra.Command, _ []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	schema, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}

	schemaType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	schemaType = strings.ToUpper(schemaType)

	refs, err := ReadSchemaReferences(cmd)
	if err != nil {
		return err
	}

	normalize, err := cmd.Flags().GetBool("normalize")
	if err != nil {
		return err
	}

	cfg := &RegisterSchemaConfigs{
		Subject:    subject,
		SchemaType: schemaType,
		SchemaPath: schema,
		Refs:       refs,
		Normalize:  normalize,
	}

	if !c.Config.IsCloudLogin() {
		dir, err := CreateTempDir()
		if err != nil {
			return err
		}
		defer func() {
			_ = os.RemoveAll(dir)
		}()
		cfg.SchemaDir = dir
	}

	metadata, err := cmd.Flags().GetString("metadata")
	if err != nil {
		return err
	}
	if metadata != "" {
		cfg.Metadata = new(srsdk.Metadata)
		if err := read(metadata, cfg.Metadata); err != nil {
			return err
		}
	}

	ruleset, err := cmd.Flags().GetString("ruleset")
	if err != nil {
		return err
	}
	if ruleset != "" {
		cfg.Ruleset = new(srsdk.RuleSet)
		if err := read(ruleset, cfg.Ruleset); err != nil {
			return err
		}
	}

	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	if _, err := RegisterSchemaWithAuth(cmd, cfg, client); err != nil {
		return err
	}

	return nil
}

func read(path string, v any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	return json.NewDecoder(file).Decode(v)
}
