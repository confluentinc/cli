package schemaregistry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/schemaregistry"
)

func (c *command) newSchemaCreateCommand(cfg *config.Config) *cobra.Command {
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
		example.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(
		example,
		examples.Example{
			Text: `Where "employee.avsc" may include the following content:`,
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
			Text: "For more information on schema types and references, see https://docs.confluent.io/platform/current/schema-registry/fundamentals/serdes-develop/index.html.",
		},
	)

	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().String("subject", "", subjectUsage)
	pcmd.AddSchemaTypeFlag(cmd)
	cmd.Flags().String("references", "", "The path to the references file.")
	cmd.Flags().String("metadata", "", "The path to metadata file.")
	cmd.Flags().String("ruleset", "", "The path to schema ruleset file.")
	cmd.Flags().Bool("normalize", false, "Alphabetize the list of schema fields.")
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

	references, err := cmd.Flags().GetString("references")
	if err != nil {
		return err
	}
	refs, err := schemaregistry.ReadSchemaReferences(references)
	if err != nil {
		return err
	}

	normalize, err := cmd.Flags().GetBool("normalize")
	if err != nil {
		return err
	}

	cfg := &schemaregistry.RegisterSchemaConfigs{
		Subject:    subject,
		SchemaType: schemaType,
		SchemaPath: schema,
		Refs:       refs,
		Normalize:  normalize,
	}

	if !c.Config.IsCloudLogin() {
		dir, err := createTempDir()
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

	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	id, err := schemaregistry.RegisterSchemaWithAuth(cfg, client)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd).IsSerialized() {
		if err := output.SerializedOutput(cmd, &schemaregistry.RegisterSchemaResponse{Id: id}); err != nil {
			return err
		}
	} else {
		output.Printf(c.Config.EnableColor, "Successfully registered schema with ID \"%d\".\n", id)
	}

	return nil
}

func createTempDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "ccloud-schema")
	err := os.MkdirAll(dir, 0755)
	return dir, err
}

func read(path string, v any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	return json.NewDecoder(file).Decode(v)
}
