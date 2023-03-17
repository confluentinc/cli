package schemaregistry

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
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
				Code: fmt.Sprintf("%s schema-registry schema create --subject payments --schema payments.avro --type avro", pversion.CLIName),
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
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

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

	response, _, err := srClient.DefaultApi.Register(ctx, subject, srsdk.RegisterSchemaRequest{Schema: string(schema), SchemaType: schemaType, References: refs})
	if err != nil {
		return err
	}

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, &outputStruct{response.Id})
	}

	output.Printf(errors.RegisteredSchemaMsg, response.Id)
	return nil
}
