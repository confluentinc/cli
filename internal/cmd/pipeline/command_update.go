package pipeline

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *command) newUpdateCommand(enableSourceCode bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <pipeline-id>",
		Short: "Update an existing pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to update Stream Designer pipeline "pipe-12345", with new name and new description.`,
				Code: `confluent pipeline update pipe-12345 --name test-pipeline --description "Description of the pipeline"`,
			},
			examples.Example{
				Text: `Grant privilege to activate Stream Designer pipeline "pipe-12345".`,
				Code: `confluent pipeline update pipe-12345 --activation-privilege true`,
			},
			examples.Example{
				Text: `Revoke privilege to activate Stream Designer pipeline "pipe-12345".`,
				Code: `confluent pipeline update pipe-12345 --activation-privilege false`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the pipeline.")
	cmd.Flags().String("description", "", "Description of the pipeline.")
	if enableSourceCode {
		cmd.Flags().String("sql-file", "", "Path to a KSQL file containing the pipeline's source code.")
		cmd.Flags().StringArray("secret", []string{}, "A named secret that can be referenced in pipeline source code, e.g. \"secret_name=secret_content\".\n"+
			"This flag can be supplied multiple times. The secret mapping must have the format <secret-name>=<secret-value>,\n"+
			"where <secret-name> consists of 1-128 lowercase, uppercase, numeric or underscore characters but may not begin with a digit.\n"+
			"If <secret-value> is empty, the named secret will be removed from Stream Designer.")
	}
	cmd.Flags().Bool("activation-privilege", true, "Grant or revoke the privilege to activate this pipeline.")
	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	if enableSourceCode {
		cobra.CheckErr(cmd.MarkFlagFilename("sql-file", "sql"))
	}

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	sqlFile, _ := cmd.Flags().GetString("sql-file")
	secrets, _ := cmd.Flags().GetStringArray("secret")

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if name == "" && description == "" && sqlFile == "" && len(secrets) == 0 && !cmd.Flags().Changed("activation-privilege") {
		return fmt.Errorf("one of the update options must be provided: --name, --description, --sql-file, --secret, --activation-privilege")
	}

	updatePipeline := streamdesignerv1.SdV1PipelineUpdate{Spec: &streamdesignerv1.SdV1PipelineSpecUpdate{}}
	if name != "" {
		updatePipeline.Spec.SetDisplayName(name)
	}
	if description != "" {
		updatePipeline.Spec.SetDescription(description)
	}
	if sqlFile != "" {
		// read pipeline source code file if provided
		fileContent, err := os.ReadFile(sqlFile)
		if err != nil {
			return err
		}
		sourceCode := string(fileContent)
		updatePipeline.Spec.SetSourceCode(streamdesignerv1.SdV1SourceCodeObject{Sql: sourceCode})
	}
	// parse and construct secret mappings
	secretMappings, err := createSecretMappings(secrets, secretMappingWithEmptyValue)
	if err != nil {
		return err
	}
	updatePipeline.Spec.SetSecrets(secretMappings)

	if cmd.Flags().Changed("activation-privilege") {
		activationPrivilege, _ := cmd.Flags().GetBool("activation-privilege")
		updatePipeline.Spec.SetActivationPrivilege(activationPrivilege)
	}

	pipeline, err := c.V2Client.UpdateSdPipeline(c.EnvironmentId(), cluster.ID, args[0], updatePipeline)
	if err != nil {
		return err
	}

	return printTable(cmd, pipeline)
}
