package schemaregistry

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type outputStruct struct {
	Id int32 `json:"id" yaml:"id"`
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

	metadataPath, err := cmd.Flags().GetString("metadata")
	if err != nil {
		return err
	}
	metadata, err := readMetadata(metadataPath)
	if err != nil {
		return err
	}

	rulesetPath, err := cmd.Flags().GetString("ruleset")
	if err != nil {
		return err
	}
	ruleset, err := readRuleset(rulesetPath)
	if err != nil {
		return err
	}

	normalize, err := cmd.Flags().GetBool("normalize")
	if err != nil {
		return err
	}

	request := srsdk.RegisterSchemaRequest{
		Schema:     string(schema),
		SchemaType: schemaType,
		References: refs,
		Metadata:   *metadata,
		RuleSet:    *ruleset,
	}
	response, _, err := srClient.DefaultApi.Register(ctx, subject, request,
		&srsdk.RegisterOpts{Normalize: optional.NewBool(normalize)})
	if err != nil {
		return err
	}

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, &outputStruct{response.Id})
	}

	output.Printf(errors.RegisteredSchemaMsg, response.Id)
	return nil
}

func readMetadata(metadataPath string) (*srsdk.Metadata, error) {
	var metadata srsdk.Metadata
	if metadataPath != "" {
		metadataBlob, err := os.ReadFile(metadataPath)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metadataBlob, &metadata); err != nil {
			return nil, err
		}
	}
	return &metadata, nil
}

func readRuleset(rulesetPath string) (*srsdk.RuleSet, error) {
	var ruleSet srsdk.RuleSet
	if rulesetPath != "" {
		rulesetBlob, err := os.ReadFile(rulesetPath)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rulesetBlob, &ruleSet); err != nil {
			return nil, err
		}
	}
	return &ruleSet, nil
}
