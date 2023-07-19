package schemaregistry

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type schemaCreateOut struct {
	Id int32 `human:"ID" serialized:"id"`
}

func (c *command) newSchemaCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a schema.",
		Args:  cobra.NoArgs,
		RunE:  c.schemaCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Register a new schema.",
				Code: "confluent schema-registry schema create --subject payments --schema payments.avro --type avro",
			},
			examples.Example{
				Text: "Where `schemafilepath` may include these contents:",
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
				Text: "For more information on schema types, see https://docs.confluent.io/current/schema-registry/serdes-develop/index.html.",
			},
			examples.Example{
				Text: "For more information on schema references, see https://docs.confluent.io/current/schema-registry/serdes-develop/index.html#schema-references.",
			},
		),
	}

	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().String("subject", "", SubjectUsage)
	pcmd.AddSchemaTypeFlag(cmd)
	cmd.Flags().String("references", "", "The path to the references file.")
	cmd.Flags().String("metadata", "", "The path to metadata file.")
	cmd.Flags().String("ruleset", "", "The path to schema ruleset file.")
	cmd.Flags().Bool("normalize", false, "Alphabetize the list of schema fields.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("schema", "avsc", "json", "proto"))
	cobra.CheckErr(cmd.MarkFlagFilename("references", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("metadata", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("ruleset", "json"))

	cobra.CheckErr(cmd.MarkFlagRequired("schema"))
	cobra.CheckErr(cmd.MarkFlagRequired("subject"))

	return cmd
}

func (c *command) schemaCreate(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	schemaPath, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	schemaType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	schemaType = strings.ToUpper(schemaType)

	refs, err := ReadSchemaRefs(cmd)
	if err != nil {
		return err
	}

	request := srsdk.RegisterSchemaRequest{
		Schema:     string(schema),
		SchemaType: schemaType,
		References: refs,
	}

	metadata, err := cmd.Flags().GetString("metadata")
	if err != nil {
		return err
	}
	if metadata != "" {
		request.Metadata = new(srsdk.Metadata)
		if err := read(metadata, request.Metadata); err != nil {
			return err
		}
	}

	ruleset, err := cmd.Flags().GetString("ruleset")
	if err != nil {
		return err
	}
	if ruleset != "" {
		request.RuleSet = new(srsdk.RuleSet)
		if err := read(ruleset, request.RuleSet); err != nil {
			return err
		}
	}

	normalize, err := cmd.Flags().GetBool("normalize")
	if err != nil {
		return err
	}
	opts := &srsdk.RegisterOpts{Normalize: optional.NewBool(normalize)}

	response, _, err := srClient.DefaultApi.Register(ctx, subject, request, opts)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&schemaCreateOut{Id: response.Id})
	return table.Print()
}

func read(path string, v any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	return json.NewDecoder(file).Decode(v)
}
