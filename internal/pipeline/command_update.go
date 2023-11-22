package pipeline

import (
	"os"

	"github.com/spf13/cobra"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/v3/pkg/dynamic-config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <pipeline-id>",
		Short:             "Update an existing pipeline.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to update Stream Designer pipeline "pipe-12345", with new name and new description.`,
				Code: `confluent pipeline update pipe-12345 --name test-pipeline --description "Description of the pipeline"`,
			},
			examples.Example{
				Text: `Grant privilege to activate Stream Designer pipeline "pipe-12345".`,
				Code: `confluent pipeline update pipe-12345 --activation-privilege=true`,
			},
			examples.Example{
				Text: `Revoke privilege to activate Stream Designer pipeline "pipe-12345".`,
				Code: `confluent pipeline update pipe-12345 --activation-privilege=false`,
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
	cmd.Flags().String("sql-file", "", "Path to a KSQL file containing the pipeline's source code.")
	cmd.Flags().StringArray("secret", []string{}, "A named secret that can be referenced in pipeline source code, for example, \"secret_name=secret_content\".\n"+
		"This flag can be supplied multiple times. The secret mapping must have the format <secret-name>=<secret-value>,\n"+
		"where <secret-name> consists of 1-128 lowercase, uppercase, numeric or underscore characters but may not begin with a digit.\n"+
		"If <secret-value> is empty, the named secret will be removed from Stream Designer.")
	cmd.Flags().Bool("activation-privilege", true, "Grant or revoke the privilege to activate this pipeline.")
	cmd.Flags().Bool("update-schema-registry", false, "Update the pipeline with the latest Schema Registry cluster.")
	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	cobra.CheckErr(cmd.MarkFlagFilename("sql-file", "sql"))

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	flags := []string{
		"activation-privilege",
		"description",
		"ksql-cluster",
		"name",
		"secret",
		"sql-file",
		"update-schema-registry",
	}
	if err := errors.CheckNoUpdate(cmd.Flags(), flags...); err != nil {
		return err
	}

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

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := dynamicconfig.GetKafkaClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	pipeline := streamdesignerv1.SdV1Pipeline{Spec: &streamdesignerv1.SdV1PipelineSpec{
		Environment:  &streamdesignerv1.ObjectReference{Id: environmentId},
		KafkaCluster: &streamdesignerv1.ObjectReference{Id: cluster.ID},
	}}

	if name != "" {
		pipeline.Spec.SetDisplayName(name)
	}
	if description != "" {
		pipeline.Spec.SetDescription(description)
	}
	if sqlFile != "" {
		// read pipeline source code file if provided
		fileContent, err := os.ReadFile(sqlFile)
		if err != nil {
			return err
		}
		sourceCode := string(fileContent)
		pipeline.Spec.SetSourceCode(streamdesignerv1.SdV1SourceCodeObject{Sql: sourceCode})
	}
	// parse and construct secret mappings
	secretMappings, err := createSecretMappings(secrets, secretMappingWithEmptyValue)
	if err != nil {
		return err
	}
	pipeline.Spec.SetSecrets(secretMappings)

	if cmd.Flags().Changed("activation-privilege") {
		activationPrivilege, _ := cmd.Flags().GetBool("activation-privilege")
		pipeline.Spec.SetActivationPrivilege(activationPrivilege)
	}

	if ksqlCluster != "" {
		if _, err := c.V2Client.DescribeKsqlCluster(ksqlCluster, environmentId); err != nil {
			return err
		}
		pipeline.Spec.SetKsqlCluster(streamdesignerv1.ObjectReference{Id: ksqlCluster})
	}

	if cmd.Flags().Changed("update-schema-registry") {
		updateSchemaRegistry, err := cmd.Flags().GetBool("update-schema-registry")
		if err != nil {
			return err
		}
		if updateSchemaRegistry {
			clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(environmentId)
			if err != nil {
				return err
			}
			if len(clusters) == 0 {
				return errors.NewSRNotEnabledError()
			}
			pipeline.Spec.SetStreamGovernanceCluster(streamdesignerv1.ObjectReference{Id: clusters[0].GetId()})
		}
	}

	pipeline, err = c.V2Client.UpdateSdPipeline(args[0], pipeline)
	if err != nil {
		return err
	}

	return printTable(cmd, pipeline)
}
