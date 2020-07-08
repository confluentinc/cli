package schema_registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type schemaCommand struct {
	*pcmd.AuthenticatedCLICommand
	srClient *srsdk.APIClient
}

func NewSchemaCommand(cliName string, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "schema",
			Short: "Manage Schema Registry schemas.",
		}, prerunner)
	schemaCmd := &schemaCommand{
		AuthenticatedCLICommand: cliCmd,
		srClient:                srClient,
	}
	schemaCmd.init(cliName)
	return schemaCmd.Command
}

func (c *schemaCommand) init(cliName string) {
	cmd := &cobra.Command{
		Use:   "create --subject <subject> --schema <schema-file> --type <schema-type> --refs <ref-file>",
		Short: "Create a schema.",
		Example: FormatDescription(`
Register a new schema

::

		{{.CLIName}} schema-registry schema create --subject payments --schema schemafilepath

Where schemafilepath may include these contents:

::

	{
	   "type" : "record",
	   "namespace" : "Example",
	   "name" : "Employee",
	   "fields" : [
		  { "name" : "Name" , "type" : "string" },
		  { "name" : "Age" , "type" : "int" }
	   ]
	}

- For more information on schema types, see
  https://docs.confluent.io/current/schema-registry/serdes-develop/index.html.
- For more information on schema references, see
  https://docs.confluent.io/current/schema-registry/serdes-develop/index.html#schema-references.
`, cliName),
		RunE: c.create,
		Args: cobra.NoArgs,
	}
	RequireSubjectFlag(cmd)
	cmd.Flags().String("schema", "", "The path to the schema file.")
	_ = cmd.MarkFlagRequired("schema")
	cmd.Flags().String("type", "", `Specify the schema type as "AVRO", "PROTOBUF", or "JSON".`)
	cmd.Flags().String("refs", "", "The path to the references file.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete --subject <subject> --version <version> --permanent",
		Short: "Delete one or more schemas.",
		Example: FormatDescription(`
Delete one or more topics. This command should only be used in extreme circumstances.

::

		{{.CLIName}} schema-registry schema delete --subject payments --version latest`, cliName),
		RunE: c.delete,
		Args: cobra.NoArgs,
	}
	RequireSubjectFlag(cmd)
	cmd.Flags().StringP("version", "V", "", "Version of the schema. Can be a specific version, 'all', or 'latest'.")
	_ = cmd.MarkFlagRequired("version")
	cmd.Flags().BoolP("permanent", "P", false, "Permanently delete the schema.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "describe <schema-id> [--subject <subject>] [--version <version>]",
		Short: "Get schema either by schema-id, or by subject/version.",
		Example: FormatDescription(`
Describe the schema string by schema ID

::

		{{.CLIName}} schema-registry schema describe 1337

Describe the schema by both subject and version

::

		{{.CLIName}} schema-registry describe --subject payments --version latest
`, cliName),
		PreRunE: c.preDescribe,
		RunE:    c.describe,
		Args:    cobra.MaximumNArgs(1),
	}
	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().StringP("version", "V", "", "Version of the schema. Can be a specific version or 'latest'.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)
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
	} else if refPath != "" {
		refBlob, err := ioutil.ReadFile(refPath)
		if err != nil {
			return err
		}
		err = json.Unmarshal(refBlob, &refs)
		if err != nil {
			return err
		}
	}

	response, _, err := srClient.DefaultApi.Register(ctx, subject, srsdk.RegisterSchemaRequest{Schema: string(schema), SchemaType: schemaType, References: refs})
	if err != nil {
		return err
	}
	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if outputFormat == output.Human.String() {
		pcmd.Printf(cmd, "Successfully registered schema with ID: %v \n", response.Id)
	} else {
		return output.StructuredOutput(outputFormat, &struct {
			Id int32 `json:"id" yaml:"id"`
		}{response.Id})
	}
	return nil
}

func (c *schemaCommand) delete(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}
	permanent, err := cmd.Flags().GetBool("permanent")
	if err != nil {
		return err
	}
	if version == "all" {
		deleteSubjectOpts := srsdk.DeleteSubjectOpts{Permanent: optional.NewBool(permanent)}
		versions, _, err := srClient.DefaultApi.DeleteSubject(ctx, subject, &deleteSubjectOpts)
		if err != nil {
			return err
		}
		pcmd.Println(cmd, "Successfully deleted all versions for subject")
		PrintVersions(versions)
		return nil
	} else {
		deleteVersionOpts := srsdk.DeleteSchemaVersionOpts{Permanent: optional.NewBool(permanent)}
		versionResult, _, err := srClient.DefaultApi.DeleteSchemaVersion(ctx, subject, version, &deleteVersionOpts)
		if err != nil {
			return err
		}
		pcmd.Println(cmd, "Successfully deleted version for subject")
		PrintVersions([]int32{versionResult})
		return nil
	}
}

func (c *schemaCommand) preDescribe(cmd *cobra.Command, args []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	if len(args) > 0 && (subject != "" || version != "") {
		return fmt.Errorf("Cannot specify both schema ID and subject/version")
	} else if len(args) == 0 && (subject == "" || version == "") {
		return fmt.Errorf("Must specify either schema ID or subject and version")
	}
	return nil
}

func (c *schemaCommand) describe(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return c.describeById(cmd, args)
	} else {
		return c.describeBySubject(cmd)
	}
}

func (c *schemaCommand) describeById(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	schema, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("unexpected argument: Must be an integer Schema ID")
	}
	schemaString, _, err := srClient.DefaultApi.GetSchema(ctx, int32(schema), nil)
	if err != nil {
		return err
	}
	return c.printSchema(cmd, schemaString.Schema, schemaString.SchemaType, schemaString.References)
}

func (c *schemaCommand) describeBySubject(cmd *cobra.Command) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}
	schemaString, _, err := srClient.DefaultApi.GetSchemaByVersion(ctx, subject, version, nil)
	if err != nil {
		return err
	}
	return c.printSchema(cmd, schemaString.Schema, schemaString.SchemaType, schemaString.References)
}

func (c *schemaCommand) printSchema(cmd *cobra.Command, schema string, sType string, refs []srsdk.SchemaReference) error {
	if sType != "" {
		pcmd.Println(cmd, "Type: "+sType)
	}
	pcmd.Println(cmd, "Schema: "+schema)
	if len(refs) > 0 {
		pcmd.Println(cmd, "References:")
		for i := 0; i < len(refs); i++ {
			pcmd.Printf(cmd, "\t%s -> %s %d\n", refs[i].Name, refs[i].Subject, refs[i].Version)
		}
	}
	return nil
}
