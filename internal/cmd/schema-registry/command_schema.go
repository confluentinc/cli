package schema_registry

import (
	"context"
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/go-printer"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
	"strconv"
)

type schemaCommand struct {
	*cobra.Command
	config *config.Config
	srSdk  *srsdk.APIClient
	ch     *pcmd.ConfigHelper
	ctx    context.Context
}

func NewSchemaCommand(config *config.Config, srSdk *srsdk.APIClient) *cobra.Command {
	ctx, err := pcmd.SrContext(config)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	schemaCmd := &schemaCommand{
		Command: &cobra.Command{
			Use:   "schema",
			Short: "Manage Schema Registry schemas",
		},
		config: config,
		srSdk:  srSdk,
		ctx:    ctx,
	}
	schemaCmd.init()
	return schemaCmd.Command
}

func (c *schemaCommand) init() {

	cmd := &cobra.Command{
		Use:   "delete --subject <subject> --version <version>",
		Short: "Delete one or more schemas",
		Example: `
Delete one or more topics. This command should only be used in extreme circumstances.

::

		ccloud schema-registry schema delete --subject payments --version latest`,
		RunE: c.delete,
		Args: cobra.NoArgs,
	}
	requireSubjectFlag(cmd)
	cmd.Flags().StringP("version", "v", "", "Version of the schema. Can be a specific version, 'all', or 'latest'")
	cmd.MarkFlagRequired("version")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "describe <schema-id>",
		Short: "Get the schema string identified by the input ID",
		Example: `
Get the schema string identified by the input ID

::

		ccloud schema-registry describe 1337`,
		RunE: c.describeById,
		Args: cobra.ExactArgs(1),
	}
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "describe --subject <subject> --version <version>",
		Short: "Describe a schema for a given subject and version",
		Example: `
Describe a schema for a given subject and version

::

		ccloud schema-registry describe --subject payments --version latest
`,
		RunE: c.describeBySubject,
		Args: cobra.NoArgs,
	}
	requireSubjectFlag(cmd)
	cmd.Flags().StringP("version", "v", "", "Version of the schema. Can be a specific version or 'latest'")
	cmd.MarkFlagRequired("version")
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "list --subject <subject>",
		Short: "List all versions of a subject",
		Example: `
Get a list of versions registered under the specified subject.

::

		ccloud schema-registry schema list --subject payments`,
		RunE: c.list,
		Args: cobra.NoArgs,
	}
	cmd.Flags().String("subject", "", "Subject of the schema")
	cmd.MarkFlagRequired("subject")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)
}

func (c *schemaCommand) list(cmd *cobra.Command, args []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	versions, _, err := c.srSdk.DefaultApi.ListVersions(c.ctx, subject)

	if err != nil {
		// TODO handle auth errors specifically and show user how to reset their API Keys
		return err
	}
	printVersions(versions)

	return nil
}

func (c *schemaCommand) delete(cmd *cobra.Command, args []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}
	if version == "all" {
		versions, _, err := c.srSdk.DefaultApi.DeleteSubject(c.ctx, subject)
		if err != nil {
			return err
		}
		pcmd.Println(cmd, "Successfully deleted all versions for subject")
		printVersions(versions)
		return nil
	}
	versionResult, _, err := c.srSdk.DefaultApi.DeleteSchemaVersion(c.ctx, subject, version)
	if err != nil {
		return err
	}
	pcmd.Println(cmd, "Successfully deleted version for subject")
	printVersions([]int32{versionResult})
	return nil
}

func (c *schemaCommand) describeById(cmd *cobra.Command, args []string) error {
	schema, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("unexpected argument: Must be an integer Schema ID")
	}
	schemaString, _, err := c.srSdk.DefaultApi.GetSchema(c.ctx, int32(schema))
	if err != nil {
		return err
	}
	pcmd.Println(cmd, schemaString.Schema)
	return nil
}

func (c *schemaCommand) describeBySubject(cmd *cobra.Command, args []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}
	schemaString, _, err := c.srSdk.DefaultApi.GetSchemaByVersion(c.ctx, subject, version)
	if err != nil {
		return err
	}
	pcmd.Println(cmd, schemaString.Schema)
	return nil
}

func printVersions(versions []int32) {
	titleRow := []string{"Version"}
	var entries [][]string
	for _, version := range versions {
		record := &struct{ Version int32 }{version}
		entries = append(entries, printer.ToRow(record, titleRow))
	}
	printer.RenderCollectionTable(entries, titleRow)
}

func requireSubjectFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("subject", "s", "", "Subject of the schema")
	cmd.MarkFlagRequired("subject")
}
