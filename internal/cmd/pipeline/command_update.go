package pipeline

import (
	"context"
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
			examples.Example{
				Text: `Update Stream Designer pipeline "pipe-12345" with KSQL cluster ID "lksqlc-123456".`,
				Code: "confluent pipeline update pipe-12345 --ksql-cluster lksqlc-123456",
			},
			examples.Example{
				Text: `Update Stream Designer pipeline "pipe-12345" with new Schema Registry cluster ID.`,
				Code: "confluent pipeline update pipe-12345 --update-schema-registry",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the pipeline.")
	cmd.Flags().String("description", "", "Description of the pipeline.")
	pcmd.AddKsqlClusterFlag(cmd, c.AuthenticatedCLICommand)
	if enableSourceCode {
		cmd.Flags().String("sql-file", "", "Path to a KSQL file containing the pipeline's source code.")
		cmd.Flags().StringArray("secret", []string{}, "A named secret that can be referenced in pipeline source code, e.g. \"secret_name=secret_content\".\n"+
			"This flag can be supplied multiple times. The secret mapping must have the format <secret-name>=<secret-value>,\n"+
			"where <secret-name> consists of 1-128 lowercase, uppercase, numeric or underscore characters but may not begin with a digit.\n"+
			"If <secret-value> is empty, the named secret will be removed from Stream Designer.")
	}
	cmd.Flags().Bool("activation-privilege", true, "Grant or revoke the privilege to activate this pipeline.")
	cmd.Flags().Bool("update-schema-registry", false, "Update the pipeline with the latest Schema Registry cluster.")
	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	if enableSourceCode {
		cobra.CheckErr(cmd.MarkFlagFilename("sql-file", "sql"))
	}

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	ksqlCluster, err := cmd.Flags().GetString("ksql-cluster")
	if err != nil {
		return err
	}

	sqlFile, err := cmd.Flags().GetString("sql-file")
	if err != nil {
		return err
	}

	secrets, err := cmd.Flags().GetStringArray("secret")
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if name == "" && description == "" && sqlFile == "" && len(secrets) == 0 && ksqlCluster == "" &&
		!cmd.Flags().Changed("activation-privilege") && !cmd.Flags().Changed("update-schema-registry") {
		return fmt.Errorf("one of the update options must be provided:" +
			" `--name`," +
			" `--description`," +
			" `--ksql-cluster`," +
			" `--sql-file`," +
			" `--secret`," +
			" `--activation-privilege`," +
			" `--update-schema-registry`")
	}

	updatePipeline := streamdesignerv1.SdV1Pipeline{Spec: &streamdesignerv1.SdV1PipelineSpec{}}

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

	environmentId, err := c.EnvironmentId()
	if err != nil {
		return err
	}

	if ksqlCluster != "" {
		if _, err := c.V2Client.DescribeKsqlCluster(ksqlCluster, environmentId); err != nil {
			return err
		}
		updatePipeline.Spec.SetKsqlCluster(streamdesignerv1.ObjectReference{Id: ksqlCluster})
	}

	if cmd.Flags().Changed("update-schema-registry") {
		updateSchemaRegistry, err := cmd.Flags().GetBool("update-schema-registry")
		if err != nil {
			return err
		}
		if updateSchemaRegistry {
			srCluster, err := c.Context.FetchSchemaRegistryByEnvironmentId(context.Background(), environmentId)
			if err != nil {
				return err
			}
			updatePipeline.Spec.SetStreamGovernanceCluster(streamdesignerv1.ObjectReference{Id: srCluster.GetId()})
		}
	}

	pipeline, err := c.V2Client.UpdateSdPipeline(environmentId, cluster.ID, args[0], updatePipeline)
	if err != nil {
		return err
	}

	return printTable(cmd, pipeline)
}
