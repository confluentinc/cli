package schema_registry

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
)

type schemaCommand struct {
	*cobra.Command
	config   *config.Config
	ch       *pcmd.ConfigHelper
	srClient *srsdk.APIClient
}

func NewSchemaCommand(config *config.Config, ch *pcmd.ConfigHelper, srClient *srsdk.APIClient) *cobra.Command {
	schemaCmd := &schemaCommand{
		Command: &cobra.Command{
			Use:   "schema",
			Short: "Manage Schema Registry schemas.",
		},
		config:   config,
		ch:       ch,
		srClient: srClient,
	}
	schemaCmd.init()
	return schemaCmd.Command
}

func (c *schemaCommand) init() {
	cmd := &cobra.Command{
		Use:   "create --subject <subject> --schema <schema-file>",
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

`, c.config.CLIName),
		RunE: c.create,
		Args: cobra.NoArgs,
	}
	RequireSubjectFlag(cmd)
	cmd.Flags().String("schema", "", "The path to the schema file.")
	_ = cmd.MarkFlagRequired("schema")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete --subject <subject> --version <version>",
		Short: "Delete one or more schemas.",
		Example: FormatDescription(`
Delete one or more topics. This command should only be used in extreme circumstances.

::

		{{.CLIName}} schema-registry schema delete --subject payments --version latest`, c.config.CLIName),
		RunE: c.delete,
		Args: cobra.NoArgs,
	}
	RequireSubjectFlag(cmd)
	cmd.Flags().StringP("version", "V", "", "Version of the schema. Can be a specific version, 'all', or 'latest'.")
	_ = cmd.MarkFlagRequired("version")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "describe <schema-id> [--subject <subject>] [--version <version]",
		Short: "Get schema either by schema-id, or by subject/version.",
		Example: FormatDescription(`
Describe the schema string by schema ID

::

		{{.CLIName}} schema-registry schema describe 1337

Describe the schema by subject and version

::

		{{.CLIName}} schema-registry schema describe --subject payments --version latest
`, c.config.CLIName),
		RunE: c.describe,
		Args: cobra.MaximumNArgs(1),
	}
	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().StringP("version", "V", "", "Version of the schema. Can be a specific version or 'latest'.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)
}

func (c *schemaCommand) create(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
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

	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	response, _, err := srClient.DefaultApi.Register(ctx, subject, srsdk.RegisterSchemaRequest{Schema: string(schema)})
	if err != nil {
		return err
	}
	pcmd.Printf(cmd, "Successfully registered schema with ID: %v \n", response.Id)
	return nil
}

func (c *schemaCommand) delete(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
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
	if version == "all" {
		versions, _, err := srClient.DefaultApi.DeleteSubject(ctx, subject)
		if err != nil {
			return err
		}
		pcmd.Println(cmd, "Successfully deleted all versions for subject")
		PrintVersions(versions)
		return nil
	} else {
		versionResult, _, err := srClient.DefaultApi.DeleteSchemaVersion(ctx, subject, version)
		if err != nil {
			return err
		}
		pcmd.Println(cmd, "Successfully deleted version for subject")
		PrintVersions([]int32{versionResult})
		return nil
	}
}

func (c *schemaCommand) describe(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return c.describeById(cmd, args)
	} else {
		return c.describeBySubject(cmd, args)
	}
}

func (c *schemaCommand) describeById(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
	if err != nil {
		return err
	}
	schema, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("unexpected argument: Must be an integer Schema ID")
	}
	schemaString, _, err := srClient.DefaultApi.GetSchema(ctx, int32(schema))
	if err != nil {
		return err
	}
	pcmd.Println(cmd, schemaString.Schema)
	return nil
}

func (c *schemaCommand) describeBySubject(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
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
	schemaString, _, err := srClient.DefaultApi.GetSchemaByVersion(ctx, subject, version)
	if err != nil {
		return err
	}
	pcmd.Println(cmd, schemaString.Schema)
	return nil
}
