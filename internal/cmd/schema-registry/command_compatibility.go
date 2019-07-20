package schema_registry

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
	"io/ioutil"
)

type compatibilityCommand struct {
	*cobra.Command
	config   *config.Config
	ch       *pcmd.ConfigHelper
	srClient *srsdk.APIClient
}

func NewCompatibilityCommand(config *config.Config, ch *pcmd.ConfigHelper, srClient *srsdk.APIClient) *cobra.Command {
	compatCmd := &compatibilityCommand{
		Command: &cobra.Command{
			Use:   "compatibility",
			Short: "Manage Schema Registry compatibility.",
		},
		config:   config,
		ch:       ch,
		srClient: srClient,
	}
	compatCmd.init()
	return compatCmd.Command
}

func (c *compatibilityCommand) init() {
	//Describe
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe compatability level [--subject <subject>].",
		Example: `
Global configuration

::
		ccloud schema-registry compatibility describe

For a specific subject

::
		ccloud schema-registry compatibility describe --subject payments`,
		RunE: c.describe,
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	// Update
	cmd = &cobra.Command{
		Use:   "update <compatability> [--subject <subject>]",
		Short: "Update the compatability level.",
		Example: `
<compatibility> can be BACKWARD, BACKWARD_TRANSITIVE, FORWARD, FORWARD_TRANSITIVE, FULL, FULL_TRANSITIVE, and NONE.

Update the global config level

::
		ccloud schema-registry compatibility update FULL

Update for a specific subject

::
		ccloud schema-registry compatibility update FULL --subject payments`,
		RunE: c.update,
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	// Check
	cmd = &cobra.Command{
		Use:   "check --subject <subject> --version <version> --schema @path/to/schema",
		Short: "Check an input schema against a version of the subject.",
		Example: `
Check an input schema against a particular version of the subject schema for compatibility in the current environment.

::
		ccloud schema-registry compatibility check --schema myschema.avro --subject payments --version latest`,
		RunE: c.testCompat,
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	_ = cmd.MarkFlagRequired("subject")
	cmd.Flags().StringP("version", "V", "", "Version of the schema. Can be a specific version or 'latest'.")
	_ = cmd.MarkFlagRequired("version")
	cmd.Flags().String("schema", "", "Path to the schema file.")
	_ = cmd.MarkFlagRequired("schema")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)
}

func (c *compatibilityCommand) describe(cmd *cobra.Command, args []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	if subject == "" {
		return c.describeTopLevel(cmd, args)
	} else {
		return c.describeSubject(cmd, args)
	}
}

func (c *compatibilityCommand) update(cmd *cobra.Command, args []string) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	if subject == "" {
		return c.updateTopLevel(cmd, args)
	} else {
		return c.updateSubject(cmd, args)
	}
}

func (c *compatibilityCommand) describeTopLevel(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
	if err != nil {
		return err
	}

	configResult, _, err := srClient.DefaultApi.GetTopLevelConfig(ctx)

	if err != nil {
		return err
	}

	fmt.Println("Compatability level: " + configResult.CompatibilityLevel)
	return nil
}

func (c *compatibilityCommand) updateTopLevel(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
	if err != nil {
		return err
	}

	updateReq := srsdk.ConfigUpdateRequest{Compatibility: args[0]}

	_, _, err = srClient.DefaultApi.UpdateTopLevelConfig(ctx, updateReq)
	if err != nil {
		return err
	}
	fmt.Println("Successfully updated")
	return nil
}

func (c *compatibilityCommand) describeSubject(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
	if err != nil {
		return err
	}
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}
	configResult, _, err := srClient.DefaultApi.GetSubjectLevelConfig(ctx, subject)
	if err != nil {
		return err
	}
	fmt.Println("Compatability level: " + configResult.CompatibilityLevel)
	return nil
}

func (c *compatibilityCommand) updateSubject(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
	if err != nil {
		return err
	}
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	updateReq := srsdk.ConfigUpdateRequest{Compatibility: args[0]}

	_, _, err = srClient.DefaultApi.UpdateSubjectLevelConfig(ctx, subject, updateReq)
	if err != nil {
		return err
	}

	fmt.Println("Successfully updated")
	return nil
}

func (c *compatibilityCommand) testCompat(cmd *cobra.Command, args []string) error {
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

	schemaPath, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}

	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	result, _, err := srClient.DefaultApi.TestCompatabilityBySubjectName(ctx, subject, version, srsdk.RegisterSchemaRequest{Schema: string(schema)}, nil)
	if err != nil {
		return err
	}

	if result.IsCompatible {
		fmt.Println("Schemas are compatible")
	} else {
		fmt.Println("Schemas are not compatible")
	}
	return nil
}
