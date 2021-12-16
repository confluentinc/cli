package schemaregistry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func (c *schemaCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a schema.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Register a new schema.",
				Code: fmt.Sprintf("%s schema-registry schema create --subject payments --schema schemafilepath", pversion.CLIName),
			},
			examples.Example{
				Text: "Where schemafilepath may include these contents.",
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
	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().String("type", "", `Specify the schema type as "AVRO", "PROTOBUF", or "JSON".`)
	cmd.Flags().String("refs", "", "The path to the references file.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("schema")
	_ = cmd.MarkFlagRequired("subject")

	pcmd.RegisterFlagCompletionFunc(cmd, "type", func(_ *cobra.Command, _ []string) []string { return []string{"AVRO", "PROTOBUF", "JSON"} })

	return cmd
}

func (c *schemaCommand) create(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
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

	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	var refs []srsdk.SchemaReference
	refPath, err := cmd.Flags().GetString("refs")
	if err != nil {
		return err
	}
	if refPath != "" {
		refBlob, err := ioutil.ReadFile(refPath)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(refBlob, &refs); err != nil {
			return err
		}
	}

	response, _, err := srClient.DefaultApi.Register(ctx, subject, srsdk.RegisterSchemaRequest{Schema: string(schema), SchemaType: schemaType, References: refs})
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if outputFormat == output.Human.String() {
		utils.Printf(cmd, errors.RegisteredSchemaMsg, response.Id)
	} else {
		return output.StructuredOutput(outputFormat, &struct {
			Id int32 `json:"id" yaml:"id"`
		}{response.Id})
	}

	return nil
}
